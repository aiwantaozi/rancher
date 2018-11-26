package client

const (
	ClusterMetricNamesInputType           = "clusterMetricNamesInput"
	ClusterMetricNamesInputFieldClusterID = "clusterId"
)

type ClusterMetricNamesInput struct {
	ClusterID string `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
}
