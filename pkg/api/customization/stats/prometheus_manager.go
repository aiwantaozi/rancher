package stats

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rancher/types/config/dialer"
)

var legendFormat *regexp.Regexp

const (
	container = "container"
	pod       = "pod"
	workload  = "workload"
	cluster   = "cluster"
	node      = "node"
)

const (
	CPULoad1                             = "_cpu_load1"
	CPULoad5                             = "_cpu_load5"
	CPULoad15                            = "_cpu_load15"
	CPUUsageSecondsSumRate               = "_cpu_usage_seconds_sum_rate"
	CPUUserSecondsSumRate                = "_cpu_user_seconds_sum_rate"
	CPUSystemSecondsSumRate              = "_cpu_system_seconds_sum_rate"
	CPUCfsThrottledSecondsSumRate        = "_cpu_cfs_throttled_seconds_sum_rate"
	DiskIOReadsBytesSumRate              = "_disk_io_reads_bytes_sum_rate"
	DiskIOWritesBytesSumRate             = "_disk_io_writes_bytes_sum_rate"
	FsByteSum                            = "_fs_byte_sum"
	FsUsagePercent                       = "_fs_usage_percent"
	MemoryUsagePercent                   = "_memory_usage_percent"
	MemoryUsageBytesSum                  = "_memory_usage_bytes_sum"
	MemoryPageOutBytesSumRate            = "_memory_page_out_bytes_sum_rate"
	MemoryPageInBytesSumRate             = "_memory_page_in_bytes_sum_rate"
	NetworkReceiveBytesSumRate           = "_network_receive_bytes_sum_rate"
	NetworkReceivePacketsDroppedSumRate  = "_network_receive_packets_dropped_sum_rate"
	NetworkReceiveErrorsSumRate          = "_network_receive_errors_sum_rate"
	NetworkReceivePacketsSumRate         = "_network_receive_packets_sum_rate"
	NetworkTransmitBytesSumRate          = "_network_transmit_bytes_sum_rate"
	NetworkTransmitPacketsDroppedSumRate = "_network_transmit_packets_dropped_sum_rate"
	NetworkTransmitErrorsSumRate         = "_network_transmit_errors_sum_rate"
	NetworkTransmitPacketsSumRate        = "_network_transmit_packets_sum_rate"
)

type PrometheusManager struct {
	ctx     context.Context
	promAPI promapiv1.API
}

type Metric struct {
	PromManager PrometheusManager
}

type ClusterMetric struct {
	Metric
}

type NodeMetric struct {
	Metric
}

type WorkloadMetric struct {
	resourceType string
	Metric
}

type PodMetric struct {
	Metric
}

type ContainerMetric struct {
	Metric
}

type ResourceStats interface {
	GetCPU(*params) (TimeSeriesSlice, error)
	GetMemory(*params) (TimeSeriesSlice, error)
	GetNetwork(*params) (TimeSeriesSlice, error)
	GetDiskIO(*params) (TimeSeriesSlice, error)
	GetFileSystem(*params) (TimeSeriesSlice, error)
}

type Manager struct {
	ResourceStats ResourceStats
}

func init() {
	legendFormat = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)
}

func NewManager(ctx context.Context, dial dialer.Dialer, resourceType, url string) (*Manager, error) {
	cfg := promapi.Config{
		Address:      url,
		RoundTripper: newHTTPTransport(dial),
	}

	client, err := promapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create prometheus client failed: %v", err)
	}

	promManager := PrometheusManager{
		ctx:     ctx,
		promAPI: promapiv1.NewAPI(client),
	}

	resourceStats := getResourceStats(resourceType, promManager)
	return &Manager{
		ResourceStats: resourceStats,
	}, nil
}

func newHTTPTransport(dial dialer.Dialer) *http.Transport {
	return &http.Transport{
		Dial: dial,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
	}
}

func getResourceStats(resourceType string, promManager PrometheusManager) ResourceStats {
	switch resourceType {
	case cluster:
		return &ClusterMetric{
			Metric: Metric{
				PromManager: promManager,
			},
		}
	case node:
		return &NodeMetric{
			Metric: Metric{
				PromManager: promManager,
			},
		}
	case workload:
		return &WorkloadMetric{
			resourceType: workload,
			Metric: Metric{
				PromManager: promManager,
			},
		}
	case pod:
		return &WorkloadMetric{
			resourceType: pod,
			Metric: Metric{
				PromManager: promManager,
			},
		}
	case container:
		return &ContainerMetric{
			Metric: Metric{
				PromManager: promManager,
			},
		}
	}
	return nil
}

