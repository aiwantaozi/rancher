package client

const (
	CommonPolicyType             = "commonPolicy"
	CommonPolicyFieldDescription = "description"
	CommonPolicyFieldDisplayName = "displayName"
	CommonPolicyFieldMetrics     = "metrics"
	CommonPolicyFieldRecipients  = "recipients"
)

type CommonPolicy struct {
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	DisplayName string      `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Metrics     []Metric    `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	Recipients  []Recipient `json:"recipients,omitempty" yaml:"recipients,omitempty"`
}
