package constant

import (
	"fmt"
)

const (
	AppName           = "rancher-logging"
	TesterAppName     = "rancher-logging-tester"
	AppInitVersion    = "0.0.1"
	systemCatalogName = "system-library"
	templateName      = "rancher-logging"
)

const (
	LoggingNamespace = "cattle-logging"
)

//daemonset, pod name
const (
	FluentdName       = "fluentd"
	FluentdHelperName = "fluentd-helper"
	LogAggregatorName = "log-aggregator"
	FluentdTesterName = "fluentd-test"
)

//config
const (
	LoggingSecretName             = "fluentd"
	LoggingSecretClusterConfigKey = "cluster.conf"
	LoggingSecretProjectConfigKey = "project.conf"
)

//target
const (
	Elasticsearch   = "elasticsearch"
	Splunk          = "splunk"
	Kafka           = "kafka"
	Syslog          = "syslog"
	FluentForwarder = "fluentforwarder"
	CustomTarget    = "customtarget"
)

const (
	GoogleKubernetesEngine = "googleKubernetesEngine"
)

//ssl
const (
	CaFileName     = "ca.pem"
	ClientCertName = "client-cert.pem"
	ClientKeyName  = "client-key.pem"
)

const (
	ClusterLevel = "cluster"
	ProjectLevel = "project"
)

func SecretDataKeyCa(level, name string) string {
	return fmt.Sprintf("%s_%s_%s", level, name, CaFileName)
}

func SecretDataKeyCert(level, name string) string {
	return fmt.Sprintf("%s_%s_%s", level, name, ClientCertName)
}

func SecretDataKeyCertKey(level, name string) string {
	return fmt.Sprintf("%s_%s_%s", level, name, ClientKeyName)
}

func RancherLoggingTemplateID() string {
	return fmt.Sprintf("%s-%s", systemCatalogName, templateName)
}

func RancherLoggingCatalogID(version string) string {
	return fmt.Sprintf("catalog://?catalog=%s&template=%s&version=%s", systemCatalogName, templateName, version)
}

func RancherLoggingConfigSecretName() string {
	return fmt.Sprintf("%s-%s", AppName, LoggingSecretName)
}
