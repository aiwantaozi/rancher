package client

const (
	QueryProjectMetricInputType           = "queryProjectMetricInput"
	QueryProjectMetricInputFieldExpr      = "expr"
	QueryProjectMetricInputFieldFrom      = "from"
	QueryProjectMetricInputFieldInterval  = "interval"
	QueryProjectMetricInputFieldProjectID = "projectId"
	QueryProjectMetricInputFieldTo        = "to"
)

type QueryProjectMetricInput struct {
	Expr      string `json:"expr,omitempty" yaml:"expr,omitempty"`
	From      string `json:"from,omitempty" yaml:"from,omitempty"`
	Interval  string `json:"interval,omitempty" yaml:"interval,omitempty"`
	ProjectID string `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	To        string `json:"to,omitempty" yaml:"to,omitempty"`
}
