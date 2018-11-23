package stats

import (
	"net/http"

	"github.com/rancher/norman/store/empty"
	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/pkg/clustermanager"
	clusterv3 "github.com/rancher/types/apis/cluster.cattle.io/v3"
	clusterSchema "github.com/rancher/types/apis/cluster.cattle.io/v3/schema"
	projectschema "github.com/rancher/types/apis/project.cattle.io/v3/schema"
	"github.com/rancher/types/config"
)

const (
	listMetricNameAction = "listmetricname"
	queryAction          = "query"
)

func Register(apiContext *config.ScaledContext, clustermanager *clustermanager.Manager) {
	metricHandler := newMetricHandler(apiContext.Dialer, clustermanager)
	apiContext.Schemas.MustImport(&clusterSchema.Version, TimeSeries{}).
		MustImport(&clusterSchema.Version, MetricNamesOutput{}).
		MustImport(&clusterSchema.Version, QueryMetricInput{}).
		MustImportAndCustomize(&clusterSchema.Version, clusterv3.MonitorMetric{}, func(schema *types.Schema) {
			schema.Formatter = metricFormatter
			schema.CollectionFormatter = metricCollectionFormatter
			schema.ActionHandler = metricHandler.action
			schema.CollectionActions = map[string]types.Action{
				queryAction: {
					Input:  "queryMetricInput",
					Output: "queryMetricOutput",
				},
				listMetricNameAction: {
					Output: "metricNamesOutput",
				},
			}
		})

	graphHandler := newGraphHandler(apiContext.Dialer, clustermanager)
	apiContext.Schemas.MustImport(&clusterSchema.Version, ResponseSeries{}).
		MustImport(&clusterSchema.Version, QueryGraphInput{}).
		MustImport(&clusterSchema.Version, QueryGraphOutput{}).
		MustImport(&clusterSchema.Version, QueryGraphOutputCollection{}).
		MustImportAndCustomize(&clusterSchema.Version, ClusterGraph{}, func(schema *types.Schema) {
			schema.Store = &empty.Store{}
			schema.PluralName = "clustergraphs"
			schema.Formatter = monitorGraphFormatter
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet}
			schema.ActionHandler = graphHandler.querySeriesAction
			schema.CollectionFormatter = queryGraphCollectionFormatter
			schema.ResourceActions = map[string]types.Action{
				queryAction: {
					Input:  "queryGraphInput",
					Output: "queryGraphOutput",
				},
			}
			schema.CollectionActions = map[string]types.Action{
				queryAction: {
					Input:  "queryGraphInput",
					Output: "queryGraphOutputCollection",
				},
			}

		})

	projectGraphHandler := newProjectGraphHandler(apiContext.Dialer, clustermanager)
	apiContext.Schemas.MustImport(&projectschema.Version, ResponseSeries{}).
		MustImport(&projectschema.Version, QueryGraphInput{}).
		MustImport(&projectschema.Version, QueryGraphOutput{}).
		MustImport(&projectschema.Version, QueryGraphOutputCollection{}).
		MustImportAndCustomize(&projectschema.Version, ProjectGraph{}, func(schema *types.Schema) {
			schema.Store = &empty.Store{}
			schema.PluralName = "projectgraphs"
			schema.Formatter = monitorGraphFormatter
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet}
			schema.ActionHandler = projectGraphHandler.querySeriesAction
			schema.CollectionFormatter = queryGraphCollectionFormatter
			schema.ResourceActions = map[string]types.Action{
				queryAction: {
					Input:  "queryGraphInput",
					Output: "queryGraphOutput",
				},
			}
			schema.CollectionActions = map[string]types.Action{
				queryAction: {
					Input:  "queryGraphInput",
					Output: "queryGraphOutputCollection",
				},
			}
		})
}
