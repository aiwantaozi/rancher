package stats

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rancher/rancher/pkg/ref"

	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rancher/pkg/clustermanager"
	monitorutil "github.com/rancher/rancher/pkg/monitoring"
	clusterv3 "github.com/rancher/types/apis/cluster.cattle.io/v3"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"github.com/rancher/types/config/dialer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	defaultExprDuring = "2m"
	projectIDAnn      = "field.cattle.io/projectId"
)

func newGraphHandler(dialerFactory dialer.Factory, clustermanager *clustermanager.Manager) *graphHandler {
	return &graphHandler{
		dialerFactory:  dialerFactory,
		clustermanager: clustermanager,
	}
}

type graphHandler struct {
	dialerFactory  dialer.Factory
	clustermanager *clustermanager.Manager
}

func (h *graphHandler) querySeriesAction(actionName string, action *types.Action, apiContext *types.APIContext) error {
	clusterName := getID(apiContext.Request.URL.Path, clusterPrefix)

	var queryGraphInput QueryGraphInput
	actionInput, err := parse.ReadBody(apiContext.Request)
	if err != nil {
		return err
	}

	if err = convert.ToObj(actionInput, &queryGraphInput); err != nil {
		return err
	}

	userContext, err := h.clustermanager.UserContext(clusterName)
	if err != nil {
		return fmt.Errorf("get usercontext failed, %v", err)
	}

	prometheusQuery, err := NewPrometheusQuery(userContext, clusterName, h.clustermanager, h.dialerFactory)
	if err != nil {
		return err
	}

	inputParser := newQueryGraphInputParser(apiContext.Request.Context(), userContext, queryGraphInput, true, "")
	if err = inputParser.parse(); err != nil {
		return err
	}

	nodeMap, err := getNodeName2InternalIPMap(userContext, clusterName)
	if err != nil {
		return err
	}

	graphs, err := userContext.ClusterV3.MonitorGraphs(monitorutil.CattleNamespaceName).List(metav1.ListOptions{LabelSelector: labels.Set(queryGraphInput.GraphSelector).AsSelector().String()})
	if err != nil {
		return fmt.Errorf("list monitor graph with label %+v failed, %v", queryGraphInput.GraphSelector, err)
	}

	var queries []*PrometheusQuery
	graphMap := make(map[string]*clusterv3.MonitorGraph)
	for _, graph := range graphs.Items {
		g := graph
		graphMap[getRefferenceGraphName(graph.Namespace, graph.Name)] = &g
		monitorMetrics, err := graph2Metrics(userContext, clusterName, &graph, inputParser.Input.MetricParams, inputParser.Input.IsDetails)
		if err != nil {
			return err
		}

		queries = append(queries, metrics2PrometheusQuery(monitorMetrics, inputParser.Start, inputParser.End, inputParser.Step, false)...)
	}

	seriesSlice, err := prometheusQuery.Do(queries)
	if err != nil {
		return fmt.Errorf("query series failed, %v", err)
	}

	if seriesSlice == nil {
		apiContext.WriteResponse(http.StatusNoContent, nil)
		return nil
	}

	timeSpliceMap := make(map[string]TimeResponseSeriesSlice)
	for _, v := range seriesSlice {
		graphName, _ := parseID(v.ID)
		timeSpliceMap[graphName] = append(timeSpliceMap[graphName], v)

	}

	collection := QueryGraphOutputCollection{Type: "collection"}
	for k, v := range timeSpliceMap {
		queryGraph := QueryGraph{
			Graph:  graphMap[k].Spec,
			Series: convertInstance(&v, nodeMap, graphMap[k].Labels["component"]),
		}
		collection.Data = append(collection.Data, queryGraph)
	}

	res, err := json.Marshal(collection)
	if err != nil {
		return fmt.Errorf("marshal query series result failed, %v", err)
	}
	apiContext.Response.Write(res)
	return nil
}

type metricWrap struct {
	clusterv3.MonitorMetric
	ExecuteExpression   string
	ReferenceGraphName  string
	ExecuteLegendFormat string
}

