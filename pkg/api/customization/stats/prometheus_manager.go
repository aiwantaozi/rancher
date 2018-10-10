package stats

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	managementv3 "github.com/rancher/types/apis/management.cattle.io/v3"

	promapi "github.com/prometheus/client_golang/api"
	promapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rancher/types/config/dialer"
	"golang.org/x/sync/errgroup"
)

var (
	legendFormat = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)
)

type ResourceStats interface {
	GetCPU(*params) map[string]*PrometheusQuery
	GetMemory(*params) map[string]*PrometheusQuery
	GetNetwork(*params) map[string]*PrometheusQuery
	GetDiskIO(*params) map[string]*PrometheusQuery
	GetFileSystem(*params) map[string]*PrometheusQuery
}

type Queries struct {
	ctx        context.Context
	api        promapiv1.API
	eg         *errgroup.Group
	queryNames []string
	queries    []*PrometheusQuery
}

func InitPromQuery(p *params, expr, format string) *PrometheusQuery {
	return &PrometheusQuery{
		Start:        p.start,
		End:          p.end,
		Step:         p.step,
		Expr:         expr,
		LegendFormat: format,
	}
}

func (q *Queries) Do(p *params) (TimeSeriesSlice, error) {
	states := getResourceStats(p.resourceType)
	q.appendQueries(states.GetCPU(p)).
		appendQueries(states.GetMemory(p)).
		appendQueries(states.GetNetwork(p)).
		appendQueries(states.GetDiskIO(p)).
		appendQueries(states.GetFileSystem(p))

	smap := &sync.Map{}
	for i := 0; i < len(q.queryNames); i++ {
		name := q.queryNames[i]
		query := q.queries[i]
		q.eg.Go(func() error {
			value, err := q.api.QueryRange(q.ctx, query.Expr, query.getRange())
			if err != nil {
				return fmt.Errorf("query range failed, %v, name: %s, expression: %s", err, name, query.Expr)
			}
			queryResult, err := parseResponse(value, query)
			if err != nil {
				return fmt.Errorf("parse prometheus query result failed, %v", err)
			}
			if queryResult != nil {
				smap.Store(name, queryResult.Series)
			}
			return nil
		})
	}
	if err := q.eg.Wait(); err != nil {
		return nil, err
	}
	rtn := TimeSeriesSlice{}
	smap.Range(func(k, v interface{}) bool {
		name := k.(string)
		series := v.(TimeSeriesSlice)
		for _, s := range series {
			s.Name = name
			rtn = append(rtn, s)
		}
		return true
	})
	return rtn, nil
}

func NewQuery(api promapiv1.API) *Queries {
	q := &Queries{
		ctx: context.Background(),
		api: api,
	}
	q.eg, q.ctx = errgroup.WithContext(q.ctx)
	return q
}

func (q *Queries) appendQueries(queryMap map[string]*PrometheusQuery) *Queries {
	for name, pq := range queryMap {
		q.queryNames = append(q.queryNames, name)
		q.queries = append(q.queries, pq)
	}
	return q
}

func getResourceStats(resourceType string) ResourceStats {
	switch resourceType {
	case managementv3.ResourceCluster:
		return &ClusterMetric{}
	case managementv3.ResourceNode:
		return &NodeMetric{}
	case managementv3.ResourceWorkload:
		return &WorkloadMetric{}
	case managementv3.ResourcePod:
		return &PodMetric{WorkloadMetric: WorkloadMetric{resourceType: managementv3.ResourcePod}}
	case managementv3.ResourceContainer:
		return &ContainerMetric{}
	}
	return nil
}

func NewPrometheusAPI(dial dialer.Dialer, url string) (promapiv1.API, error) {
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

func parseResponse(value model.Value, query *PrometheusQuery) (*QueryResult, error) {
	data, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("Unsupported result format: %s", value.Type().String())
	}

	if data.Len() == 0 {
		return nil, nil
	}

	queryRes := &QueryResult{
		Series: TimeSeriesSlice{},
	}
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

func newHTTPTransport(dial dialer.Dialer) *http.Transport {
	return &http.Transport{
		Dial: dial,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
	}
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