func (m *Manager) GetMetric(params *params) (*CommontMetric, error) {
	var tss []*TimeSeries
	cpus, err := m.ResourceStats.GetCPU(params)
	if err != nil {
		return nil, err
	}

	memorys, err := m.ResourceStats.GetMemory(params)
	if err != nil {
		return nil, err
	}

	diskIOs, err := m.ResourceStats.GetDiskIO(params)
	if err != nil {
		return nil, err
	}

	filesystem, err := m.ResourceStats.GetFileSystem(params)
	if err != nil {
		return nil, err
	}

	networks, err := m.ResourceStats.GetNetwork(params)
	if err != nil {
		return nil, err
	}
	tss = append(tss, cpus...)
	tss = append(tss, memorys...)
	tss = append(tss, diskIOs...)
	tss = append(tss, networks...)
	tss = append(tss, filesystem...)
	stats := &CommontMetric{
		ID:     params.nodeName,
		Series: tss,
	}
	return stats, nil
}

func (m *PrometheusManager) Query(name, exprs string, start, end time.Time, step time.Duration, result TimeSeriesSlice) (TimeSeriesSlice, error) {
	resp, err := m.query([]string{exprs}, start, end, step)
	if err != nil {
		return nil, fmt.Errorf("query %s from prometheus failed, %v", name, err)
	}

	if len(resp.Results) > 0 {
		for _, v1 := range resp.Results {
			for _, v2 := range v1.Series {
				v2.Name = name
				result = append(result, v2)
			}
		}
	} else {
		result = append(result, &TimeSeries{
			Name:   name,
			Points: [][2]float64{},
		})
	}

	return result, nil
}

func (m *PrometheusManager) query(exprs []string, start time.Time, end time.Time, step time.Duration) (*Response, error) {
	result := &Response{
		Results: []*QueryResult{},
	}

	querys, err := parseQuery(exprs, start, end, step)
	if err != nil {
		return nil, fmt.Errorf("parse prometheus query failed, %v", err)
	}

	for _, query := range querys {
		timeRange := promapiv1.Range{
			Start: query.Start,
			End:   query.End,
			Step:  query.Step,
		}

		value, err := m.promAPI.QueryRange(m.ctx, query.Expr, timeRange)
		if err != nil {
			return nil, fmt.Errorf("query range failed, %v", err)
		}

		queryResult, err := parseResponse(value, query)
		if err != nil {
			return nil, fmt.Errorf("parse prometheus query result failed, %v", err)
		}

		if queryResult != nil {
			result.Results = append(result.Results, queryResult)
		}
	}

	return result, nil
}

func newQueryResult() *QueryResult {
	return &QueryResult{
		Series: make(TimeSeriesSlice, 0),
	}
}

func parseQuery(exprs []string, start time.Time, end time.Time, step time.Duration) ([]*PrometheusQuery, error) {
	qs := []*PrometheusQuery{}
	for _, expr := range exprs {
		qs = append(qs, &PrometheusQuery{
			Expr:  expr,
			Step:  step,
			Start: start,
			End:   end,
		})
	}

	return qs, nil
}

func parseResponse(value model.Value, query *PrometheusQuery) (*QueryResult, error) {
	data, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("Unsupported result format: %s", value.Type().String())
	}

	if data.Len() == 0 {
		return nil, nil
	}

	queryRes := newQueryResult()
	for _, v := range data {
		series := TimeSeries{
			Name: formatLegend(v.Metric, query),
			Tags: map[string]string{},
		}

		for k, v := range v.Metric {
			series.Tags[string(k)] = string(v)
		}

		for _, k := range v.Values {
			series.Points = append(series.Points, NewTimePoint(float64(k.Value), float64(k.Timestamp.Unix()*1000)))
		}

		queryRes.Series = append(queryRes.Series, &series)
	}

	return queryRes, nil
}

