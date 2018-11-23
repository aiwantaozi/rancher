package v3

import (
	"github.com/rancher/norman/condition"
	"github.com/rancher/norman/types"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MonitoringStatus struct {
	GrafanaEndpoint string                `json:"grafanaEndpoint,omitempty"`
	Conditions      []MonitoringCondition `json:"conditions,omitempty"`
}

type MonitoringCondition struct {
	// Type of cluster condition.
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition
	Message string `json:"message,omitempty"`
}

const (
	MonitoringConditionGrafanaDeployed           condition.Cond = "GrafanaDeployed"
	MonitoringConditionPrometheusDeployed        condition.Cond = "PrometheusDeployed"
	MonitoringConditionAlertmaanagerDeployed     condition.Cond = "AlertmanagerDeployed"
	MonitoringConditionNodeExporterDeployed      condition.Cond = "NodeExporterDeployed"
	MonitoringConditionKubeStateExporterDeployed condition.Cond = "KubeStateExporterDeployed"
)

type ClusterMonitorGraph struct {
	types.Namespaced

	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MonitorGraphSpec `json:"spec"`
}

type ProjectMonitorGraph struct {
	types.Namespaced

	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MonitorGraphSpec `json:"spec"`
}

type MonitorGraphSpec struct {
	Enable                 bool              `json:"enable,omitempty" norman:"required,default=true"`
	Title                  string            `json:"title,omitempty"`
	Description            string            `json:"description,omitempty"`
	Thresholds             float64           `json:"thresholds,omitempty" norman:"type=float"`
	MetricsSelector        map[string]string `json:"metricsSelector,omitempty"`
	DetailsMetricsSelector map[string]string `json:"detailsMetricsSelector,omitempty"`
	XAxis                  XAxis             `json:"xAxis,omitempty"`
	YAxis                  YAxis             `json:"yAxis,omitempty"`
	Priority               int               `json:"priority,omitempty"`
	Type                   string            `json:"type,omitempty" norman:"type=enum,options=graph|singlestat"`
}

type XAxis struct {
	Show bool `json:"show,omitempty"`
}

type YAxis struct {
	Show bool   `json:"show,omitempty"`
	Unit string `json:"unit,omitempty"`
}

type MonitorMetric struct {
	types.Namespaced

	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MonitorMetricSpec `json:"spec"`
}

type MonitorMetricSpec struct {
	Expression   string `json:"expression,omitempty" norman:"required"`
	LegendFormat string `json:"legendFormat,omitempty"`
	Description  string `json:"description,omitempty"`
}

type QueryGraphInput struct {
	From            string            `json:"from,omitempty"`
	To              string            `json:"to,omitempty"`
	Interval        string            `json:"interval,omitempty"`
	MetricParams    map[string]string `json:"metricParams,omitempty"`
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
	Graph  MonitorGraphSpec `json:"graph"`
	Series []*Serie         `json:"series" norman:"type=array[reference[serie]]"`
}

type QueryMetricInput struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Interval string `json:"interval,omitempty"`
	During   string `json:"during,omitempty"`
	Expr     string `json:"expr,omitempty" norman:"required"`
}

type QueryMetricOutput struct {
	Type   string   `json:"type,omitempty"`
	Series []*Serie `json:"series" norman:"type=array[reference[serie]]"`
}

type Serie struct {
	TimeSeries
	Expression string `json:"expression,omitempty"`
}

type TimeSeries struct {
	Name   string            `json:"name"`
	Points [][]float64       `json:"points" norman:"type=array[array[float]]"`
	Tags   map[string]string `json:"tags"`
}

type MetricNamesOutput struct {
	Type  string   `json:"type,omitempty"`
	Names []string `json:"names" norman:"type=array[string]"`
}
