package common

import (
	"github.com/rancher/norman/condition"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
)

const (
	AppName                = "rancher-istio"
	AppInitVersion         = "v1.0.1"
	TemplateName           = "rancher-istio"
	IstioDeployedNamespace = "istio-system"
)

// for 1.1
// var (
// 	CertManagerSelector            = map[string]string{"app": "certmanager"}
// 	GalleySelector                 = map[string]string{"app": "galley"}
// 	GatewaysSelector               = map[string]string{"app": "gateways"}
// 	GrafanaSelector                = map[string]string{"app": "grafana"}
// 	KialiSelector                  = map[string]string{"app": "kiali"}
// 	MixerSelector                  = map[string]string{"app": "mixer"}
// 	PilotSelector                  = map[string]string{"app": "pilot"}
// 	PrometheusSelector             = map[string]string{"app": "prometheus"}
// 	SecuritySelector               = map[string]string{"app": "security"}
// 	ServicegraphSelector           = map[string]string{"app": "servicegraph"}
// 	SidecarInjectorWebhookSelector = map[string]string{"app": "sidecarInjectorWebhook"}
// 	TracingSelector                = map[string]string{"app": "tracing"}
// 	CoreDNSSelector                = map[string]string{"app": "istiocoredns"} //todo: istio 1.1 add this
// 	NodeAgentSelector              = map[string]string{"app": "nodeagent"}    //todo: istio 1.1 add this
// )

// for 1.0
var (
	CertManagerSelector = map[string]string{"app": "certmanager"}
	GalleySelector      = map[string]string{"istio": "galley"} //todo: istio 1.1 change it
	// GatewaysSelector               = map[string]string{"app": ""}         //todo: istio 1.1 change it, and too many gate in catalog
	GrafanaSelector    = map[string]string{"app": "grafana"}
	IngressSelector    = map[string]string{"istio": "ingress"} //todo: istio 1.1 remove it
	KialiSelector      = map[string]string{"app": "kiali"}
	MixerSelector      = map[string]string{"istio": "mixer"} //todo: change label
	PilotSelector      = map[string]string{"app": "pilot"}
	PrometheusSelector = map[string]string{"app": "prometheus"}
	SecuritySelector   = map[string]string{"istio": "citadel"} //todo: change label
	// ServicegraphSelector           = map[string]string{"app": "servicegraph"}
	SidecarInjectorWebhookSelector = map[string]string{"istio": "sidecar-injector"}
	// TelemetryGatewaySelector       = map[string]string{"app": "telemetry-gateway"} //todo: istio 1.1 remove it
	TracingSelector = map[string]string{"app": "jaeger"}
)

type IstioComponents struct {
	Name     string
	Cond     condition.Cond
	Selector map[string]string
}

var (
	Certmanager            = IstioComponents{Name: "certmanager", Cond: mgmtv3.IstioConditionCertManagerDeployed, Selector: CertManagerSelector}
	Galley                 = IstioComponents{Name: "galley", Cond: mgmtv3.IstioConditionGalleyDeployedDeployed, Selector: GalleySelector}
	Grafana                = IstioComponents{Name: "grafana", Cond: mgmtv3.IstioConditionGrafanaDeployedDeployed, Selector: GrafanaSelector}
	Ingress                = IstioComponents{Name: "ingress", Cond: mgmtv3.IstioConditionIngressDeployedDeployed, Selector: IngressSelector}
	Kiali                  = IstioComponents{Name: "kiali", Cond: mgmtv3.IstioConditionKialiDeployedDeployed, Selector: KialiSelector}
	Mixer                  = IstioComponents{Name: "mixer", Cond: mgmtv3.IstioConditionMixerDeployedDeployed, Selector: MixerSelector}
	Pilot                  = IstioComponents{Name: "pilot", Cond: mgmtv3.IstioConditionPilotDeployedDeployed, Selector: PilotSelector}
	Prometheus             = IstioComponents{Name: "prometheus", Cond: mgmtv3.IstioConditionPrometheusDeployedDeployed, Selector: PrometheusSelector}
	Security               = IstioComponents{Name: "security", Cond: mgmtv3.IstioConditionSecurityDeployed, Selector: SecuritySelector}
	SidecarInjectorWebhook = IstioComponents{Name: "sidecarInjectorWebhook", Cond: mgmtv3.IstioConditionSidecarInjectorWebhookDeployed, Selector: SidecarInjectorWebhookSelector}
	Tracing                = IstioComponents{Name: "tracing", Cond: mgmtv3.IstioConditionTracingDeployed, Selector: TracingSelector}
)

var (
	AllIstioComponents = []IstioComponents{
		Certmanager,
		Galley,
		Grafana,
		Ingress,
		Kiali,
		Mixer,
		Pilot,
		Prometheus,
		Security,
		SidecarInjectorWebhook,
		Tracing,
	}
)

func defaultAnswers() map[string]string {
	return map[string]string{
		"certmanager.enabled":            "true",
		"galley.enabled":                 "true",
		"gateways.enabled":               "true",
		"grafana.enabled":                "true",
		"ingress.enabled":                "true",
		"kiali.enabled":                  "true",
		"mixer.enabled":                  "true",
		"pilot.enabled":                  "true",
		"prometheus.enabled":             "true",
		"security.enabled":               "true",
		"sidecarInjectorWebhook.enabled": "true",
		"tracing.enabled":                "true",
	}
}
