package client

const (
	ClusterGroupSpecType                       = "clusterGroupSpec"
	ClusterGroupSpecFieldClusterID             = "clusterId"
	ClusterGroupSpecFieldClusterRules          = "clusterRules"
	ClusterGroupSpecFieldDescription           = "description"
	ClusterGroupSpecFieldDisplayName           = "displayName"
	ClusterGroupSpecFieldGroupIntervalSeconds  = "groupIntervalSeconds"
	ClusterGroupSpecFieldGroupWaitSeconds      = "groupWaitSeconds"
	ClusterGroupSpecFieldRecipients            = "recipients"
	ClusterGroupSpecFieldRepeatIntervalSeconds = "repeatIntervalSeconds"
)

type ClusterGroupSpec struct {
	ClusterID             string             `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	ClusterRules          []ClusterAlertRule `json:"clusterRules,omitempty" yaml:"clusterRules,omitempty"`
	Description           string             `json:"description,omitempty" yaml:"description,omitempty"`
	DisplayName           string             `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	GroupIntervalSeconds  int64              `json:"groupIntervalSeconds,omitempty" yaml:"groupIntervalSeconds,omitempty"`
	GroupWaitSeconds      int64              `json:"groupWaitSeconds,omitempty" yaml:"groupWaitSeconds,omitempty"`
	Recipients            []Recipient        `json:"recipients,omitempty" yaml:"recipients,omitempty"`
	RepeatIntervalSeconds int64              `json:"repeatIntervalSeconds,omitempty" yaml:"repeatIntervalSeconds,omitempty"`
}
