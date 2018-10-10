package stats

type ClusterStats struct {
	CommontMetric
}

type ContainerStats struct {
	CommontMetric
}

type WorkloadStats struct {
	CommontMetric
}

type HostStats struct {
	CommontMetric
}

type PodStats struct {
	CommontMetric
}

type CommontMetric struct {
	ID     string        `json:"id,omitempty"`
	Series []*TimeSeries `json:"series" norman:"type=array[reference[timeseries]]"`
}
