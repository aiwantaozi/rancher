package stats

import (
	statutil "github.com/rancher/rancher/pkg/stats"
	managementv3 "github.com/rancher/types/apis/management.cattle.io/v3"
)

type TemplateContentGenerator func(*params) interface{}

type ClusterMetric struct{}

func (s *ClusterMetric) generate(p *params) interface{} {
	queryData := statutil.NewMetricQueryData(managementv3.ResourceNode, statutil.ResourceQuery{
		HostName: ".*",
		During:   statutil.DefaultDuring,
	})
	return queryData.Generate()
}

func (s *ClusterMetric) GetCPU(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceCluster, p, []string{
		managementv3.CPUUsageSecondsSumRate,
		managementv3.CPUSystemSecondsSumRate,
		managementv3.CPUUserSecondsSumRate,
		managementv3.CPULoad1,
		managementv3.CPULoad5,
		managementv3.CPULoad15,
	}, s.generate)
}

func (s *ClusterMetric) GetMemory(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceCluster, p, []string{
		managementv3.MemoryUsagePercent,
		managementv3.MemoryPageInBytesSumRate,
		managementv3.MemoryPageOutBytesSumRate,
	}, s.generate)
}

func (s *ClusterMetric) GetDiskIO(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceCluster, p, []string{
		managementv3.DiskIOReadsBytesSumRate,
		managementv3.DiskIOWritesBytesSumRate,
	}, s.generate)
}

func (s *ClusterMetric) GetFileSystem(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceCluster, p, []string{
		managementv3.FilesystemUseagePercent,
	}, s.generate)
}

func (s *ClusterMetric) GetNetwork(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceCluster, p, []string{
		managementv3.NetworkReceiveBytesSumRate,
		managementv3.NetworkTransmitBytesSumRate,

		managementv3.NetworkReceiveErrorsSumRate,
		managementv3.NetworkTransmitErrorsSumRate,

		managementv3.NetworkReceivePacketsSumRate,
		managementv3.NetworkTransmitPacketsSumRate,

		managementv3.NetworkReceivePacketsDroppedSumRate,
		managementv3.NetworkTransmitPacketsDroppedSumRate,
	}, s.generate)
}

type NodeMetric struct{}

func (s *NodeMetric) generate(p *params) interface{} {
	queryData := statutil.NewMetricQueryData(managementv3.ResourceNode, statutil.ResourceQuery{
		HostName: p.nodeName + ".*",
		During:   statutil.DefaultDuring,
		GroupBy:  "instance",
	})
	return queryData.Generate()
}

func (s *NodeMetric) GetCPU(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceNode, p, []string{
		managementv3.CPUUsageSecondsSumRate,
		managementv3.CPUSystemSecondsSumRate,
		managementv3.CPUUserSecondsSumRate,
		managementv3.CPULoad1,
		managementv3.CPULoad5,
		managementv3.CPULoad15,
	}, s.generate)
}

func (s *NodeMetric) GetMemory(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceNode, p, []string{
		managementv3.MemoryUsagePercent,
		managementv3.MemoryPageInBytesSumRate,
		managementv3.MemoryPageOutBytesSumRate,
	}, s.generate)
}

func (s *NodeMetric) GetDiskIO(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceNode, p, []string{
		managementv3.DiskIOReadsBytesSumRate,
		managementv3.DiskIOWritesBytesSumRate,
	}, s.generate)
}

func (s *NodeMetric) GetFileSystem(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceNode, p, []string{
		managementv3.FilesystemUseagePercent,
	}, s.generate)
}

func (s *NodeMetric) GetNetwork(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceNode, p, []string{
		managementv3.NetworkReceiveBytesSumRate,
		managementv3.NetworkTransmitBytesSumRate,

		managementv3.NetworkReceiveErrorsSumRate,
		managementv3.NetworkTransmitErrorsSumRate,

		managementv3.NetworkReceivePacketsSumRate,
		managementv3.NetworkTransmitPacketsSumRate,

		managementv3.NetworkReceivePacketsDroppedSumRate,
		managementv3.NetworkTransmitPacketsDroppedSumRate,
	}, s.generate)
}

type WorkloadMetric struct {
	resourceType string
}

