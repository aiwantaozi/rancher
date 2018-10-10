package stats

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/sirupsen/logrus"
)

const (
	DefaultDuring = "2m"
)

const (
	CPUUsageSeconds               = "cpu_usage_seconds_total"
	CPUUserSeconds                = "cpu_user_seconds_total"
	CPUSystemSeconds              = "cpu_system_seconds_total"
	CPUCfsThrottledSeconds        = "cpu_cfs_throttled_seconds_total"
	NetworkReceiveBytes           = "network_receive_bytes_total"
	NetworkReceivePacketsDropped  = "network_receive_packets_dropped_total"
	NetworkReceiveErrors          = "network_receive_errors_total"
	NetworkReceivePackets         = "network_receive_packets_total"
	NetworkTransmitBytes          = "network_transmit_bytes_total"
	NetworkTransmitPacketsDropped = "network_transmit_packets_dropped_total"
	NetworkTransmitErrors         = "network_transmit_errors_total"
	NetworkTransmitPackets        = "network_transmit_packets_total"
	DiskBytesRead                 = "fs_reads_bytes_total"
	DiskBytesWrite                = "fs_writes_bytes_total"
	NodeLoad1                     = "node_load1"
	NodeLoad5                     = "node_load5"
	NodeLoad15                    = "node_load15"
)

