package client

const (
	LoggingInputType           = "loggingInput"
	LoggingInputFieldEndpoints = "endpoint"
	LoggingInputFieldProtocol  = "protocol"
)

type LoggingInput struct {
	Endpoints []string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Protocol  string   `json:"protocol,omitempty" yaml:"protocol,omitempty"`
}
