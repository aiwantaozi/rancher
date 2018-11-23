package client

const (
	QueryGraphOutputType        = "queryGraphOutput"
	QueryGraphOutputFieldGraph  = "graph"
	QueryGraphOutputFieldSeries = "series"
	QueryGraphOutputFieldType   = "type"
)

type QueryGraphOutput struct {
	Graph  *MonitorGraphSpec `json:"graph,omitempty" yaml:"graph,omitempty"`
	Series []string          `json:"series,omitempty" yaml:"series,omitempty"`
	Type   string            `json:"type,omitempty" yaml:"type,omitempty"`
}
