package v3

import (
	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatsStatus struct {
	MonitorGraphIDs []string `json:"monitorGraphIds,omitempty" norman:"type=array[reference[monitorGraph]]"`
}

type MonitorGraph struct {
	types.Namespaced

	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitorGraphSpec   `json:"spec"`
	Status MonitorGraphStatus `json:"status"`
}

type MonitorGraphSpec struct {
	Enable                 bool              `json:"enable,omitempty" norman:"required,default=true"`
	Title                  string            `json:"title,omitempty"`
	Description            string            `json:"description,omitempty"`
	Thresholds             float64           `json:"thresholds,omitempty" norman:"type=float"`
	MetricsSelector        map[string]string `json:"metricsSelector,omitempty"`
	DetailsMetricsSelector map[string]string `json:"detailsMetricsSelector,omitempty"`
	XAxis                  XAxis             `json:"xAxis,omitempty"`
	YAxis                  YAxis             `json:"yAxis,omitemoty"`
	Priority               int               `json:"priority,omitemoty"`
}

type MonitorGraphStatus struct {
	MonitorMetricIDs []string `json:"monitorMetricIds,omitempty" norman:"type=array[reference[monitorMetric]]"`
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
	Expression          string            `json:"expression,omitempty" norman:"required"`
	LegendFormat        string            `json:"legendFormat,omitempty"`
	DetailsLegendFormat string            `json:"detailsLegendFormat,omitempty"`
	Step                int64             `json:"step,omitempty"`
	Description         string            `json:"description,omitempty"`
	ExtraAddedTags      map[string]string `json:"extraAddedTags,omitempty"`
	GroupBy             string            `json:"groupBy,omitempty"`
	DetailsGroupBy      string            `json:"detailsGroupBy,omitempty"`
}
