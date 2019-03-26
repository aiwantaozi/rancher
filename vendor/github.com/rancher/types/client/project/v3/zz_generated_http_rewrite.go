package client

const (
	HTTPRewriteType           = "httpRewrite"
	HTTPRewriteFieldAuthority = "authority"
	HTTPRewriteFieldUri       = "uri"
)

type HTTPRewrite struct {
	Authority string `json:"authority,omitempty" yaml:"authority,omitempty"`
	Uri       string `json:"uri,omitempty" yaml:"uri,omitempty"`
}
