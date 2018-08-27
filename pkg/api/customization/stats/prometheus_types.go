package stats

import "time"

type PrometheusQuery struct {
	Expr         string
	Step         time.Duration
	Start        time.Time
	End          time.Time
	LegendFormat string
}

type Response struct {
	Results []*QueryResult `json:"results"`
	Message string         `json:"message,omitempty"`
}

type QueryResult struct {
	Error       error           `json:"-"`
	ErrorString string          `json:"error,omitempty"`
	Series      TimeSeriesSlice `json:"series"`
}

type TimeSeries struct {
	Name   string            `json:"name"`
	Points [][2]float64      `json:"points" norman:"type=array[array[float]]"`
	Tags   map[string]string `json:"-"`
}

type TimeSeriesSlice []*TimeSeries
