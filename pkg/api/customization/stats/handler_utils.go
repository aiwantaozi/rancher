package stats

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/rancher/pkg/ref"

	"github.com/rancher/rancher/pkg/controllers/user/workload"
	"github.com/rancher/rancher/pkg/monitoring"
	"github.com/rancher/types/config"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	clusterPrefix = "/v3/cluster/"
	projectPrefix = "/v3/project/"
	clusterLevel  = "cluster"
	projectLevel  = "project"
)

var (
	defaultQueryDuring = "30s"
	defaultTo          = "now"
	defaultFrom        = "now-" + defaultQueryDuring
)

var prometheusLabels = labels.Set{
	"app":        "prometheus",
	"prometheus": "prometheus-operator",
	"release":    "prometheus-operator",
}

func getID(url, prefix string) string {
	path := strings.TrimPrefix(url, prefix)
	slashIndex := strings.Index(path, "/")
	return path[:slashIndex]
}

func getPrometheusEndpoint(userContext *config.UserContext) (string, error) {
	appName, appNamespace, _ := monitoring.ClusterPrometheusEndpoint()
	svc, err := userContext.Core.Services("").Controller().Lister().Get(appNamespace, appName)
	if err != nil {
		return "", fmt.Errorf("Failed to get service for prometheus-server %v", err)
	}

	port := svc.Spec.Ports[0].Port
	url := "http://" + svc.Name + "." + svc.Namespace + ".svc.cluster.local:" + strconv.Itoa(int(port))
	return url, nil
}

func parseWorkloadName(id string) (typeName, namespace, name string, err error) {
	arr := strings.Split(id, ":")
	if len(arr) < 3 {
		return "", "", "", fmt.Errorf("invalid workload name: %s", id)
	}
	return arr[0], arr[1], arr[2], nil
}

type queryGraphInputParser struct {
	Input              *QueryGraphInput
	UserContext        *config.UserContext
	WorkloadController workload.CommonController
	IsClusterLevel     bool
	ProjectName        string
	Start              time.Time
	End                time.Time
	Step               time.Duration
}

func newQueryGraphInputParser(ctx context.Context, userContext *config.UserContext, input QueryGraphInput, isClusterLevel bool, projectName string) *queryGraphInputParser {
	return &queryGraphInputParser{
		Input:              &input,
		UserContext:        userContext,
		IsClusterLevel:     isClusterLevel,
		ProjectName:        projectName,
		WorkloadController: workload.NewWorkloadController(ctx, userContext.UserOnlyContext(), nil),
	}
}

func (p *queryGraphInputParser) parse() (err error) {
	if p.Input.MetricParams == nil {
		p.Input.MetricParams = make(map[string]string)
	}

	if err = p.parseTimeParams(); err != nil {
		return
	}

	// if err = p.checkResourceType(); err != nil {
	// 	return
	// }
	// if err = parseExpressionmetricParams(); err != nil {
	// 	return
	// }
	if !p.IsClusterLevel {
		if err = p.parseNamespace(); err != nil {
			return
		}
	}
	return nil
}

func parseMetricParams(userContext *config.UserContext, resourceType, clusterName string, metricParams MetricParams) (map[string]string, error) {
	newMetricParams := make(map[string]string)
	for k, v := range metricParams {
		newMetricParams[k] = v
	}

	var ip string
	var err error
	switch resourceType {
	case ResourceNode:
		instance := newMetricParams["instance"]
		if instance == "" {
			return nil, fmt.Errorf("instance in metric params is empty")
		}
		ip, err = nodeName2InternalIP(userContext, clusterName, instance)
		if err != nil {
			return newMetricParams, err
		}

	case ResourceWorkload:
		workloadName := newMetricParams["workloadName"]
		rcType, ns, name, err := parseWorkloadName(workloadName)
		if err != nil {
			return newMetricParams, err
		}

		var podOwners []string
		if workloadName != "" {
			if rcType == workload.ReplicaSetType || rcType == workload.ReplicationControllerType || rcType == workload.DaemonSetType || rcType == workload.StatefulSetType || rcType == workload.JobType || rcType == workload.CronJobType {
				podOwners = []string{name}
			}

			if rcType == workload.DeploymentType {
				rcType = workload.ReplicaSetType
				rcs, err := userContext.Apps.ReplicaSets(ns).List(metav1.ListOptions{})
				if err != nil {
					return newMetricParams, fmt.Errorf("list replicasets failed, %v", err)
				}

				for _, rc := range rcs.Items {
					if rc.OwnerReferences != nil && strings.ToLower(rc.OwnerReferences[0].Kind) == strings.ToLower(rcType) && rc.OwnerReferences[0].Name == name {
						podOwners = append(podOwners, rc.Name)
					}
				}
			}

			var podNames []string
			pods, err := userContext.Core.Pods(ns).List(metav1.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("list pod failed, %v", err)
			}
			for _, pod := range pods.Items {
				podRefName := pod.OwnerReferences[0].Name
				podRefKind := pod.OwnerReferences[0].Kind
				if pod.OwnerReferences != nil && contains(podRefName, podOwners...) && strings.ToLower(podRefKind) == strings.ToLower(rcType) {
					podNames = append(podNames, pod.Name)
				}
			}
			newMetricParams["podName"] = strings.Join(podNames, "|")
		}
	case ResourcePod:
		podName := newMetricParams["podName"]
		if podName == "" {
			return nil, fmt.Errorf("pod name is empty")
		}
		ns, name := ref.Parse(podName)
		newMetricParams["namespace"] = ns
		newMetricParams["podName"] = name
	case ResourceContainer:
		podName := newMetricParams["podName"]
		if podName == "" {
			return nil, fmt.Errorf("pod name is empty")
		}
		ns, name := ref.Parse(podName)
		newMetricParams["namespace"] = ns
		newMetricParams["podName"] = name

		containerName := newMetricParams["containerName"]
		if containerName == "" {
			return nil, fmt.Errorf("container name is empty")
		}
	}
	newMetricParams["instance"] = ip + ".*"
	return newMetricParams, nil
}