func formatLegend(metric model.Metric, query *PrometheusQuery) string {
	if query.LegendFormat == "" {
		return metric.String()
	}

	result := legendFormat.ReplaceAllFunc([]byte(query.LegendFormat), func(in []byte) []byte {
		labelName := strings.Replace(string(in), "{{", "", 1)
		labelName = strings.Replace(labelName, "}}", "", 1)
		labelName = strings.TrimSpace(labelName)
		if val, exists := metric[model.LabelName(labelName)]; exists {
			return []byte(val)
		}

		return in
	})

	return string(result)
}

func (s *NodeMetric) GetCPU(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(node+CPUUsageSecondsSumRate, nodeCPUUsageSumRateExp(params.nodeName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+CPUSystemSecondsSumRate, nodeCPUSystemSecondsSumRateExp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+CPUUserSecondsSumRate, nodeCPUUserSecondsSumRateExp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+CPULoad1, nodeCPULoad1Exp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+CPULoad5, nodeCPULoad5Exp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(node+CPULoad15, nodeCPULoad15Exp(params.nodeName), params.start, params.end, params.step, result)
}

func (s *NodeMetric) GetMemory(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(node+MemoryUsagePercent, nodeMemoryUsagePercentExp(params.nodeName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+MemoryPageInBytesSumRate, nodeMemoryPageInSumRateExp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(node+MemoryPageOutBytesSumRate, nodeMemoryPageOutSumRateExp(params.nodeName), params.start, params.end, params.step, result)
}

func (s *NodeMetric) GetDiskIO(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(node+DiskIOReadsBytesSumRate, nodeDiskBytesReadSumRateExp(params.nodeName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(node+DiskIOWritesBytesSumRate, nodeDiskBytesWriteSumRateExp(params.nodeName), params.start, params.end, params.step, result)
}

func (s *NodeMetric) GetFileSystem(params *params) (TimeSeriesSlice, error) {
	return s.PromManager.Query(node+FsUsagePercent, nodeFilesystemUseagePercentExp(params.nodeName), params.start, params.end, params.step, TimeSeriesSlice{})
}

func (s *NodeMetric) GetNetwork(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(node+NetworkReceiveBytesSumRate, nodeNetworkReceiveBytesSumRateExp(params.nodeName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+NetworkReceiveErrorsSumRate, nodeNetworkReceiveErrorsSumRateExp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+NetworkReceivePacketsSumRate, nodeNetworkReceivePacketsSumRateExp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+NetworkReceivePacketsDroppedSumRate, nodeNetworkReceivePacketsDroppedSumRateExp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+NetworkTransmitBytesSumRate, nodeNetworkTransmittedBytesSumRateExp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+NetworkTransmitErrorsSumRate, nodeNetworkTransmittedErrorsSumRateExp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(node+NetworkTransmitPacketsSumRate, nodeNetworkTransmittedPacketsSumRateExp(params.nodeName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(node+NetworkTransmitPacketsDroppedSumRate, nodeNetworkTransmittedPacketsDroppedSumRateExp(params.nodeName), params.start, params.end, params.step, result)
}

func (s *ClusterMetric) GetCPU(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(cluster+CPUUsageSecondsSumRate, nodeCPUUsageSumRateExp(""), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+CPUSystemSecondsSumRate, nodeCPUSystemSecondsSumRateExp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+CPUUserSecondsSumRate, nodeCPUUserSecondsSumRateExp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+CPULoad1, nodeCPULoad1Exp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+CPULoad5, nodeCPULoad5Exp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(cluster+CPULoad15, nodeCPULoad15Exp(""), params.start, params.end, params.step, result)
}

func (s *ClusterMetric) GetMemory(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(cluster+MemoryUsagePercent, nodeMemoryUsagePercentExp(""), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+MemoryPageInBytesSumRate, nodeMemoryPageInSumRateExp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(cluster+MemoryPageOutBytesSumRate, nodeMemoryPageOutSumRateExp(""), params.start, params.end, params.step, result)
}

func (s *ClusterMetric) GetDiskIO(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(cluster+DiskIOReadsBytesSumRate, nodeDiskBytesReadSumRateExp(""), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(cluster+DiskIOWritesBytesSumRate, nodeDiskBytesWriteSumRateExp(""), params.start, params.end, params.step, result)
}

func (s *ClusterMetric) GetFileSystem(params *params) (TimeSeriesSlice, error) {
	return s.PromManager.Query(cluster+FsUsagePercent, nodeFilesystemUseagePercentExp(""), params.start, params.end, params.step, TimeSeriesSlice{})
}

func (s *ClusterMetric) GetNetwork(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(cluster+NetworkReceiveBytesSumRate, nodeNetworkReceiveBytesSumRateExp(""), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+NetworkReceiveErrorsSumRate, nodeNetworkReceiveErrorsSumRateExp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+NetworkReceivePacketsSumRate, nodeNetworkReceivePacketsSumRateExp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+NetworkReceivePacketsDroppedSumRate, nodeNetworkReceivePacketsDroppedSumRateExp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+NetworkTransmitBytesSumRate, nodeNetworkTransmittedBytesSumRateExp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+NetworkTransmitErrorsSumRate, nodeNetworkTransmittedErrorsSumRateExp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(cluster+NetworkTransmitPacketsSumRate, nodeNetworkTransmittedPacketsSumRateExp(""), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(cluster+NetworkTransmitPacketsDroppedSumRate, nodeNetworkTransmittedPacketsDroppedSumRateExp(""), params.start, params.end, params.step, result)
}

func (s *WorkloadMetric) GetCPU(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(s.resourceType+CPUUsageSecondsSumRate, workloadCPUUserSecondsSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(s.resourceType+CPUSystemSecondsSumRate, workloadCPUSystemSecondsSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(s.resourceType+CPUUserSecondsSumRate, workloadCPUUserSecondsSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(s.resourceType+CPUCfsThrottledSecondsSumRate, workloadCPUCfsThrottledSecondsSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
}

func (s *WorkloadMetric) GetMemory(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(s.resourceType+MemoryUsageBytesSum, workloadMemoryUsageBytesSumExp(params.namespace, params.podName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}
	return s.PromManager.Query(fmt.Sprintf("%s%s", s.resourceType, MemoryUsagePercent), workloadMemoryUsagePercentExp(params.namespace, params.podName), params.start, params.end, params.step, result)
}

func (s *WorkloadMetric) GetDiskIO(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(s.resourceType+DiskIOReadsBytesSumRate, workloadDiskBytesReadSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(s.resourceType+DiskIOWritesBytesSumRate, workloadDiskBytesWriteSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
}

func (s *WorkloadMetric) GetFileSystem(params *params) (TimeSeriesSlice, error) {
	return s.PromManager.Query(s.resourceType+FsByteSum, workloadFilesystemBytesSumExp(params.namespace, params.podName), params.start, params.end, params.step, TimeSeriesSlice{})
}

func (s *WorkloadMetric) GetNetwork(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(s.resourceType+NetworkReceiveBytesSumRate, workloadNetworkReceiveBytesSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(s.resourceType+NetworkReceiveErrorsSumRate, workloadNetworkReceiveErrorsSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(s.resourceType+NetworkReceivePacketsSumRate, workloadNetworkReceivePacketsSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(s.resourceType+NetworkReceivePacketsDroppedSumRate, workloadNetworkReceivePacketsDroppedSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(s.resourceType+NetworkTransmitBytesSumRate, workloadNetworkTransmittedBytesSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(s.resourceType+NetworkTransmitErrorsSumRate, workloadNetworkTransmittedErrorsSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(s.resourceType+NetworkTransmitPacketsSumRate, workloadNetworkTransmittedPacketsSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(s.resourceType+NetworkTransmitPacketsDroppedSumRate, workloadNetworkTransmittedPacketsDroppedSumRateExp(params.namespace, params.podName), params.start, params.end, params.step, result)
}

func (s *ContainerMetric) GetCPU(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(container+CPUUsageSecondsSumRate, containerCPUUsageSecondsSumRateExp(params.namespace, params.containerName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(container+CPUSystemSecondsSumRate, containerCPUSystemSecondsSumRateExp(params.namespace, params.containerName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	result, err = s.PromManager.Query(container+CPUUserSecondsSumRate, containerCPUUserSecondsSumRateExp(params.namespace, params.containerName), params.start, params.end, params.step, result)
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(container+CPUCfsThrottledSecondsSumRate, containerCPUCfsThrottledSecondsSumRateExp(params.namespace, params.containerName), params.start, params.end, params.step, result)
}

func (s *ContainerMetric) GetMemory(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(container+MemoryUsageBytesSum, containerMemoryUsageBytesSumExp(params.namespace, params.containerName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(container+MemoryUsagePercent, containerMemoryUsagePercentExp(params.namespace, params.containerName), params.start, params.end, params.step, result)
}

func (s *ContainerMetric) GetDiskIO(params *params) (TimeSeriesSlice, error) {
	result, err := s.PromManager.Query(container+DiskIOReadsBytesSumRate, containerDiskBytesReadSumRateExp(params.namespace, params.containerName), params.start, params.end, params.step, TimeSeriesSlice{})
	if err != nil {
		return nil, err
	}

	return s.PromManager.Query(container+DiskIOWritesBytesSumRate, containerDiskBytesWriteSumRateExp(params.namespace, params.containerName), params.start, params.end, params.step, result)
}

func (s *ContainerMetric) GetFileSystem(params *params) (TimeSeriesSlice, error) {
	return s.PromManager.Query(container+FsByteSum, containerFilesystemBytesSumExp(params.namespace, params.containerName), params.start, params.end, params.step, TimeSeriesSlice{})
}

func (s *ContainerMetric) GetNetwork(params *params) (TimeSeriesSlice, error) {
	return []*TimeSeries{}, nil
}

func nodeCPULoad1Exp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum(node_load1`)
	if host != "" {
		buffer.WriteString(`{instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"}`)
	}

	buffer.WriteString(`) / count(node_cpu{mode="system"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`})`)
	return buffer.String()
}

func nodeCPULoad5Exp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum(node_load5`)
	if host != "" {
		buffer.WriteString(`{instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"}`)
	}

	buffer.WriteString(`) / count(node_cpu{mode="system"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`})`)
	return buffer.String()
}

func nodeCPULoad15Exp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum(node_load15`)
	if host != "" {
		buffer.WriteString(`{instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"}`)
	}

	buffer.WriteString(`) / count(node_cpu{mode="system"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`})`)
	return buffer.String()
}

func nodeCPUUsageSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum(rate(node_cpu{mode!="idle", mode!="iowait", mode!~"^(?:guest.*)$"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeCPUUserSecondsSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum(rate(node_cpu{mode="user"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeCPUSystemSecondsSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum(rate(node_cpu{mode="system"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeMemoryUsagePercentExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`1 - sum(node_memory_MemAvailable`)
	if host != "" {
		buffer.WriteString(`{instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
		buffer.WriteString(`}`)
	}

	buffer.WriteString(`) / sum(node_memory_MemTotal`)
	if host != "" {
		buffer.WriteString(`{instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
		buffer.WriteString(`}`)
	}
	buffer.WriteString(")")
	return buffer.String()
}

func nodeMemoryPageInSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`1e3 * sum((rate(node_vmstat_pgpgin`)
	if host != "" {
		buffer.WriteString(`{instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"}`)
	}
	buffer.WriteString(`[2m])))`)
	return buffer.String()
}

func nodeMemoryPageOutSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`1e3 * sum((rate(node_vmstat_pgpgout`)
	if host != "" {
		buffer.WriteString(`{instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"}`)
	}
	buffer.WriteString(`[2m])))`)
	return buffer.String()
}

func nodeNetworkReceiveBytesSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_network_receive_bytes{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeNetworkReceivePacketsDroppedSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_network_receive_drop{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeNetworkReceiveErrorsSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_network_receive_errs{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeNetworkReceivePacketsSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_network_receive_packets{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeNetworkTransmittedBytesSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_network_transmit_bytes{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeNetworkTransmittedPacketsDroppedSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_network_transmit_drop{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeNetworkTransmittedErrorsSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_network_transmit_errs{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeNetworkTransmittedPacketsSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_network_transmit_packets{device!~"lo|veth.*|docker.*|flannel.*|cali.*|cbr.*"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}[2m]))`)
	return buffer.String()
}

func nodeFilesystemUseagePercentExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`(sum(node_filesystem_size{mountpoint="/"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}) - sum(node_filesystem_free{mountpoint="/"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`}))  / sum(node_filesystem_size{mountpoint="/"`)
	if host != "" {
		buffer.WriteString(`, instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"`)
	}
	buffer.WriteString(`})`)
	return buffer.String()
}

func nodeDiskBytesReadSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_disk_bytes_read`)
	if host != "" {
		buffer.WriteString(`{instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"}`)
	}
	buffer.WriteString(`[2m]))`)
	return buffer.String()
}

func nodeDiskBytesWriteSumRateExp(host string) string {
	var buffer bytes.Buffer
	buffer.WriteString(`sum ( rate(node_disk_bytes_written`)
	if host != "" {
		buffer.WriteString(`{instance=~"`)
		buffer.WriteString(host)
		buffer.WriteString(`.*"}`)
	}
	buffer.WriteString(`[2m]))`)
	return buffer.String()
}

func workloadCPUSecondsSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s",pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadCPUUserSecondsSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_cpu_user_seconds_total{namespace="%s",pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadCPUSystemSecondsSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_cpu_system_seconds_total{namespace="%s",pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadCPUCfsThrottledSecondsSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_cpu_cfs_throttled_seconds_total{namespace="%s",pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadMemoryUsageBytesSumExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(container_memory_working_set_bytes{name!~"POD", namespace="%s",pod_name=~"%s.*"})`, namespace, pod)
}

func workloadMemoryUsagePercentExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace="%s", pod_name=~"%s.*"}) / sum(label_join(kube_pod_container_resource_limits_memory_bytes{namespace="%s", pod=~"%s.*"},
		"pod_name", "", "pod"))`, namespace, pod, namespace, pod)
}

func workloadNetworkReceiveBytesSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_network_receive_bytes_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadNetworkReceivePacketsDroppedSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_network_receive_packets_dropped_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadNetworkReceiveErrorsSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_network_receive_errors_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadNetworkReceivePacketsSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_network_receive_packets_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadNetworkTransmittedBytesSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_network_transmit_bytes_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadNetworkTransmittedPacketsDroppedSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_network_transmit_packets_dropped_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadNetworkTransmittedErrorsSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_network_transmit_errors_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadNetworkTransmittedPacketsSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_network_transmit_packets_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadFilesystemBytesSumExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(container_fs_usage_bytes{namespace="%s", pod_name=~"%s.*"})`, namespace, pod)
}

func workloadDiskBytesReadSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_fs_reads_bytes_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func workloadDiskBytesWriteSumRateExp(namespace, pod string) string {
	return fmt.Sprintf(`sum(rate(container_fs_writes_bytes_total{namespace="%s", pod_name=~"%s.*"}[2m]))`, namespace, pod)
}

func containerCPUUsageSecondsSumRateExp(namespace, container string) string {
	return fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s",container_name=~"%s.*"}[2m]))`, namespace, container)
}

func containerCPUUserSecondsSumRateExp(namespace, container string) string {
	return fmt.Sprintf(`sum(rate(container_cpu_user_seconds_total{namespace="%s",container_name=~"%s.*"}[2m]))`, namespace, container)
}

func containerCPUSystemSecondsSumRateExp(namespace, container string) string {
	return fmt.Sprintf(`sum(rate(container_cpu_system_seconds_total{namespace="%s",container_name=~"%s.*"}[2m]))`, namespace, container)
}

func containerCPUCfsThrottledSecondsSumRateExp(namespace, container string) string {
	return fmt.Sprintf(`sum(rate(container_cpu_cfs_throttled_seconds_total{namespace="%s",container_name=~"%s.*"}[2m]))`, namespace, container)
}

func containerMemoryUsageBytesSumExp(namespace, container string) string {
	return fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace="%s",container_name=~"%s.*"})`, namespace, container)
}

func containerMemoryUsagePercentExp(namespace, container string) string {
	return fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace="%s", container_name=~"%s"}) / sum(label_join(kube_container_container_resource_limits_memory_bytes{namespace="%s",container=~"%s"},
		"container_name", "", "container"))`, namespace, container, namespace, container)
}

func containerFilesystemBytesSumExp(namespace, container string) string {
	return fmt.Sprintf(`sum(container_fs_usage_bytes{namespace="%s", container_name=~"%s"})`, namespace, container)
}

func containerDiskBytesReadSumRateExp(namespace, container string) string {
	return fmt.Sprintf(`sum(rate(container_fs_reads_bytes_total{namespace="%s", container_name=~"%s"}[2m]))`, namespace, container)
}

func containerDiskBytesWriteSumRateExp(namespace, container string) string {
	return fmt.Sprintf(`sum(rate(container_fs_writes_bytes_total{namespace="%s", container_name=~"%s"}[2m]))`, namespace, container)
}
