package client

const (
	XAxisType      = "xAxis"
	XAxisFieldShow = "show"
)

type XAxis struct {
	Show bool `json:"show,omitempty" yaml:"show,omitempty"`
}