var (
	templateMap = func(m map[string]string) map[string]*template.Template {
		rtn := map[string]*template.Template{}
		for k, v := range m {
			rtn[k] = template.Must(template.New(k).Parse(v))
		}
		return rtn
	}(map[string]string{
		v3.CPULoad1:                             `sum(node_load1{instance=~"{{ .HostName}}"}) / count(node_cpu_seconds_total{mode="system",instance=~"{{ .HostName}}"}) by ({{ .GroupBy}})`,
		v3.CPULoad5:                             `sum(node_load5{instance=~"{{ .HostName}}"}) / count(node_cpu_seconds_total{mode="system",instance=~"{{ .HostName}}"}) by ({{ .GroupBy}})`,
		v3.CPULoad15:                            `sum(node_load15{instance=~"{{ .HostName}}"}) / count(node_cpu_seconds_total{mode="system",instance=~"{{ .HostName}}"}) by ({{ .GroupBy}})`,
		v3.CPUUsageSecondsSumRate:               `sum(rate(node_cpu_seconds_total{mode!="idle", mode!="iowait", mode!~"^(?:guest.*)$", instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.CPUUserSecondsSumRate:                `sum(rate(node_cpu_seconds_total{mode="user", instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.CPUSystemSecondsSumRate:              `sum(rate(node_cpu_seconds_total{mode="system", instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.MemoryUsagePercent:                   `1 - sum(node_memory_MemAvailable{instance=~"{{ .HostName}}"}) / sum(node_memory_MemTotal{instance=~"{{ .HostName}}"}) by ({{ .GroupBy}})`,
		v3.MemoryPageInBytesSumRate:             `1e3 * sum(rate(node_vmstat_pgpgin{instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.MemoryPageOutBytesSumRate:            `1e3 * sum(rate(node_vmstat_pgpgout{instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.NetworkReceiveBytesSumRate:           `sum(rate(node_network_receive_bytes{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*",instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.NetworkReceivePacketsDroppedSumRate:  `sum(rate(node_network_receive_drop{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*",instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.NetworkReceiveErrorsSumRate:          `sum(rate(node_network_receive_errs{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*",instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.NetworkReceivePacketsSumRate:         `sum(rate(node_network_receive_packets{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*",instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.NetworkTransmitBytesSumRate:          `sum(rate(node_network_transmit_bytes{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*",instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.NetworkTransmitPacketsDroppedSumRate: `sum(rate(node_network_transmit_drop{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*",instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.NetworkTransmitPacketsSumRate:        `sum(rate(node_network_transmit_packets{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*",instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.NetworkTransmitErrorsSumRate:         `sum(rate(node_network_transmit_errs{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*",instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.FilesystemUseagePercent:              `(sum(node_filesystem_size{mountpoint="/",instance=~"{{ .HostName}}"}) - sum(node_filesystem_free{mountpoint="/",instance=~"{{ .HostName}}"})) / sum(node_filesystem_size{mountpoint="/",instance=~"{{ .HostName}}"}) by ({{ .GroupBy}})`,
		v3.DiskIOReadsBytesSumRate:              `sum(rate(node_disk_bytes_read{instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,
		v3.DiskIOWritesBytesSumRate:             `sum(rate(node_disk_bytes_written{instance=~"{{ .HostName}}"}[{{ .During}}])) by ({{ .GroupBy}})`,

		v3.ResourcePod + v3.CPUUsageSecondsSumRate:               getPodSumRateExp(CPUUsageSeconds),
		v3.ResourcePod + v3.CPUUserSecondsSumRate:                getPodSumRateExp(CPUUserSeconds),
		v3.ResourcePod + v3.CPUSystemSecondsSumRate:              getPodSumRateExp(CPUSystemSeconds),
		v3.ResourcePod + v3.CPUCfsThrottledSecondsSumRate:        getPodSumRateExp(CPUCfsThrottledSeconds),
		v3.ResourcePod + v3.MemoryUsageBytesSum:                  `sum(container_memory_working_set_bytes{name!~"POD", namespace="{{ .Namespace}}",pod_name=~"{{ .PodName}}"}) by ({{ .GroupBy}})`,
		v3.ResourcePod + v3.MemoryUsagePercent:                   `sum(container_memory_working_set_bytes{namespace="{{ .Namespace}}", pod_name=~"{{ .PodName}}"}) / sum(label_join(kube_pod_container_resource_limits_memory_bytes{namespace="{{ .Namespace}}", pod=~"{{ .PodName}}"},"pod_name", "", "pod")) by ({{ .GroupBy}})`,
		v3.ResourcePod + v3.DiskIOReadsBytesSumRate:              getPodSumRateExp(DiskBytesRead),
		v3.ResourcePod + v3.DiskIOWritesBytesSumRate:             getPodSumRateExp(DiskBytesWrite),
		v3.ResourcePod + v3.FsByteSum:                            `sum(container_fs_usage_bytes{namespace="{{ .Namespace}}", pod_name=~"{{ .PodName}}"}) by ({{ .GroupBy}})`,
		v3.ResourcePod + v3.NetworkReceiveBytesSumRate:           getPodSumRateExp(NetworkReceiveBytes),
		v3.ResourcePod + v3.NetworkTransmitBytesSumRate:          getPodSumRateExp(NetworkTransmitBytes),
		v3.ResourcePod + v3.NetworkReceiveErrorsSumRate:          getPodSumRateExp(NetworkReceiveErrors),
		v3.ResourcePod + v3.NetworkTransmitErrorsSumRate:         getPodSumRateExp(NetworkTransmitErrors),
		v3.ResourcePod + v3.NetworkReceivePacketsSumRate:         getPodSumRateExp(NetworkReceivePackets),
		v3.ResourcePod + v3.NetworkTransmitPacketsSumRate:        getPodSumRateExp(NetworkTransmitPackets),
		v3.ResourcePod + v3.NetworkReceivePacketsDroppedSumRate:  getPodSumRateExp(NetworkReceivePacketsDropped),
		v3.ResourcePod + v3.NetworkTransmitPacketsDroppedSumRate: getPodSumRateExp(NetworkTransmitPacketsDropped),

		v3.ResourceContainer + v3.CPUUsageSecondsSumRate:        getContainerSumRateExp(CPUUsageSeconds),
		v3.ResourceContainer + v3.CPUUserSecondsSumRate:         getContainerSumRateExp(CPUUserSeconds),
		v3.ResourceContainer + v3.CPUSystemSecondsSumRate:       getContainerSumRateExp(CPUSystemSeconds),
		v3.ResourceContainer + v3.CPUCfsThrottledSecondsSumRate: getContainerSumRateExp(CPUCfsThrottledSeconds),
		v3.ResourceContainer + v3.MemoryUsageBytesSum:           `sum(container_memory_working_set_bytes{name!~"POD", namespace="{{ .Namespace}}",container_name=~"{{ .ContainerName}}"}) by ({{ .GroupBy}})`,
		v3.ResourceContainer + v3.MemoryUsagePercent:            `sum(container_memory_working_set_bytes{namespace="{{ .Namespace}}", container_name=~"{{ .ContainerName}}"}) / sum(label_join(kube_container_container_resource_limits_memory_bytes{namespace="{{ .Namespace}}",container=~"{{ .ContainerName}}"},"container_name", "", "container")) by ({{ .GroupBy}})`,
		v3.ResourceContainer + v3.DiskIOReadsBytesSumRate:       getContainerSumRateExp(DiskBytesRead),
		v3.ResourceContainer + v3.DiskIOWritesBytesSumRate:      getContainerSumRateExp(DiskBytesWrite),
		v3.ResourceContainer + v3.FsByteSum:                     `sum(container_fs_usage_bytes{namespace="{{ .Namespace}}", container_name=~"{{ .ContainerName}}"}) by ({{ .GroupBy}})`,
	})
)

func execute(expName string, data interface{}) string {
	t, ok := templateMap[expName]
	if !ok {
		return ""
	}
	buffer := bytes.NewBuffer([]byte{})
	if err := t.Execute(buffer, data); err != nil {
		logrus.Warnf("execute %s failed, data: %+v, error: %v", expName, data, err)
		return ""
	}

	return buffer.String()
}

func GetExp(exptype, expName string, data interface{}) string {
	switch exptype {
	case v3.ResourceNode, v3.ResourceCluster:
		return execute(expName, data)
	case v3.ResourceContainer:
		return execute(v3.ResourceContainer+expName, data)
	case v3.ResourceWorkload, v3.ResourcePod:
		fallthrough
	default:
		return execute(v3.ResourcePod+expName, data)
	}
}

func getPodSumRateExp(metrics string) string {
	return fmt.Sprintf(`sum(rate(%s_%s{namespace="{{ .Namespace}}",pod_name=~"{{ .PodName}}"}[{{ .During}}])) by ({{ .GroupBy}})`, v3.ResourceContainer, metrics)
}

func getContainerSumRateExp(metrics string) string {
	return fmt.Sprintf(`sum(rate(%s_%s{namespace="{{ .Namespace}}",container_name=~"{{ .ContainerName}}.*"}[{{ .During}}])) by ({{ .GroupBy}})`, v3.ResourceContainer, metrics)
}

type ResourceQuery struct {
	ContainerName string
	During        string
	HostName      string
	Namespace     string
	PodName       string
	GroupBy       string
}

type MetricQueryData interface {
	Generate() interface{}
}

type ContainerQueryData struct {
	ResourceQuery
}
type PodQueryQueryData struct {
	ResourceQuery
}
type WorkloadQueryData struct {
	ResourceQuery
}
type ClusterQueryData struct {
	ResourceQuery
}
type NodeQueryData struct {
	ResourceQuery
}

func NewMetricQueryData(resourceType string, resoureQuey ResourceQuery) MetricQueryData {
	switch resourceType {
	case v3.ResourceNode:
		return &NodeQueryData{resoureQuey}
	case v3.ResourceWorkload, v3.ResourcePod:
		return &WorkloadQueryData{resoureQuey}
	case v3.ResourceContainer:
		return &ContainerQueryData{resoureQuey}
	default:
		return &ClusterQueryData{resoureQuey}
	}
}

func (s *WorkloadQueryData) Generate() interface{} {
	return struct {
		Namespace string
		PodName   string
		During    string
		GroupBy   string
	}{
		Namespace: s.Namespace,
		PodName:   s.PodName,
		During:    s.During,
		GroupBy:   s.GroupBy,
	}
}

func (s *ContainerQueryData) Generate() interface{} {
	return struct {
		Namespace     string
		ContainerName string
		During        string
		GroupBy       string
	}{
		Namespace:     s.Namespace,
		ContainerName: s.ContainerName,
		During:        s.During,
		GroupBy:       s.GroupBy,
	}
}

func (s *NodeQueryData) Generate() interface{} {
	return struct {
		HostName string
		During   string
		GroupBy  string
	}{
		HostName: s.HostName,
		During:   s.During,
		GroupBy:  s.GroupBy,
	}
}

func (s *ClusterQueryData) Generate() interface{} {
	return struct {
		HostName string
		During   string
		GroupBy  string
	}{
		HostName: ".*",
		During:   s.During,
		GroupBy:  "",
	}
}
