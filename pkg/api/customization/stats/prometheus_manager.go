package stats

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rancher/rancher/pkg/clustermanager"
	"github.com/rancher/types/config"
	"github.com/rancher/types/config/dialer"
	"golang.org/x/sync/errgroup"
)

var (
	legendFormat = regexp.MustCompile(`\[\[\s*(.+?)\s*\]\]`)
)

type Queries struct {
	ctx context.Context
	api promapiv1.API
	eg  *errgroup.Group
}

func InitPromQuery(id string, start, end time.Time, step time.Duration, expr, format string, extraTags map[string]string, isInstanceQuery bool) *PrometheusQuery {
	return &PrometheusQuery{
		ID:              id,
		Start:           start,
		End:             end,
		Step:            step,
		Expr:            expr,
		LegendFormat:    format,
		ExtraAddedTags:  extraTags,
		IsInstanceQuery: isInstanceQuery,
	}
}

func NewPrometheusQuery(userContext *config.UserContext, clusterName string, clustermanager *clustermanager.Manager, dialerFactory dialer.Factory) (*Queries, error) {
	dial, err := dialerFactory.ClusterDialer(clusterName)
	if err != nil {
		return nil, fmt.Errorf("get dail from usercontext failed, %v", err)
	}

	endpoint, err := getPrometheusEndpoint(userContext)
	if err != nil {
		return nil, err
	}

	api, err := newPrometheusAPI(dial, endpoint)
	if err != nil {
		return nil, err
	}
	return newQuery(api), nil
}

func (q *Queries) QueryRange(query *PrometheusQuery) (TimeResponseSeriesSlice, error) {
	fmt.Println("---expr:", query.Expr)
	value, err := q.api.QueryRange(q.ctx, query.Expr, query.getRange())
	if err != nil {
		return nil, fmt.Errorf("query range failed, %v, expression: %s", err, query.Expr)
	}
	seriesSlice, err := parseMatrix(value, query)
	if err != nil {
		return nil, fmt.Errorf("parse prometheus query result failed, %v", err)
	}
	return seriesSlice, nil
}

func (q *Queries) Query(query *PrometheusQuery) (TimeResponseSeriesSlice, error) {
	fmt.Println("---expr:", query.Expr)

	value, err := q.api.Query(q.ctx, query.Expr, time.Now())
	if err != nil {
		return nil, fmt.Errorf("query range failed, %v, expression: %s", err, query.Expr)
	}
	seriesSlice, err := parseVector(value, query)
	if err != nil {
		return nil, fmt.Errorf("parse prometheus query result failed, %v", err)
	}
	return seriesSlice, nil
}

func (q *Queries) Do(querys []*PrometheusQuery) (TimeResponseSeriesSlice, error) {
	smap := &sync.Map{}
	for _, v := range querys {
		query := v
		q.eg.Go(func() error {
			var seriesSlice TimeResponseSeriesSlice
			var err error
			if query.IsInstanceQuery {
				seriesSlice, err = q.Query(query)
			} else {
				seriesSlice, err = q.QueryRange(query)
			}
			if err != nil {
				return err
			}

			if seriesSlice != nil {
				smap.Store(query.ID, seriesSlice)
			}
			return nil
		})
	}
	if err := q.eg.Wait(); err != nil {
		return nil, err
	}
	rtn := TimeResponseSeriesSlice{}
	smap.Range(func(k, v interface{}) bool {
		series := v.(TimeResponseSeriesSlice)
		for _, s := range series {
			rtn = append(rtn, s)
		}
		return true
	})

	return rtn, nil
}

func (q *Queries) GetLabelValues(labelName string) ([]string, error) {
	value, err := q.api.LabelValues(q.ctx, labelName)
	if err != nil {
		return nil, fmt.Errorf("get prometheus metric list failed, %v", err)
	}

	var metricNames []string
	for _, v := range value {
		metricNames = append(metricNames, fmt.Sprint(v))
	}
	return metricNames, nil
}

func newQuery(api promapiv1.API) *Queries {
	q := &Queries{
		ctx: context.Background(),
		api: api,
	}
	q.eg, q.ctx = errgroup.WithContext(q.ctx)
	return q
}

func newPrometheusAPI(dial dialer.Dialer, url string) (promapiv1.API, error) {
	cfg := promapi.Config{
		Address:      url,
		RoundTripper: newHTTPTransport(dial),
	}

	client, err := promapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create prometheus client failed: %v", err)
	}
	return promapiv1.NewAPI(client), nil
}

func newHTTPTransport(dial dialer.Dialer) *http.Transport {
	return &http.Transport{
		Dial:                  dial,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
	}
}

func parseVector(value model.Value, query *PrometheusQuery) (TimeResponseSeriesSlice, error) {
	data, ok := value.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("Unsupported result format: %s", value.Type().String())
	}

	if data.Len() == 0 {
		return nil, nil
	}

	vec := data[0]
	series := ResponseSeries{
		ID:         query.ID,
		Expression: query.Expr,
		TimeSeries: TimeSeries{
			Name: formatLegend(vec.Metric, query),
			Tags: map[string]string{},
		},
	}

	for k, v := range vec.Metric {
		series.Tags[string(k)] = string(v)
	}

	po, isValid := NewTimePoint(float64(vec.Value), float64(vec.Timestamp.Unix()*1000))
	if isValid {
		series.Points = append(series.Points, po)
	}

	return TimeResponseSeriesSlice{&series}, nil
}

func parseMatrix(value model.Value, query *PrometheusQuery) (TimeResponseSeriesSlice, error) {
	data, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("Unsupported result format: %s", value.Type().String())
	}

	if data.Len() == 0 {
		return nil, nil
	}

	var seriesSlice TimeResponseSeriesSlice
	for _, v := range data {
		series := ResponseSeries{
			ID:         query.ID,
			Expression: query.Expr,
			TimeSeries: TimeSeries{
				Name: formatLegend(v.Metric, query),
				Tags: map[string]string{},
			},
		}

		for k, v := range v.Metric {
			series.Tags[string(k)] = string(v)
		}

		for k, v := range query.ExtraAddedTags {
			series.Tags[k] = v
		}

		for _, k := range v.Values {
			po, isValid := NewTimePoint(float64(k.Value), float64(k.Timestamp.Unix()*1000))
			if isValid {
				series.Points = append(series.Points, po)
			}

		}

		seriesSlice = append(seriesSlice, &series)
	}

	return seriesSlice, nil
}

func formatLegend(metric model.Metric, query *PrometheusQuery) string {
	if query.LegendFormat == "" {
		return metric.String()
	}

	result := legendFormat.ReplaceAllFunc([]byte(query.LegendFormat), func(in []byte) []byte {
		labelName := strings.Replace(string(in), "[[", "", 1)
		labelName = strings.Replace(labelName, "]]", "", 1)
		labelName = strings.TrimSpace(labelName)
		if val, exists := metric[model.LabelName(labelName)]; exists {
			return []byte(val)
		}

		return in
	})

	return string(result)
}
