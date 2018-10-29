package client

const (
	ClusterLoggingInputType                        = "clusterLoggingInput"
	ClusterLoggingInputFieldDisplayName            = "displayName"
	ClusterLoggingInputFieldElasticsearchConfig    = "elasticsearchConfig"
	ClusterLoggingInputFieldFluentForwarderConfig  = "fluentForwarderConfig"
	ClusterLoggingInputFieldIncludeSystemComponent = "includeSystemComponent"
	ClusterLoggingInputFieldKafkaConfig            = "kafkaConfig"
	ClusterLoggingInputFieldOutputFlushInterval    = "outputFlushInterval"
	ClusterLoggingInputFieldOutputTags             = "outputTags"
	ClusterLoggingInputFieldSplunkConfig           = "splunkConfig"
	ClusterLoggingInputFieldSyslogConfig           = "syslogConfig"
)

type ClusterLoggingInput struct {
	DisplayName            string                 `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	ElasticsearchConfig    *ElasticsearchConfig   `json:"elasticsearchConfig,omitempty" yaml:"elasticsearchConfig,omitempty"`
	FluentForwarderConfig  *FluentForwarderConfig `json:"fluentForwarderConfig,omitempty" yaml:"fluentForwarderConfig,omitempty"`
	IncludeSystemComponent bool                   `json:"includeSystemComponent,omitempty" yaml:"includeSystemComponent,omitempty"`
	KafkaConfig            *KafkaConfig           `json:"kafkaConfig,omitempty" yaml:"kafkaConfig,omitempty"`
	OutputFlushInterval    int64                  `json:"outputFlushInterval,omitempty" yaml:"outputFlushInterval,omitempty"`
	OutputTags             map[string]string      `json:"outputTags,omitempty" yaml:"outputTags,omitempty"`
	SplunkConfig           *SplunkConfig          `json:"splunkConfig,omitempty" yaml:"splunkConfig,omitempty"`
	SyslogConfig           *SyslogConfig          `json:"syslogConfig,omitempty" yaml:"syslogConfig,omitempty"`
}
