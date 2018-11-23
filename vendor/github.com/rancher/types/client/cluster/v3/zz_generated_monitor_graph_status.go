package client

const (
	MonitorGraphStatusType                  = "monitorGraphStatus"
	MonitorGraphStatusFieldMonitorMetricIDs = "monitorMetricIds"
)

type MonitorGraphStatus struct {
	MonitorMetricIDs []string `json:"monitorMetricIds,omitempty" yaml:"monitorMetricIds,omitempty"`
}
