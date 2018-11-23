package stats

import (
	"github.com/rancher/norman/types"
)

func monitorGraphFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	resource.AddAction(apiContext, queryAction)
}

func queryGraphCollectionFormatter(apiContext *types.APIContext, collection *types.GenericCollection) {
	collection.AddAction(apiContext, queryAction)
}
