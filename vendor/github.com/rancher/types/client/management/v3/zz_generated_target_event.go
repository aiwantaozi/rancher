package client

const (
	TargetEventType              = "targetEvent"
	TargetEventFieldEventType    = "eventType"
	TargetEventFieldResourceKind = "resourceKind"
	TargetEventFieldRuleID       = "ruleID"
	TargetEventFieldSeverity     = "severity"
)

type TargetEvent struct {
	EventType    string `json:"eventType,omitempty" yaml:"eventType,omitempty"`
	ResourceKind string `json:"resourceKind,omitempty" yaml:"resourceKind,omitempty"`
	RuleID       string `json:"ruleID,omitempty" yaml:"ruleID,omitempty"`
	Severity     string `json:"severity,omitempty" yaml:"severity,omitempty"`
}
