package client

const (
	IstioConfigSpecType            = "istioConfigSpec"
	IstioConfigSpecFieldAppAnswers = "answers"
	IstioConfigSpecFieldClusterID  = "clusterId"
	IstioConfigSpecFieldEnable     = "enable"
)

type IstioConfigSpec struct {
	AppAnswers map[string]string `json:"answers,omitempty" yaml:"answers,omitempty"`
	ClusterID  string            `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	Enable     bool              `json:"enable,omitempty" yaml:"enable,omitempty"`
}
