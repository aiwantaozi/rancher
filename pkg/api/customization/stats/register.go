package stats

import (
	"net/http"

	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/pkg/clustermanager"
	clusterSchema "github.com/rancher/types/apis/cluster.cattle.io/v3/schema"
	projectschema "github.com/rancher/types/apis/project.cattle.io/v3/schema"
	"github.com/rancher/types/config"
)

const (
	clusterstats   = "clusterstats"
	hoststats      = "hoststats"
	workloadstats  = "workloadstats"
	podstats       = "podstats"
	containerstats = "containerstats"
)

func Register(apiContext *config.ScaledContext, clustermanager *clustermanager.Manager) {
	stats := Handler{
		clustermanager: clustermanager,
	}

	apiContext.Schemas.MustImport(&clusterSchema.Version, TimeSeries{}).
		MustImportAndCustomize(&clusterSchema.Version, ClusterStats{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet}
			schema.ListHandler = stats.Handler
			schema.PluralName = clusterstats
		}).
		MustImportAndCustomize(&clusterSchema.Version, HostStats{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet}
			schema.ListHandler = stats.Handler
			schema.PluralName = hoststats
		})

	apiContext.Schemas.MustImport(&projectschema.Version, TimeSeries{}).
		MustImportAndCustomize(&projectschema.Version, WorkloadStats{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet}
			schema.ListHandler = stats.Handler
			schema.PluralName = workloadstats
		}).
		MustImportAndCustomize(&projectschema.Version, PodStats{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet}
			schema.ListHandler = stats.Handler
			schema.PluralName = podstats
		}).
		MustImportAndCustomize(&projectschema.Version, ContainerStats{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet}
			schema.ListHandler = stats.Handler
			schema.PluralName = containerstats
		})
}