func (s *WorkloadMetric) generate(p *params) interface{} {
	queryData := statutil.NewMetricQueryData(managementv3.ResourcePod, statutil.ResourceQuery{
		Namespace: p.namespace,
		PodName:   p.podName + ".*",
		During:    statutil.DefaultDuring,
		GroupBy:   "pod_name",
	})
	return queryData.Generate()
}

func (s *WorkloadMetric) GetCPU(p *params) map[string]*PrometheusQuery {
	return generateQueries(s.resourceType, p, []string{
		managementv3.CPUUsageSecondsSumRate,
		managementv3.CPUSystemSecondsSumRate,
		managementv3.CPUUserSecondsSumRate,
		managementv3.CPUCfsThrottledSecondsSumRate,
	}, s.generate)
}

func (s *WorkloadMetric) GetMemory(p *params) map[string]*PrometheusQuery {
	return generateQueries(s.resourceType, p, []string{
		managementv3.MemoryUsageBytesSum,
		managementv3.MemoryUsagePercent,
	}, s.generate)
}

func (s *WorkloadMetric) GetDiskIO(p *params) map[string]*PrometheusQuery {
	return generateQueries(s.resourceType, p, []string{
		managementv3.DiskIOReadsBytesSumRate,
		managementv3.DiskIOWritesBytesSumRate,
	}, s.generate)
}

func (s *WorkloadMetric) GetFileSystem(p *params) map[string]*PrometheusQuery {
	return generateQueries(s.resourceType, p, []string{
		managementv3.FsByteSum,
	}, s.generate)
}

func (s *WorkloadMetric) GetNetwork(p *params) map[string]*PrometheusQuery {
	return generateQueries(s.resourceType, p, []string{
		managementv3.NetworkReceiveBytesSumRate,
		managementv3.NetworkTransmitBytesSumRate,

		managementv3.NetworkReceiveErrorsSumRate,
		managementv3.NetworkTransmitErrorsSumRate,

		managementv3.NetworkReceivePacketsSumRate,
		managementv3.NetworkTransmitPacketsSumRate,

		managementv3.NetworkReceivePacketsDroppedSumRate,
		managementv3.NetworkTransmitPacketsDroppedSumRate,
	}, s.generate)
}

type PodMetric struct {
	WorkloadMetric
}

type ContainerMetric struct{}

func (s *ContainerMetric) generate(p *params) interface{} {
	queryData := statutil.NewMetricQueryData(managementv3.ResourceContainer, statutil.ResourceQuery{
		Namespace:     p.namespace,
		ContainerName: p.containerName,
		During:        statutil.DefaultDuring,
		GroupBy:       "container_name",
	})
	return queryData.Generate()
}

func (s *ContainerMetric) GetCPU(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceContainer, p, []string{
		managementv3.CPUUsageSecondsSumRate,
		managementv3.CPUSystemSecondsSumRate,
		managementv3.CPUUserSecondsSumRate,
		managementv3.CPUCfsThrottledSecondsSumRate,
	}, s.generate)
}

func (s *ContainerMetric) GetMemory(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceContainer, p, []string{
		managementv3.MemoryUsageBytesSum,
		managementv3.MemoryUsagePercent,
	}, s.generate)
}

func (s *ContainerMetric) GetDiskIO(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceContainer, p, []string{
		managementv3.DiskIOReadsBytesSumRate,
		managementv3.DiskIOWritesBytesSumRate,
	}, s.generate)
}

func (s *ContainerMetric) GetFileSystem(p *params) map[string]*PrometheusQuery {
	return generateQueries(managementv3.ResourceContainer, p, []string{
		managementv3.FsByteSum,
	}, s.generate)
}

func (s *ContainerMetric) GetNetwork(params *params) map[string]*PrometheusQuery {
	return map[string]*PrometheusQuery{}
}

func generateQueries(t string, p *params, metrics []string, generator TemplateContentGenerator) map[string]*PrometheusQuery {
	rtn := map[string]*PrometheusQuery{}
	for _, k := range metrics {
		var v interface{} = struct {
			During string
		}{
			During: statutil.DefaultDuring,
		}
		if generator != nil {
			v = generator(p)
		}
		rtn[t+k] = InitPromQuery(p, statutil.GetExp(t, k, v), "")
	}
	return rtn
}
