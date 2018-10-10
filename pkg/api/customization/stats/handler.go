package stats

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/pkg/clustermanager"
	"github.com/rancher/rancher/pkg/monitoring"
	managementv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	clusterPrefix = "/v3/clusters/"
	projectPrefix = "/v3/projects/"
)

const (
	defaultFrom = "now-1m"
	defaultTo   = "now"
)

var upgrader = websocket.Upgrader{}

var prometheusLabels = labels.Set{
	"app":        "prometheus",
	"prometheus": "prometheus-operator",
	"release":    "prometheus-operator",
}

type Handler struct {
	clustermanager *clustermanager.Manager
}

type params struct {
	id            string
	clusterName   string
	containerName string
	workloadName  string
	nodeName      string
	podName       string
	namespace     string
	start         time.Time
	end           time.Time
	step          time.Duration
	resourceType  string
}

func (s *Handler) Handler(apiContext *types.APIContext, _ types.RequestHandler) error {
	err := s.handler(apiContext)
	if err != nil {
		logrus.Errorf("Error during handle stats request %v", err)
	}
	return err
}

func (s *Handler) handler(apiContext *types.APIContext) error {
	params, err := parseParams(apiContext.Query, apiContext.Request.URL.Path)
	if err != nil {
		return fmt.Errorf("parse params failed, %v", err)
	}

	userContext, err := s.clustermanager.UserContext(params.clusterName)
	if err != nil {
		return fmt.Errorf("get usercontext failed, %v", err)
	}

	dial, err := userContext.Management.Dialer.ClusterDialer(params.clusterName)
	if err != nil {
		return fmt.Errorf("get dail from usercontext failed, %v", err)
	}

	endpoint, err := getPrometheusEndpoint(userContext)
	if err != nil {
		return err
	}

	api, err := NewPrometheusAPI(dial, endpoint)
	if err != nil {
		return err
	}

	series, err := NewQuery(api).Do(params)
	if err != nil {
		return err
	}
	metric := &CommontMetric{
		ID:     params.id,
		Series: series,
	}

	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	apiContext.Response.WriteHeader(200)
	apiContext.Response.Write(data)
	return nil
}

func getPrometheusEndpoint(userContext *config.UserContext) (string, error) {
	svc, err := userContext.Core.Services("").Controller().Lister().Get(monitoring.SystemMonitoringNamespaceName, monitoring.SystemMonitoringPrometheus)
	if err != nil {
		return "", fmt.Errorf("Failed to get service for prometheus-server %v", err)
	}

	ip := svc.Spec.ClusterIP
	port := svc.Spec.Ports[0].Port
	url := "http://" + ip + ":" + strconv.Itoa(int(port))

	return url, nil
}

func parseURL(path string) (clusterIDOrProjectName, resourceType, id string, err error) {
	path = strings.TrimPrefix(path, "/v3/clusters/")
	path = strings.TrimPrefix(path, "/v3/projects/")
	arr := strings.Split(path, "/")
	if len(arr) < 3 {
		err = fmt.Errorf("invalid request path %s", path)
		return
	}

	clusterIDOrProjectName = arr[0]
	resourceType = arr[1]
	id = arr[2]
	return
}

func parseParams(query url.Values, reqURL string) (*params, error) {
	from := query.Get("from")
	if from == "" {
		from = defaultFrom
	}

	to := query.Get("to")
	if to == "" {
		to = defaultTo
	}

	timeRange := NewTimeRange(from, to)
	start, err := timeRange.ParseFrom()
	if err != nil {
		return nil, fmt.Errorf("parse param from value %s faild, %v", from, err)
	}

	end, err := timeRange.ParseTo()
	if err != nil {
		return nil, fmt.Errorf("parse param to value %s faild, %v", to, err)
	}

	i, err := getIntervalFrom(query.Get("interval"), defaultMinInterval)
	if err != nil {
		return nil, fmt.Errorf("parse param interval value %s faild, %v", i, err)
	}
	intervalCalculator := newIntervalCalculator(&IntervalOptions{MinInterval: i})
	interval := intervalCalculator.Calculate(timeRange, i)
	step := time.Duration(int64(interval.Value))

	clusterOrProjectName, resourceType, id, err := parseURL(reqURL)
	if err != nil {
		return nil, err
	}

	var clusterName, nodeName, workloadName, containerName, podName, namespace string
	switch resourceType {
	case clusterstats:
		clusterName = clusterOrProjectName
		resourceType = managementv3.ResourceCluster
	case hoststats:
		clusterName = clusterOrProjectName
		nodeName = id
		resourceType = managementv3.ResourceNode
	case workloadstats:
		clusterName, _, err = parseProjectName(clusterOrProjectName)
		if err != nil {
			return nil, err
		}
		workloadName = id
		_, namespace, podName, err = parseWorkloadName(id) //todo workload not accurate
		if err != nil {
			return nil, err
		}
		resourceType = managementv3.ResourceWorkload
	case podstats:
		clusterName, _, err = parseProjectName(clusterOrProjectName)
		if err != nil {
			return nil, err
		}
		arr2 := strings.Split(id, ":")
		if len(arr2) < 2 {
			return nil, fmt.Errorf("invalid pod name: %s", id)
		}
		namespace = arr2[0]
		podName = arr2[1]
		resourceType = managementv3.ResourcePod
	case containerstats:
		clusterName, _, err = parseProjectName(clusterOrProjectName)
		if err != nil {
			return nil, err
		}
		containerName = id
		_, namespace, _, err = parseWorkloadName(query.Get("workloadName"))
		if err != nil {
			return nil, err
		}
		resourceType = managementv3.ResourceContainer
	}

	if resourceType == "" {
		return nil, fmt.Errorf("query for not supported resource type, url: %s", reqURL)
	}

	return &params{
		id:            id,
		clusterName:   clusterName,
		containerName: containerName,
		nodeName:      nodeName,
		workloadName:  workloadName,
		podName:       podName,
		resourceType:  resourceType,
		step:          step,
		start:         start,
		end:           end,
		namespace:     namespace,
	}, nil
}

func parseProjectName(id string) (clusterName, projectName string, err error) {
	arr := strings.Split(id, ":")
	if len(arr) < 2 {
		return "", "", fmt.Errorf("invalid project name: %s", id)
	}
	return arr[0], arr[1], nil
}

func parseWorkloadName(id string) (typeName, namespace, podName string, err error) {
	arr := strings.Split(id, ":")
	if len(arr) < 3 {
		return "", "", "", fmt.Errorf("invalid workload name: %s", id)
	}
	return arr[0], arr[1], arr[2], nil
}
