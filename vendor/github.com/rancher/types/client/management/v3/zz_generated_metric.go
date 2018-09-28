package client

const (
	MetricType                    = "metric"
	MetricFieldConditionThreshold = "conditionThreshold"
	MetricFieldLabelSelector      = "labelSelector"
	MetricFieldMetricType         = "metricType"
	MetricFieldName               = "name"
	MetricFieldNameSelector       = "nameSelector"
	MetricFieldSeverity           = "severity"
)

type Metric struct {
	ConditionThreshold *ConditionThreshold `json:"conditionThreshold,omitempty" yaml:"conditionThreshold,omitempty"`
	LabelSelector      map[string]string   `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
	MetricType         string              `json:"metricType,omitempty" yaml:"metricType,omitempty"`
	Name               string              `json:"name,omitempty" yaml:"name,omitempty"`
	NameSelector       string              `json:"nameSelector,omitempty" yaml:"nameSelector,omitempty"`
	Severity           string              `json:"severity,omitempty" yaml:"severity,omitempty"`
}
