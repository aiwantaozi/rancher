package logging

import (
	"net/http"
	"strings"

	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rancher/pkg/clustermanager"
	"github.com/rancher/rancher/pkg/controllers/user/logging/utils"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config/dialer"
)

func LoggingFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	resource.AddAction(apiContext, "test")
}

func LoggingCollectionFormatter(apiContext *types.APIContext, collection *types.GenericCollection) {
	collection.AddAction(apiContext, "test")
}

func NewHandler(clusterDialer dialer.Factory, clustermanager *clustermanager.Manager) handler {
	return handler{
		clusterDialer:     clusterDialer,
		clusterManagement: clustermanager,
	}
}

type handler struct {
	clusterDialer     dialer.Factory
	clusterManagement *clustermanager.Manager
}

func (h *handler) ActionHandler(actionName string, action *types.Action, apiContext *types.APIContext) error {
	clusterName := h.clusterManagement.ClusterName(apiContext)
	actionInput, err := parse.ReadBody(apiContext.Request)
	if err != nil {
		return err
	}

	var input v3.ClusterLoggingInput
	if err = convert.ToObj(actionInput, &input); err != nil {
		return err
	}

	arr := strings.Split(apiContext.ID, ":")
	clusterName = arr[0]

	err = utils.CheckClusterEndpoint(h.clusterDialer, input)
	if err != nil {
		return err
	}
	apiContext.WriteResponse(http.StatusNoContent, nil)
	return nil
}
