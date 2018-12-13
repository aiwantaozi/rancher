package logging

import (
	"fmt"
	"net/http"

	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config/dialer"

	"github.com/rancher/rancher/pkg/controllers/user/logging/utils"
)

type Handler struct {
	DialerFactory dialer.Factory
}

func Formatter(apiContext *types.APIContext, resource *types.RawResource) {
	resource.AddAction(apiContext, "test")
}

func (h *Handler) ActionHandler(actionName string, action *types.Action, apiContext *types.APIContext) error {
	switch actionName {
	case "test":
		if err := h.testLoggingTarget(actionName, action, apiContext); err != nil {
			return httperror.NewAPIError(httperror.ServerError, err.Error())
		}

		apiContext.Response.WriteHeader(http.StatusNoContent)
	}

	return httperror.NewAPIError(httperror.InvalidAction, "invalid action: "+actionName)

}

func (h *Handler) testLoggingTarget(actionName string, action *types.Action, apiContext *types.APIContext) error {
	var input v3.TestInput
	actionInput, err := parse.ReadBody(apiContext.Request)
	if err != nil {
		return err
	}

	if err = convert.ToObj(actionInput, &input); err != nil {
		return err
	}

	clusterDialer, err := h.DialerFactory.ClusterDialer(input.ClusterName)
	if err != nil {
		return fmt.Errorf("get cluster dialer failed, %v", err)
	}

	wp := utils.NewWrapLogging(input.ElasticsearchConfig, input.SplunkConfig, input.SyslogConfig, input.KafkaConfig, input.FluentForwarderConfig)
	loggingTargt := wp.GetLoggingTarget()
	return loggingTargt.TestReachable(clusterDialer)
}
