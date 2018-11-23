package client

const (
	QueryGraphInputType                 = "queryGraphInput"
	QueryGraphInputFieldFrom            = "from"
	QueryGraphInputFieldGraphSelector   = "graphSelector"
	QueryGraphInputFieldInterval        = "interval"
	QueryGraphInputFieldIsDetails       = "isDetails"
	QueryGraphInputFieldIsInstanceQuery = "isInstanceQuery"
	QueryGraphInputFieldMetricParams    = "metricParams"
	QueryGraphInputFieldTo              = "to"
)

type QueryGraphInput struct {
	From            string            `json:"from,omitempty" yaml:"from,omitempty"`
	GraphSelector   map[string]string `json:"graphSelector,omitempty" yaml:"graphSelector,omitempty"`
	Interval        string            `json:"interval,omitempty" yaml:"interval,omitempty"`
	IsDetails       bool              `json:"isDetails,omitempty" yaml:"isDetails,omitempty"`
	IsInstanceQuery bool              `json:"isInstanceQuery,omitempty" yaml:"isInstanceQuery,omitempty"`
	MetricParams    map[string]string `json:"metricParams,omitempty" yaml:"metricParams,omitempty"`
	To              string            `json:"to,omitempty" yaml:"to,omitempty"`
}
