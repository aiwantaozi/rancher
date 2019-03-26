package v3

import (
	"github.com/rancher/norman/condition"
	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	IstioConditionAppInstalled                   condition.Cond = "IstioAppInstalled"
	IstioConditionCertManagerDeployed            condition.Cond = "IstioCertManagerDeployed"
	IstioConditionGalleyDeployedDeployed         condition.Cond = "IstioGalleyDeployed"
	IstioConditionGatewaysDeployedDeployed       condition.Cond = "IstioGatewaysDeployed"
	IstioConditionGrafanaDeployedDeployed        condition.Cond = "IstioGrafanaDeployed"
	IstioConditionKialiDeployedDeployed          condition.Cond = "IstioKialiDeployed"
	IstioConditionMixerDeployedDeployed          condition.Cond = "IstioMixerDeployed"
	IstioConditionPilotDeployedDeployed          condition.Cond = "IstioPilotDeployed"
	IstioConditionPrometheusDeployedDeployed     condition.Cond = "IstioPrometheusDeployed"
	IstioConditionSecurityDeployed               condition.Cond = "IstioSecurityDeployed"
	IstioConditionServiceGraphDeployed           condition.Cond = "IstioServiceGraphDeployed"
	IstioConditionSidecarInjectorWebhookDeployed condition.Cond = "IstioSidecarInjectorWebhookDeployed"
	IstioConditionTracingDeployed                condition.Cond = "IstioTracingDeployed"

	// istio 1.1 added
	IstioConditionCoreDNSDeployed   condition.Cond = "CoreDNSDeployed"
	IstioConditionNodeAgentDeployed condition.Cond = "NodeAgentDeployed"
	// todo: istio 1.1 removed
	IstioConditionIngressDeployedDeployed  condition.Cond = "IngressDeployed"
	IstioConditionTelemetryGatewayDeployed condition.Cond = "TelemetryGatewayDeployed"
)

type IstioConfig struct {
	types.Namespaced

	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IstioConfigSpec `json:"spec"`

	Status IstioConfigStatus `json:"status"`
}

type IstioConfigSpec struct {
	ClusterName string            `json:"clusterName" norman:"type=reference[cluster]"`
	Enable      bool              `json:"enable,omitempty" norman:"required,default=false"`
	AppAnswers  map[string]string `json:"answers,omitempty"`
}

type IstioConfigStatus struct {
	Conditions  []Condition     `json:"conditions"`
	AppliedSpec IstioConfigSpec `json:"appliedSpec,omitempty"`
}
