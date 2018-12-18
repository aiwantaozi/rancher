package client

const (
	TestInputType                       = "testInput"
	TestInputFieldClusterName           = "clusterId"
	TestInputFieldElasticsearchConfig   = "elasticsearchConfig"
	TestInputFieldFluentForwarderConfig = "fluentForwarderConfig"
	TestInputFieldKafkaConfig           = "kafkaConfig"
	TestInputFieldSplunkConfig          = "splunkConfig"
	TestInputFieldSyslogConfig          = "syslogConfig"
)

type TestInput struct {
	ClusterName           string                 `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	ElasticsearchConfig   *ElasticsearchConfig   `json:"elasticsearchConfig,omitempty" yaml:"elasticsearchConfig,omitempty"`
	FluentForwarderConfig *FluentForwarderConfig `json:"fluentForwarderConfig,omitempty" yaml:"fluentForwarderConfig,omitempty"`
	KafkaConfig           *KafkaConfig           `json:"kafkaConfig,omitempty" yaml:"kafkaConfig,omitempty"`
	SplunkConfig          *SplunkConfig          `json:"splunkConfig,omitempty" yaml:"splunkConfig,omitempty"`
	SyslogConfig          *SyslogConfig          `json:"syslogConfig,omitempty" yaml:"syslogConfig,omitempty"`
}
