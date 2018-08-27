package stats

import (
	"testing"
)

func TestNodeDiskWrite(t *testing.T) {
	str := nodeCPULoad1Exp("$node")
	t.Log(str)
	str = nodeCPULoad5Exp("$node")
	t.Log(str)
	str = nodeCPULoad15Exp("$node")
	t.Log(str)
	str = nodeCPUUsageSumRateExp("$node")
	t.Log(str)
	str = nodeCPUUserSecondsSumRateExp("$node")
	t.Log(str)
	str = nodeCPUSystemSecondsSumRateExp("$node")
	t.Log(str)
	str = nodeMemoryUsagePercentExp("")
	t.Log(str)
	// 	str = nodeMemoryPageInSumRateExp("$node")
	// 	t.Log(str)
	// 	str = nodeMemoryPageOutSumRateExp("$node")
	// 	t.Log(str)
	// 	str = nodeFilesystemUseagePercentExp("$node")
	// 	t.Log(str)
	// 	str = nodeDiskBytesReadSumRateExp("$node")
	// 	t.Log(str)
	// 	str = nodeDiskBytesWriteSumRateExp("$node")
	// 	t.Log(str)

	// 	str = nodeNetworkReceiveBytesSumRateExp("$node")
	// 	t.Log(str)
	// 	str = nodeNetworkReceiveErrorsSumRateExp("$node")
	// 	t.Log(str)
	// 	str = nodeNetworkReceivePacketsDroppedSumRateExp("$node")
	// 	t.Log(str)
	// 	str = nodeNetworkReceivePacketsSumRateExp("$node")
	// 	t.Log(str)

	// 	str = nodeNetworkTransmittedBytesSumRateExp("$node")
	// 	t.Log(str)
	// 	str = nodeNetworkTransmittedErrorsSumRateExp("$node")
	// 	t.Log(str)
	// 	str = nodeNetworkTransmittedPacketsDroppedSumRateExp("$node")
	// 	t.Log(str)
	// 	str = nodeNetworkTransmittedPacketsSumRateExp("$node")
	// 	t.Log(str)
	// }

	// func TestWorkloadDiskWrite(t *testing.T) {
	// 	str := workloadCPUSecondsSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadCPUUserSecondsSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadCPUSystemSecondsSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadMemoryUsagePercentExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadDiskBytesReadSumRateExp("$namespace", "$pod")
	// 	t.Log(str)

	// 	str = workloadDiskBytesWriteSumRateExp("$namespace", "$pod")
	// 	t.Log(str)

	// 	str = workloadNetworkReceiveBytesSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadNetworkReceiveErrorsSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadNetworkReceivePacketsDroppedSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadNetworkReceivePacketsSumRateExp("$namespace", "$pod")
	// 	t.Log(str)

	// 	str = workloadNetworkTransmittedBytesSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadNetworkTransmittedErrorsSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadNetworkTransmittedPacketsDroppedSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
	// 	str = workloadNetworkTransmittedPacketsSumRateExp("$namespace", "$pod")
	// 	t.Log(str)
}

// func TestContainerQueryCPUUsageSecondsTotal(t *testing.T) {
// 	str = containerQueryCPUUsageSecondsTotal("mypod", "myns")
// 	t.Log(str)
// }

// func TestnodeDiskWriteExp(t *testing.T) {
// 	str = nodeDiskWriteExp("mynode")
// 	t.Log(str)
// }

// func TestnodeCPUUsageExp(t *testing.T) {
// 	url = "http://167.99.232.247:31515"
// 	dialer = net.Dialer{
// 		Timeout:   10 * time.Second,
// 		KeepAlive: 30 * time.Second,
// 	}
// 	manager, err = NewManager(dialer.Dial, url)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	timeRange = NewTimeRange("now-1m", defaultTo)

// 	start, _ = timeRange.ParseFrom()

// 	end, _ = timeRange.ParseTo()

// 	step = time.Duration(int64(defaultMinInterval * 1000))

// 	ctx = context.Background()
// 	expr = containerCPUCfsThrottledSecondsTotalExp("myns", "mycontainer")
// 	resp, err = manager.query(ctx, []string{expr}, start, end, step)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	for _, v = range resp.Results {
// 		fmt.Printf("---%v", v)
// 	}

// }

// func TestnodeQueryExps(t *testing.T) {
// 	metric = nodeQueryExps("minikube")
// 	t.Log(metric)
// }

// func TestPrometheusClient(t *testing.T) {
// 	ctx = context.Background()
// 	cfg = promapi.Config{
// 		Address:      "http://167.99.232.247:31515",
// 		RoundTripper: promapi.DefaultRoundTripper,
// 	}
// 	client, err = promapi.NewClient(cfg)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	api = promapiv1.NewAPI(client)
// 	duration, err = time.ParseDuration("-1h")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	step, err = time.ParseDuration("5m")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	timeRange = promapiv1.Range{
// 		Start: time.Now().Add(duration),
// 		End:   time.Now(),
// 		Step:  step,
// 	}

// 	expr = containerCPUCfsThrottledSecondsTotalExp("myns", "mycontainer")
// 	// query = nodeNetworkTransmitted("192.168.99.100:9100")
// 	// t.Log(query)

// 	querys, err = parseQuery([]string{expr}, start, end, step)
// 	if err != nil {
// 		return nil, fmt.Errorf("parse prometheus query failed, %v", err)
// 	}

// 	value, err = api.QueryRange(ctx, expr, timeRange)
// 	parseResponse(value)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	t.Log("-------------here:")
// 	t.Log(value)
// }
