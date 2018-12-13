package constant

import (
	"fmt"
)

const (
	LoggingNamespace   = "cattle-logging"
	ClusterLoggingName = "cluster-logging"
	ProjectLoggingName = "project-logging"
)

//daemonset
const (
	FluentdName       = "fluentd"
	FluentdHelperName = "fluentd-helper"
	LogAggregatorName = "log-aggregator"
)

//config
const (
	ClusterFileName   = "cluster.conf"
	ProjectFileName   = "project.conf"
	ClusterConfigPath = "/tmp/cluster.conf"
	ProjectConfigPath = "/tmp/project.conf"
)

//target
const (
	Elasticsearch   = "elasticsearch"
	Splunk          = "splunk"
	Kafka           = "kafka"
	Syslog          = "syslog"
	FluentForwarder = "fluentforwarder"
)

//app label
const (
	LabelK8sApp = "k8s-app"
)

const (
	GoogleKubernetesEngine = "googleKubernetesEngine"
)

//ssl
const (
	SSLSecretName  = "sslconfig"
	CaFileName     = "ca.pem"
	ClientCertName = "client-cert.pem"
	ClientKeyName  = "client-key.pem"
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
