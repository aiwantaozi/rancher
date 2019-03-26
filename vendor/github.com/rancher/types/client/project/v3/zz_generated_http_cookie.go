package client

const (
	HTTPCookieType      = "httpCookie"
	HTTPCookieFieldName = "name"
	HTTPCookieFieldPath = "path"
	HTTPCookieFieldTtl  = "ttl"
)

type HTTPCookie struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	Ttl  string `json:"ttl,omitempty" yaml:"ttl,omitempty"`
}
