package client

const (
	HTTPRedirectType           = "httpRedirect"
	HTTPRedirectFieldAuthority = "authority"
	HTTPRedirectFieldUri       = "uri"
)

type HTTPRedirect struct {
	Authority string `json:"authority,omitempty" yaml:"authority,omitempty"`
	Uri       string `json:"uri,omitempty" yaml:"uri,omitempty"`
}