func graph2Metrics(userContext *config.UserContext, clusterName string, graph *clusterv3.MonitorGraph, metricParams MetricParams, isDetails bool) ([]*metricWrap, error) {
	resourceType := graph.Labels["component"]
	newMetricParams, err := parseMetricParams(userContext, resourceType, clusterName, metricParams)
	if err != nil {
		return nil, err
	}

	var excuteMetrics []*metricWrap
	var set labels.Set
	if isDetails && graph.Spec.DetailsMetricsSelector != nil {
		set = labels.Set(graph.Spec.DetailsMetricsSelector)
	} else {
		set = labels.Set(graph.Spec.MetricsSelector)
	}
	metrics, err := userContext.ClusterV3.MonitorMetrics(monitorutil.CattleNamespaceName).List(metav1.ListOptions{LabelSelector: set.AsSelector().String()}) //todo: why cache not work
	if err != nil {
		return nil, fmt.Errorf("list metrics failed, %v", err)
	}

	for _, v := range metrics.Items {
		groupBy := getGroupBy(v.Spec.GroupBy, v.Spec.DetailsGroupBy, resourceType, isDetails)
		executeExpression := replaceParams(newMetricParams, v.Spec.Expression, groupBy)
		excuteMetrics = append(excuteMetrics, &metricWrap{
			MonitorMetric:       *v.DeepCopy(),
			ExecuteExpression:   executeExpression,
			ReferenceGraphName:  getRefferenceGraphName(graph.Namespace, graph.Name),
			ExecuteLegendFormat: getLengendFormat(v.Spec.LegendFormat, v.Spec.DetailsLegendFormat, resourceType, isDetails),
		})
	}
	return excuteMetrics, nil
}

// func parseExtraAddTags(oldExtraAddTags map[string]string, resourceType string, metricParams *MetricParams) map[string]string {
// 	newExtraAddTags := make(map[string]string)
// 	for k, v := range oldExtraAddTags {
// 		newExtraAddTags[k] = v
// 	}
// 	switch resourceType {
// 	case ResourceCluster, ResourceNode, ResourceEtcd, ResourceAPIServer, ResourceScheduler, ResourceControllerManager:
// 		if MetricParams["InstanceName"] != "" {
// 			newExtraAddTags["instance"] = MetricParams["InstanceName"]
// 		}
// 	case ResourcePod:
// 		if MetricParams["InstanceName"] != "" {
// 			newExtraAddTags["pod_name"] = MetricParams["InstanceName"]
// 		}
// 	case ResourceContainer:
// 		if MetricParams["InstanceName"] != "" {
// 			newExtraAddTags["container_name"] = MetricParams["InstanceName"]
// 		}
// 	}
// 	return newExtraAddTags
// }

func metrics2PrometheusQuery(metrics []*metricWrap, start, end time.Time, step time.Duration, isInstanceQuery bool) []*PrometheusQuery {
	var queries []*PrometheusQuery
	for _, v := range metrics {
		id := getPrometheusQueryID(v.ReferenceGraphName, v.Name)
		queries = append(queries, InitPromQuery(id, start, end, step, v.ExecuteExpression, v.ExecuteLegendFormat, v.Spec.ExtraAddedTags, isInstanceQuery))
	}
	return queries
}

func nodeName2InternalIP(userContext *config.UserContext, clusterName, nodeName string) (string, error) {
	_, name := ref.Parse(nodeName)
	node, err := userContext.Management.Management.Nodes(metav1.NamespaceAll).Controller().Lister().Get(clusterName, name)
	if err != nil {
		return "", fmt.Errorf("get node from mgnt faild, %v", err)
	}

	internalNodeIP := getInternalNodeIP(node)
	if internalNodeIP == "" {
		return "", fmt.Errorf("could not find endpoint ip address for node %s", nodeName)
	}

	return internalNodeIP, nil
}

