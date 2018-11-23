package stats

import (
	"strings"

	clusterv3 "github.com/rancher/types/apis/cluster.cattle.io/v3"
)

const (
	ResourceNode              = "node"
	ResourceCluster           = "cluster"
	ResourceEtcd              = "etcd"
	ResourceAPIServer         = "apiserver"
	ResourceScheduler         = "scheduler"
	ResourceControllerManager = "controllermanager"
	ResourceWorkload          = "workload"
	ResourcePod               = "pod"
	ResourceContainer         = "container"
)

type ClusterGraph struct{}
type ProjectGraph struct{}

type MetricNamesOutput struct {
	Type  string   `json:"type,omitempty"`
	Names []string `json:"names" norman:"type=array[string]"`
}

type QueryGraphInput struct {
	From            string            `json:"from,omitempty"`
	To              string            `json:"to,omitempty"`
	Interval        string            `json:"interval,omitempty"`
	MetricParams    MetricParams      `json:"metricParams,omitempty"`
	GraphSelector   map[string]string `json:"graphSelector,omitempty"`
	IsDetails       bool              `json:"isDetails,omitempty"`
	IsInstanceQuery bool              `json:"isInstanceQuery,omitempty"`
}

type QueryGraphOutput struct {
	Type string `json:"type"`
	QueryGraph
}

type QueryGraphOutputCollection struct {
	Type string       `json:"type,omitempty"`
	Data []QueryGraph `json:"data,omitempty"`
}

type QueryGraph struct {
	Graph  clusterv3.MonitorGraphSpec `json:"graph"`
	Series TimeResponseSeriesSlice    `json:"series" norman:"type=array[reference[responseSeries]]"`
}

type MetricParams map[string]string

type QueryMetricInput struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Interval string `json:"interval,omitempty"`
	During   string `json:"during,omitempty"`
	Expr     string `json:"expr,omitempty" norman:"required"`
}

type QueryMetricOutput struct {
	Type   string                  `json:"type,omitempty"`
	Series TimeResponseSeriesSlice `json:"series" norman:"type=array[reference[responseSeries]]"`
}

type TimeResponseSeriesSlice []*ResponseSeries

type ResponseSeries struct {
	ID string `json:"-"`
	TimeSeries
	Expression string `json:"expression,omitempty"`
}

func replaceParams(metricParams map[string]string, expr, groupBy string) string {
	if metricParams["groupBy"] == "" {
		metricParams["groupBy"] = groupBy
	}

	var replacer []string
	for k, v := range metricParams {
		replacer = append(replacer, "$"+k)
		replacer = append(replacer, v)
	}
	srp := strings.NewReplacer(replacer...)
	return srp.Replace(expr)
}
