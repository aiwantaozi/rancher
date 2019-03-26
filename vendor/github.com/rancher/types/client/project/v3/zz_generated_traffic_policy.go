package client

const (
	TrafficPolicyType                   = "trafficPolicy"
	TrafficPolicyFieldConnectionPool    = "connectionPool"
	TrafficPolicyFieldLoadBalancer      = "loadBalancer"
	TrafficPolicyFieldOutlierDetection  = "outlierDetection"
	TrafficPolicyFieldPortLevelSettings = "portLevelSettings"
	TrafficPolicyFieldTls               = "tls"
)

type TrafficPolicy struct {
	ConnectionPool    *ConnectionPoolSettings `json:"connectionPool,omitempty" yaml:"connectionPool,omitempty"`
	LoadBalancer      *LoadBalancerSettings   `json:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty"`
	OutlierDetection  *OutlierDetection       `json:"outlierDetection,omitempty" yaml:"outlierDetection,omitempty"`
	PortLevelSettings []PortTrafficPolicy     `json:"portLevelSettings,omitempty" yaml:"portLevelSettings,omitempty"`
	Tls               *TLSSettings            `json:"tls,omitempty" yaml:"tls,omitempty"`
}