func getNodeName2InternalIPMap(userContext *config.UserContext, clusterName string) (map[string]string, error) {
	nodeMap := make(map[string]string)
	nodes, err := userContext.Management.Management.Nodes(metav1.NamespaceAll).Controller().Lister().List(clusterName, labels.NewSelector())
	if err != nil {
		return nil, fmt.Errorf("list node from mgnt failed, %v", err)
	}

	for _, node := range nodes {
		internalNodeIP := getInternalNodeIP(node)
		if internalNodeIP != "" {
			nodeMap[internalNodeIP] = node.Status.NodeName
		}
	}
	return nodeMap, nil
}

func getInternalNodeIP(node *mgmtv3.Node) string {
	for _, ip := range node.Status.InternalNodeStatus.Addresses {
		if ip.Type == "InternalIP" && ip.Address != "" {
			return ip.Address
		}
	}
	return ""
}

func getPrometheusQueryID(graphName, metricName string) string {
	return fmt.Sprintf("%s_%s", graphName, metricName)
}

func getLengendFormat(lengendFormat, detailsLengendFormat, resourceType string, isDetails bool) string {
	if isDetails && detailsLengendFormat != "" {
		return detailsLengendFormat
	}

	if lengendFormat != "" {
		return lengendFormat
	}

	var defaultLengendFormat string
	if isDetails {
		if detailsLengendFormat != "" {
			return detailsLengendFormat
		}
		switch resourceType {
		case ResourceCluster, ResourceNode, ResourceEtcd, ResourceAPIServer, ResourceScheduler, ResourceControllerManager:
			defaultLengendFormat = "{{instance}}"
		case ResourceWorkload:
			defaultLengendFormat = "{{pod_name}}"
		case ResourcePod:
			defaultLengendFormat = "{{container_name}}"
		}
		return defaultLengendFormat
	}

	switch resourceType {
	case ResourceNode:
		defaultLengendFormat = "{{instance}}"
	case ResourceWorkload:
		// defaultLengendFormat = "{{pod_name}}"
	case ResourcePod:
		defaultLengendFormat = "{{pod_name}}"
	}

	return defaultLengendFormat
}

func getGroupBy(groupBy, detailsGroupBy, resourceType string, isDetails bool) string {
	if isDetails && detailsGroupBy != "" {
		return detailsGroupBy
	}

	var defaultGroupBy string
	if isDetails {
		if detailsGroupBy != "" {
			return detailsGroupBy
		}
		switch resourceType {
		case ResourceCluster, ResourceNode, ResourceEtcd, ResourceAPIServer, ResourceScheduler, ResourceControllerManager:
			defaultGroupBy = "instance"
		case ResourceWorkload:
			defaultGroupBy = "pod_name"
		case ResourcePod:
			defaultGroupBy = "container_name"
		}
		return defaultGroupBy
	}

	if groupBy != "" {
		return groupBy
	}

	switch resourceType {
	case ResourceNode:
		defaultGroupBy = "instance"
	case ResourcePod:
		defaultGroupBy = "pod_name"
	}

	return defaultGroupBy
}

func parseID(ref string) (id1 string, id2 string) {
	parts := strings.SplitN(ref, "_", 2)
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}

func convertInstance(seriesSlice *TimeResponseSeriesSlice, nodeMap map[string]string, resourceType string) TimeResponseSeriesSlice {
	if resourceType == ResourceWorkload || resourceType == ResourceContainer || resourceType == ResourcePod {
		return *seriesSlice
	}

	for _, v := range *seriesSlice {
		hostName := strings.Split(v.Tags["instance"], ":")[0]
		if v.Name != "" {
			v.Name = strings.Replace(v.Name, v.Tags["instance"], nodeMap[hostName], -1)
		}

		if v.Tags["instance"] != "" {
			v.Tags["instance"] = nodeMap[hostName]
		}
	}
	return *seriesSlice
}

func getRefferenceGraphName(namespace, name string) string {
	return fmt.Sprintf("%s:%s", namespace, name)
}