func (p *queryGraphInputParser) parseNamespace() error {
	if p.Input.MetricParams["namespace"] != "" {
		if !p.isAuthorizeNamespace() {
			return fmt.Errorf("could not query unauthorize namespace")
		}
		return nil
	}

	nss, err := p.getAuthroizeNamespace()
	if err != nil {
		return err
	}
	p.Input.MetricParams["namespace"] = nss
	return nil
}

func (p *queryGraphInputParser) parseTimeParams() (err error) {
	p.Start, p.End, p.Step, err = parseTimeParams(p.Input.From, p.Input.To, p.Input.Interval)
	return err
}

func (p *queryGraphInputParser) isAuthorizeNamespace() bool {
	ns, err := p.UserContext.Core.Namespaces(metav1.NamespaceAll).Get(p.Input.MetricParams["namespace"], metav1.GetOptions{})
	if err != nil {
		logrus.Errorf("get namespace %s info failed, %v", p.Input.MetricParams["namespace"], err)
		return false
	}
	return ns.Annotations[projectIDAnn] == p.ProjectName
}

func (p *queryGraphInputParser) getAuthroizeNamespace() (string, error) {
	nss, err := p.UserContext.Core.Namespaces(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("list namespace failed, %v", err)
	}
	var authNs []string
	for _, v := range nss.Items {
		if v.Annotations[projectIDAnn] == p.ProjectName {
			authNs = append(authNs, v.Name)
		}
	}
	return strings.Join(authNs, "|"), nil
}

// func (p *queryGraphInputParser) checkResourceType() error {
// 	if p.ResourceType == "" {
// 		return fmt.Errorf("graph seletor must have component type selector")
// 	}

// 	if p.IsClusterLevel {
// 		if contains(p.ResourceType, ResourceCluster, ResourceEtcd, ResourceNode, ResourceAPIServer, ResourceScheduler, ResourceControllerManager) {
// 			return nil
// 		}
// 	}

// 	if contains(p.ResourceType, ResourceWorkload, ResourcePod) {
// 		return nil
// 	}
// 	return fmt.Errorf("invalid resource type %s", p.ResourceType)
// }

// func (p *queryGraphInputParser) parseExpressionmetricParams() error {
// 	var instance, namespace string
// 	var err error
// 	switch p.ResourceType {
// 	case ResourceCluster, ResourceNode, ResourceEtcd, ResourceAPIServer, ResourceScheduler, ResourceControllerManager:
// 	case ResourceWorkload:
// 		if p.Input.MetricParams["InstanceName"] != "" {
// 			_, namespace, instance, err = parseWorkloadName(p.Input.MetricParams["InstanceName"])
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	case ResourcePod:
// 		if p.Input.MetricParams["InstanceName"] != "" {
// 			arr2 := strings.Split(p.Input.MetricParams["InstanceName"], ":")
// 			if len(arr2) < 2 {
// 				return fmt.Errorf("invalid pod name: %s", p.Input.MetricParams["InstanceName"])
// 			}
// 			namespace = arr2[0]
// 			instance = arr2[1]
// 		}
// 	}

// 	var during string
// 	if p.Input.MetricParams["during"] != "" {
// 		during = p.Input.MetricParams["during"]
// 	} else {
// 		during = defaultExprDuring
// 	}

// 	p.Input.MetricParams = MetricParams{
// 		"InstanceName": instance + ".*",
// 		"Namespace":    namespace,
// 		"GroupBy":      p.Input.MetricParams["groupBy"],
// 		"During":       during,
// 	}
// 	return nil
// }

func contains(str string, arr ...string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func parseTimeParams(from, to, interval string) (start, end time.Time, step time.Duration, err error) {
	if from == "" {
		from = defaultFrom
	}

	if to == "" {
		to = defaultTo
	}

	timeRange := NewTimeRange(from, to)
	start, err = timeRange.ParseFrom()
	if err != nil {
		err = fmt.Errorf("parse param from value %s faild, %v", from, err)
		return
	}

	end, err = timeRange.ParseTo()
	if err != nil {
		err = fmt.Errorf("parse param to value %s faild, %v", to, err)
		return
	}

	i, err := getIntervalFrom(interval, defaultMinInterval)
	if err != nil {
		err = fmt.Errorf("parse param interval value %s faild, %v", i, err)
		return
	}
	intervalCalculator := newIntervalCalculator(&IntervalOptions{MinInterval: i})
	calInterval := intervalCalculator.Calculate(timeRange, i)
	step = time.Duration(int64(calInterval.Value))
	return
}
