package client

const (
	ConditionThresholdType                = "conditionThreshold"
	ConditionThresholdFieldComparison     = "comparison"
	ConditionThresholdFieldDuration       = "duration"
	ConditionThresholdFieldThresholdValue = "thresholdValue"
)

type ConditionThreshold struct {
	Comparison     string   `json:"comparison,omitempty" yaml:"comparison,omitempty"`
	Duration       string   `json:"duration,omitempty" yaml:"duration,omitempty"`
	ThresholdValue *float64 `json:"thresholdValue,omitempty" yaml:"thresholdValue,omitempty"`
}
