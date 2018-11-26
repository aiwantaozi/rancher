package client

const (
	ClusterMonitorGraphSpecType                        = "clusterMonitorGraphSpec"
	ClusterMonitorGraphSpecFieldClusterID              = "clusterId"
	ClusterMonitorGraphSpecFieldDescription            = "description"
	ClusterMonitorGraphSpecFieldDetailsMetricsSelector = "detailsMetricsSelector"
	ClusterMonitorGraphSpecFieldDisplayResourceType    = "displayResourceType"
	ClusterMonitorGraphSpecFieldMetricsSelector        = "metricsSelector"
	ClusterMonitorGraphSpecFieldPriority               = "priority"
	ClusterMonitorGraphSpecFieldResourceType           = "resourceType"
	ClusterMonitorGraphSpecFieldType                   = "type"
	ClusterMonitorGraphSpecFieldYAxis                  = "yAxis"
)

type ClusterMonitorGraphSpec struct {
	ClusterID              string            `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	Description            string            `json:"description,omitempty" yaml:"description,omitempty"`
	DetailsMetricsSelector map[string]string `json:"detailsMetricsSelector,omitempty" yaml:"detailsMetricsSelector,omitempty"`
	DisplayResourceType    string            `json:"displayResourceType,omitempty" yaml:"displayResourceType,omitempty"`
	MetricsSelector        map[string]string `json:"metricsSelector,omitempty" yaml:"metricsSelector,omitempty"`
	Priority               int64             `json:"priority,omitempty" yaml:"priority,omitempty"`
	ResourceType           string            `json:"resourceType,omitempty" yaml:"resourceType,omitempty"`
	Type                   string            `json:"type,omitempty" yaml:"type,omitempty"`
	YAxis                  *YAxis            `json:"yAxis,omitempty" yaml:"yAxis,omitempty"`
}
