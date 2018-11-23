package stats

import (
	"github.com/rancher/norman/types"
)

func metricFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	resource.AddAction(apiContext, queryAction)
	resource.AddAction(apiContext, listMetricNameAction)
}

func metricCollectionFormatter(apiContext *types.APIContext, collection *types.GenericCollection) {
	collection.AddAction(apiContext, queryAction)
	collection.AddAction(apiContext, listMetricNameAction)
}
