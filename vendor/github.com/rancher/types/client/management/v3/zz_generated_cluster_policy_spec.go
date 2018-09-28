package client

const (
	ClusterPolicySpecType                      = "clusterPolicySpec"
	ClusterPolicySpecFieldClusterID            = "clusterId"
	ClusterPolicySpecFieldDescription          = "description"
	ClusterPolicySpecFieldDisplayName          = "displayName"
	ClusterPolicySpecFieldMetrics              = "metrics"
	ClusterPolicySpecFieldRecipients           = "recipients"
	ClusterPolicySpecFieldTargetEvents         = "targetEvents"
	ClusterPolicySpecFieldTargetSystemServices = "targetSystemServices"
)

type ClusterPolicySpec struct {
	ClusterID            string                `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	Description          string                `json:"description,omitempty" yaml:"description,omitempty"`
	DisplayName          string                `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Metrics              []Metric              `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	Recipients           []Recipient           `json:"recipients,omitempty" yaml:"recipients,omitempty"`
	TargetEvents         []TargetEvent         `json:"targetEvents,omitempty" yaml:"targetEvents,omitempty"`
	TargetSystemServices []TargetSystemService `json:"targetSystemServices,omitempty" yaml:"targetSystemServices,omitempty"`
}
