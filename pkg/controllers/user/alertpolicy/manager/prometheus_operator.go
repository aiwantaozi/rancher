package manager

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/rancher/rancher/pkg/controllers/user/workload"
	monitorutil "github.com/rancher/rancher/pkg/monitoring"
	nodeutil "github.com/rancher/rancher/pkg/node"
	"github.com/rancher/rancher/pkg/stats"
	"github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	monitoringv1 "github.com/rancher/types/apis/monitoring.cattle.io/v1"
	"github.com/rancher/types/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	comparison = map[string]string{
		"equal":            "==",
		"not-equal":        "!=",
		"greater-than":     ">",
		"less-than":        "<",
		"greater-or-equal": ">=",
		"less-or-equal":    "<=",
	}
)

type PromOperatorCRDManager struct {
	clusterName        string
	NodeLister         v3.NodeLister
	PrometheusRules    monitoringv1.PrometheusRuleInterface
	podLister          v1.PodLister
	workloadController workload.CommonController
}

type ResourceQueryCondition interface {
	GetQueryCondition(nameSelector string, labelSelector map[string]string) (interface{}, error)
}

type NodeQueryCondition struct {
	clusterName string
	NodeLister  v3.NodeLister
}

type ClusterQueryCondition struct{}

type WorkloadQueryCondition struct {
	podLister v1.PodLister
}

func NewPrometheusCRDManager(cluster *config.UserContext) *PromOperatorCRDManager {
	return &PromOperatorCRDManager{
		clusterName:        cluster.ClusterName,
		NodeLister:         cluster.Management.Management.Nodes(cluster.ClusterName).Controller().Lister(),
		PrometheusRules:    cluster.Monitoring.PrometheusRules(monitorutil.SystemMonitoringNamespaceName),
		workloadController: workload.NewWorkloadController(cluster.UserOnlyContext(), nil),
	}
}

func (c *PromOperatorCRDManager) GetResourceQueryType(resourceType string) ResourceQueryCondition {
	switch resourceType {
	case v3.ResourceWorkload, v3.ResourcePod:
		return &WorkloadQueryCondition{
			podLister: c.podLister,
		}

	case v3.ResourceNode:
		return &NodeQueryCondition{
			clusterName: c.clusterName,
			NodeLister:  c.NodeLister,
		}
	default:
		return &ClusterQueryCondition{}
	}
}

func (n *ClusterQueryCondition) GetQueryCondition(nameSelector string, labelSelector map[string]string) (interface{}, error) {
	return stats.NewMetricQueryData(v3.ResourceCluster, stats.ResourceQuery{
		During:   stats.DefaultDuring,
		HostName: ".*",
	}), nil
}

func (n *WorkloadQueryCondition) GetQueryCondition(nameSelector string, labelSelector map[string]string) (interface{}, error) {
	var podName string
	if nameSelector != "" {
		podName = nameSelector + ".*"
	} else {
		var podNames []string
		podList, err := n.podLister.List("", labels.Set(labelSelector).AsSelector())
		if err != nil {
			return nil, err
		}
		for _, v := range podList {
			podNames = append(podNames, v.Name)
		}
		podName = fmt.Sprintf("(%s)", strings.Join(podNames, "|"))
	}

	queryData := stats.NewMetricQueryData(v3.ResourcePod, stats.ResourceQuery{
		Namespace: ".*",
		PodName:   podName,
		During:    stats.DefaultDuring,
		GroupBy:   "pod_name",
	})
	return queryData.Generate(), nil
}

func (n *NodeQueryCondition) GetQueryCondition(nameSelector string, labelSelector map[string]string) (interface{}, error) {
	var nodeName string
	if nameSelector != "" {
		nodeName = nameSelector + ".*"
	} else {
		var nodeNames []string
		nodeList, err := n.NodeLister.List(n.clusterName, labels.Set(labelSelector).AsSelector())
		if err != nil {
			return nil, err
		}
		for _, v := range nodeList {
			nodeNames = append(nodeNames, nodeutil.GetEndpointNodeIP(v)+".*")
		}
		nodeName = fmt.Sprintf("(%s)", strings.Join(nodeNames, "|"))
	}

	queryData := stats.NewMetricQueryData(v3.ResourceNode, stats.ResourceQuery{
		HostName: nodeName,
		During:   stats.DefaultDuring,
		GroupBy:  "instance",
	})
	return queryData.Generate(), nil
}

func GetDefaultPrometheusRule(name string) *monitoringv1.PrometheusRule {
	return &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: monitorutil.SystemMonitoringNamespaceName,
			Labels: map[string]string{
				"release": "system-monitoring",
			},
		},
	}
}

func AddRuleGroup(promRule *monitoringv1.PrometheusRule, ruleGroup monitoringv1.RuleGroup) {
	if promRule.Spec.Groups == nil {
		promRule.Spec.Groups = []monitoringv1.RuleGroup{ruleGroup}
		return
	}

	for _, v := range promRule.Spec.Groups {
		if v.Name == ruleGroup.Name {
			v = ruleGroup
			return
		}
	}
	promRule.Spec.Groups = append(promRule.Spec.Groups, ruleGroup)
}

func (c *PromOperatorCRDManager) AddRule(id, name string, metrics []v3.Metric) (*monitoringv1.RuleGroup, error) {
	var rules []monitoringv1.Rule
	for _, v := range metrics {
		resourceType, _ := getTypeAndMetric(v.MetricType)
		queryData, err := c.GetResourceQueryType(resourceType).GetQueryCondition(v.NameSelector, v.LabelSelector)
		if err != nil {
			return nil, err
		}

		r := Metric2Rule(id, name, c.clusterName, resourceType, v, queryData)
		rules = append(rules, r)
	}

	return &monitoringv1.RuleGroup{
		Name:  name,
		Rules: rules,
	}, nil
}

func Metric2Rule(id, name, clusterName, resourceType string, metric v3.Metric, queryData interface{}) monitoringv1.Rule {
	labels := map[string]string{
		"alert_type":                        "metric",
		"group_id":                          id,
		"group_name":                        name,
		"cluster_name":                      clusterName,
		"metric":                            metric.MetricType,
		"severity":                          metric.Severity,
		"conditionThreshold_comparison":     metric.ConditionThreshold.Comparison,
		"conditionThreshold_duration":       metric.ConditionThreshold.Duration,
		"conditionThreshold_thresholdValue": fmt.Sprintf("%v", metric.ConditionThreshold.ThresholdValue),
		"resource_type":                     resourceType,
		//todo identity for each rule
	}

	expr := buildExpr(metric, queryData)

	return monitoringv1.Rule{
		Alert:  metric.MetricType,
		Expr:   expr,
		For:    metric.ConditionThreshold.Duration,
		Labels: labels,
	}
}

func buildExpr(metric v3.Metric, queryData interface{}) string {
	resourceType, metricType := getTypeAndMetric(metric.MetricType)
	expr := stats.GetExp(resourceType, metricType, queryData)
	return fmt.Sprintf("%s %s %v", expr, comparison[metric.ConditionThreshold.Comparison], metric.ConditionThreshold.ThresholdValue)
}

func getTypeAndMetric(metricType string) (string, string) {
	arr := strings.Split(metricType, "_")
	resourceType := arr[0]
	return resourceType, strings.TrimPrefix(metricType, resourceType)
}
