package client

const (
	HTTPRouteType                       = "httpRoute"
	HTTPRouteFieldAppendHeaders         = "appendHeaders"
	HTTPRouteFieldCorsPolicy            = "corsPolicy"
	HTTPRouteFieldFault                 = "fault"
	HTTPRouteFieldMatch                 = "match"
	HTTPRouteFieldMirror                = "mirror"
	HTTPRouteFieldRedirect              = "redirect"
	HTTPRouteFieldRemoveResponseHeaders = "removeResponseHeaders"
	HTTPRouteFieldRetries               = "retries"
	HTTPRouteFieldRewrite               = "rewrite"
	HTTPRouteFieldRoute                 = "route"
	HTTPRouteFieldTimeout               = "timeout"
	HTTPRouteFieldWebsocketUpgrade      = "websocketUpgrade"
)

type HTTPRoute struct {
	AppendHeaders         map[string]string   `json:"appendHeaders,omitempty" yaml:"appendHeaders,omitempty"`
	CorsPolicy            *CorsPolicy         `json:"corsPolicy,omitempty" yaml:"corsPolicy,omitempty"`
	Fault                 *HTTPFaultInjection `json:"fault,omitempty" yaml:"fault,omitempty"`
	Match                 []HTTPMatchRequest  `json:"match,omitempty" yaml:"match,omitempty"`
	Mirror                *Destination        `json:"mirror,omitempty" yaml:"mirror,omitempty"`
	Redirect              *HTTPRedirect       `json:"redirect,omitempty" yaml:"redirect,omitempty"`
	RemoveResponseHeaders map[string]string   `json:"removeResponseHeaders,omitempty" yaml:"removeResponseHeaders,omitempty"`
	Retries               *HTTPRetry          `json:"retries,omitempty" yaml:"retries,omitempty"`
	Rewrite               *HTTPRewrite        `json:"rewrite,omitempty" yaml:"rewrite,omitempty"`
	Route                 []DestinationWeight `json:"route,omitempty" yaml:"route,omitempty"`
	Timeout               string              `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	WebsocketUpgrade      bool                `json:"websocketUpgrade,omitempty" yaml:"websocketUpgrade,omitempty"`
}
