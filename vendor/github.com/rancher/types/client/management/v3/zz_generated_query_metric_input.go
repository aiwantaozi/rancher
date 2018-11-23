package client

const (
	QueryMetricInputType          = "queryMetricInput"
	QueryMetricInputFieldDuring   = "during"
	QueryMetricInputFieldExpr     = "expr"
	QueryMetricInputFieldFrom     = "from"
	QueryMetricInputFieldInterval = "interval"
	QueryMetricInputFieldTo       = "to"
)

type QueryMetricInput struct {
	During   string `json:"during,omitempty" yaml:"during,omitempty"`
	Expr     string `json:"expr,omitempty" yaml:"expr,omitempty"`
	From     string `json:"from,omitempty" yaml:"from,omitempty"`
	Interval string `json:"interval,omitempty" yaml:"interval,omitempty"`
	To       string `json:"to,omitempty" yaml:"to,omitempty"`
}
