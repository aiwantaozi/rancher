package client

const (
	ConnectionPoolSettingsType      = "connectionPoolSettings"
	ConnectionPoolSettingsFieldHttp = "http"
	ConnectionPoolSettingsFieldTcp  = "tcp"
)

type ConnectionPoolSettings struct {
	Http *HTTPSettings `json:"http,omitempty" yaml:"http,omitempty"`
	Tcp  *TCPSettings  `json:"tcp,omitempty" yaml:"tcp,omitempty"`
}
