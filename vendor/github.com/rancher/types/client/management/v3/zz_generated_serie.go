package client

const (
	SerieType            = "serie"
	SerieFieldExpression = "expression"
	SerieFieldName       = "name"
	SerieFieldPoints     = "points"
	SerieFieldTags       = "tags"
)

type Serie struct {
	Expression string            `json:"expression,omitempty" yaml:"expression,omitempty"`
	Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
	Points     [][]float64       `json:"points,omitempty" yaml:"points,omitempty"`
	Tags       map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}
