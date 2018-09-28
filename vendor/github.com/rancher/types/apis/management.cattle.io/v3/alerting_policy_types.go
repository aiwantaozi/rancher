package v3

import (
	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterAlertPolicy struct {
	types.Namespaced

	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterPolicySpec `json:"spec"`
	// Most recent observed status of the alert. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
	Status AlertStatus `json:"status"`
}

type ProjectAlertPolicy struct {
	types.Namespaced

	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProjectPolicySpec `json:"spec"`
	// Most recent observed status of the alert. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#spec-and-status
	Status AlertStatus `json:"status"`
}

type ClusterPolicySpec struct {
	ClusterName          string                `json:"clusterName" norman:"type=reference[cluster]"`
	TargetEvents         []TargetEvent         `json:"targetEvents,omitempty"`
	TargetSystemServices []TargetSystemService `json:"targetSystemServices,omitempty"`
	CommonPolicy
}

type ProjectPolicySpec struct {
	ProjectName     string           `json:"projectName" norman:"type=reference[project]"`
	TargetPods      []TargetPod      `json:"targetPods,omitempty"`
	TargetWorkloads []TargetWorkload `json:"targetWorkloads,omitempty"`
	CommonPolicy
}

type CommonPolicy struct {
	DisplayName string      `json:"displayName,omitempty" norman:"required"`
	Description string      `json:"description,omitempty"`
	Metrics     []Metric    `json:"metrics,omitempty"`
	Recipients  []Recipient `json:"recipients,omitempty" norman:"required"`
}

type CommentField struct {
	Severity string `json:"severity,omitempty" norman:"required,options=info|critical|warning,default=critical"`
}

type Metric struct {
	CommentField
	Name               string             `json:"name,omitempty"`
	MetricType         string             `json:"metricType,omitempty"`
	LabelSelector      map[string]string  `json:"labelSelector,omitempty"`
	NameSelector       string             `json:"nameSelector,omitempty"`
	ConditionThreshold ConditionThreshold `json:"conditionThreshold,omitempty"`
}

type ConditionThreshold struct {
	Comparison     string  `json:"comparison,omitempty" norman:"required,type=enum,options=equal|not-equal|greater-than|less-than|greater-or-equal|less-or-equal"`
	Duration       string  `json:"duration",omitempty`
	ThresholdValue float64 `json:"thresholdValue,omitempty" norman:"required,type=float"`
}
