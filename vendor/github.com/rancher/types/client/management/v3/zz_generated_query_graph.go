package client

const (
	QueryGraphType        = "queryGraph"
	QueryGraphFieldGraph  = "graph"
	QueryGraphFieldSeries = "series"
)

type QueryGraph struct {
	Graph  *MonitorGraphSpec `json:"graph,omitempty" yaml:"graph,omitempty"`
	Series []string          `json:"series,omitempty" yaml:"series,omitempty"`
}
