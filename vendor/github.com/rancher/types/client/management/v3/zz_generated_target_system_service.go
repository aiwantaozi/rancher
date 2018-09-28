package client

const (
	TargetSystemServiceType           = "targetSystemService"
	TargetSystemServiceFieldCondition = "condition"
	TargetSystemServiceFieldRuleID    = "ruleID"
	TargetSystemServiceFieldSeverity  = "severity"
)

type TargetSystemService struct {
	Condition string `json:"condition,omitempty" yaml:"condition,omitempty"`
	RuleID    string `json:"ruleID,omitempty" yaml:"ruleID,omitempty"`
	Severity  string `json:"severity,omitempty" yaml:"severity,omitempty"`
}
