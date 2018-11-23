package stats

import (
	"time"

	promapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type PrometheusQuery struct {
	ID              string
	Expr            string
	Step            time.Duration
	Start           time.Time
	End             time.Time
	LegendFormat    string
	ExtraAddedTags  map[string]string
	IsInstanceQuery bool
}

type TimeSeries struct {
	Name   string            `json:"name"`
	Points [][2]float64      `json:"points" norman:"type=array[array[float]]"`
	Tags   map[string]string `json:"tags"`
}

type TimeSeriesSlice []*TimeSeries

func (pq *PrometheusQuery) getRange() promapiv1.Range {
	return promapiv1.Range{
		Start: pq.Start,
		End:   pq.End,
		Step:  pq.Step,
	}
}
