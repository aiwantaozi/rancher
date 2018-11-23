package client

const (
	YAxisType      = "yAxis"
	YAxisFieldShow = "show"
	YAxisFieldUnit = "unit"
)

type YAxis struct {
	Show bool   `json:"show,omitempty" yaml:"show,omitempty"`
	Unit string `json:"unit,omitempty" yaml:"unit,omitempty"`
}
