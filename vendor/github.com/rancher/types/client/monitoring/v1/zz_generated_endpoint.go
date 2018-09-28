package client

import "k8s.io/apimachinery/pkg/util/intstr"

const (
	EndpointType                      = "endpoint"
	EndpointFieldBasicAuth            = "basicAuth"
	EndpointFieldBearerTokenFile      = "bearerTokenFile"
	EndpointFieldHonorLabels          = "honorLabels"
	EndpointFieldInterval             = "interval"
	EndpointFieldMetricRelabelConfigs = "metricRelabelings"
	EndpointFieldParams               = "params"
	EndpointFieldPath                 = "path"
	EndpointFieldPort                 = "port"
	EndpointFieldProxyURL             = "proxyUrl"
	EndpointFieldScheme               = "scheme"
	EndpointFieldScrapeTimeout        = "scrapeTimeout"
	EndpointFieldTLSConfig            = "tlsConfig"
	EndpointFieldTargetPort           = "targetPort"
)

type Endpoint struct {
	BasicAuth            *BasicAuth          `json:"basicAuth,omitempty" yaml:"basicAuth,omitempty"`
	BearerTokenFile      string              `json:"bearerTokenFile,omitempty" yaml:"bearerTokenFile,omitempty"`
	HonorLabels          bool                `json:"honorLabels,omitempty" yaml:"honorLabels,omitempty"`
	Interval             string              `json:"interval,omitempty" yaml:"interval,omitempty"`
	MetricRelabelConfigs []RelabelConfig     `json:"metricRelabelings,omitempty" yaml:"metricRelabelings,omitempty"`
	Params               map[string][]string `json:"params,omitempty" yaml:"params,omitempty"`
	Path                 string              `json:"path,omitempty" yaml:"path,omitempty"`
	Port                 string              `json:"port,omitempty" yaml:"port,omitempty"`
	ProxyURL             string              `json:"proxyUrl,omitempty" yaml:"proxyUrl,omitempty"`
	Scheme               string              `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	ScrapeTimeout        string              `json:"scrapeTimeout,omitempty" yaml:"scrapeTimeout,omitempty"`
	TLSConfig            *TLSConfig          `json:"tlsConfig,omitempty" yaml:"tlsConfig,omitempty"`
	TargetPort           intstr.IntOrString  `json:"targetPort,omitempty" yaml:"targetPort,omitempty"`
}
