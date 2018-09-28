package v3

const (
	ResourceContainer = "container"
	ResourcePod       = "pod"
	ResourceWorkload  = "workload"
	ResourceCluster   = "cluster"
	ResourceNode      = "node"
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
	FilesystemUseagePercent              = "_fs_usage_percent"
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

var (
	MetricEnum = []string{
		ResourceCluster + CPULoad1,
		ResourceCluster + CPULoad5,
		ResourceCluster + CPULoad15,
		ResourceCluster + CPUUsageSecondsSumRate,
		ResourceCluster + CPUUserSecondsSumRate,
		ResourceCluster + CPUSystemSecondsSumRate,
		ResourceCluster + MemoryUsagePercent,
		ResourceCluster + MemoryPageInBytesSumRate,
		ResourceCluster + MemoryPageOutBytesSumRate,
		ResourceCluster + NetworkReceiveBytesSumRate,
		ResourceCluster + NetworkReceivePacketsDroppedSumRate,
		ResourceCluster + NetworkReceiveErrorsSumRate,
		ResourceCluster + NetworkReceivePacketsSumRate,
		ResourceCluster + NetworkTransmitBytesSumRate,
		ResourceCluster + NetworkTransmitPacketsDroppedSumRate,
		ResourceCluster + NetworkTransmitErrorsSumRate,
		ResourceCluster + FilesystemUseagePercent,
		ResourceCluster + DiskIOReadsBytesSumRate,
		ResourceCluster + DiskIOWritesBytesSumRate,

		ResourceNode + CPULoad1,
		ResourceNode + CPULoad5,
		ResourceNode + CPULoad15,
		ResourceNode + CPUUsageSecondsSumRate,
		ResourceNode + CPUUserSecondsSumRate,
		ResourceNode + CPUSystemSecondsSumRate,
		ResourceNode + MemoryUsagePercent,
		ResourceNode + MemoryPageInBytesSumRate,
		ResourceNode + MemoryPageOutBytesSumRate,
		ResourceNode + NetworkReceiveBytesSumRate,
		ResourceNode + NetworkReceivePacketsDroppedSumRate,
		ResourceNode + NetworkReceiveErrorsSumRate,
		ResourceNode + NetworkReceivePacketsSumRate,
		ResourceNode + NetworkTransmitBytesSumRate,
		ResourceNode + NetworkTransmitPacketsDroppedSumRate,
		ResourceNode + NetworkTransmitErrorsSumRate,
		ResourceNode + FilesystemUseagePercent,
		ResourceNode + DiskIOReadsBytesSumRate,
		ResourceNode + DiskIOWritesBytesSumRate,

		ResourcePod + CPUUsageSecondsSumRate,
		ResourcePod + CPUUserSecondsSumRate,
		ResourcePod + CPUSystemSecondsSumRate,
		ResourcePod + CPUCfsThrottledSecondsSumRate,
		ResourcePod + MemoryUsageBytesSum,
		ResourcePod + MemoryUsagePercent,
		ResourcePod + DiskIOReadsBytesSumRate,
		ResourcePod + DiskIOWritesBytesSumRate,
		ResourcePod + FsByteSum,
		ResourcePod + NetworkReceiveBytesSumRate,
		ResourcePod + NetworkTransmitBytesSumRate,
		ResourcePod + NetworkReceiveErrorsSumRate,
		ResourcePod + NetworkTransmitErrorsSumRate,
		ResourcePod + NetworkReceivePacketsSumRate,
		ResourcePod + NetworkTransmitPacketsSumRate,
		ResourcePod + NetworkReceivePacketsDroppedSumRate,
		ResourcePod + NetworkTransmitPacketsDroppedSumRate,
		ResourceContainer + CPUUsageSecondsSumRate,
		ResourceContainer + CPUUserSecondsSumRate,
		ResourceContainer + CPUSystemSecondsSumRate,
		ResourceContainer + CPUCfsThrottledSecondsSumRate,
		ResourceContainer + MemoryUsageBytesSum,
		ResourceContainer + MemoryUsagePercent,
		ResourceContainer + DiskIOReadsBytesSumRate,
		ResourceContainer + DiskIOWritesBytesSumRate,
		ResourceContainer + FsByteSum,
	}
)
