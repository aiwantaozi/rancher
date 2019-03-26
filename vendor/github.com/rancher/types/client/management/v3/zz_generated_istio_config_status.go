package client

const (
	IstioConfigStatusType             = "istioConfigStatus"
	IstioConfigStatusFieldAppliedSpec = "appliedSpec"
	IstioConfigStatusFieldConditions  = "conditions"
)

type IstioConfigStatus struct {
	AppliedSpec *IstioConfigSpec `json:"appliedSpec,omitempty" yaml:"appliedSpec,omitempty"`
	Conditions  []Condition      `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}
