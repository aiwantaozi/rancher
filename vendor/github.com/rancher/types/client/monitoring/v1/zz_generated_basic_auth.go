package client

const (
	BasicAuthType          = "basicAuth"
	BasicAuthFieldPassword = "password"
	BasicAuthFieldUsername = "username"
)

type BasicAuth struct {
	Password *SecretKeySelector `json:"password,omitempty" yaml:"password,omitempty"`
	Username *SecretKeySelector `json:"username,omitempty" yaml:"username,omitempty"`
}
