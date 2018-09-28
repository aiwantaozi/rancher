package client

const (
	ProjectPolicySpecType                 = "projectPolicySpec"
	ProjectPolicySpecFieldDescription     = "description"
	ProjectPolicySpecFieldDisplayName     = "displayName"
	ProjectPolicySpecFieldMetrics         = "metrics"
	ProjectPolicySpecFieldProjectID       = "projectId"
	ProjectPolicySpecFieldRecipients      = "recipients"
	ProjectPolicySpecFieldTargetPods      = "targetPods"
	ProjectPolicySpecFieldTargetWorkloads = "targetWorkloads"
)

type ProjectPolicySpec struct {
	Description     string           `json:"description,omitempty" yaml:"description,omitempty"`
	DisplayName     string           `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Metrics         []Metric         `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	ProjectID       string           `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	Recipients      []Recipient      `json:"recipients,omitempty" yaml:"recipients,omitempty"`
	TargetPods      []TargetPod      `json:"targetPods,omitempty" yaml:"targetPods,omitempty"`
	TargetWorkloads []TargetWorkload `json:"targetWorkloads,omitempty" yaml:"targetWorkloads,omitempty"`
}
