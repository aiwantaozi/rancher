package client

const (
	MonitorMetricSpecType                     = "monitorMetricSpec"
	MonitorMetricSpecFieldDescription         = "description"
	MonitorMetricSpecFieldDetailsGroupBy      = "detailsGroupBy"
	MonitorMetricSpecFieldDetailsLegendFormat = "detailsLegendFormat"
	MonitorMetricSpecFieldExpression          = "expression"
	MonitorMetricSpecFieldExtraAddedTags      = "extraAddedTags"
	MonitorMetricSpecFieldGroupBy             = "groupBy"
	MonitorMetricSpecFieldLegendFormat        = "legendFormat"
	MonitorMetricSpecFieldStep                = "step"
)

type MonitorMetricSpec struct {
	Description         string            `json:"description,omitempty" yaml:"description,omitempty"`
	DetailsGroupBy      string            `json:"detailsGroupBy,omitempty" yaml:"detailsGroupBy,omitempty"`
	DetailsLegendFormat string            `json:"detailsLegendFormat,omitempty" yaml:"detailsLegendFormat,omitempty"`
	Expression          string            `json:"expression,omitempty" yaml:"expression,omitempty"`
	ExtraAddedTags      map[string]string `json:"extraAddedTags,omitempty" yaml:"extraAddedTags,omitempty"`
	GroupBy             string            `json:"groupBy,omitempty" yaml:"groupBy,omitempty"`
	LegendFormat        string            `json:"legendFormat,omitempty" yaml:"legendFormat,omitempty"`
	Step                int64             `json:"step,omitempty" yaml:"step,omitempty"`
}
