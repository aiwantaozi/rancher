package client

const (
	ConsistentHashLBType                 = "consistentHashLB"
	ConsistentHashLBFieldHttpCookie      = "httpCookie"
	ConsistentHashLBFieldHttpHeaderName  = "httpHeaderName"
	ConsistentHashLBFieldMinimumRingSize = "minimumRingSize"
	ConsistentHashLBFieldUseSourceIp     = "useSourceIp"
)

type ConsistentHashLB struct {
	HttpCookie      *HTTPCookie `json:"httpCookie,omitempty" yaml:"httpCookie,omitempty"`
	HttpHeaderName  string      `json:"httpHeaderName,omitempty" yaml:"httpHeaderName,omitempty"`
	MinimumRingSize int64       `json:"minimumRingSize,omitempty" yaml:"minimumRingSize,omitempty"`
	UseSourceIp     bool        `json:"useSourceIp,omitempty" yaml:"useSourceIp,omitempty"`
}
