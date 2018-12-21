package logging

import (
	"net/http"

	loggingconfig "github.com/rancher/rancher/pkg/controllers/user/logging/config"
	"github.com/rancher/rancher/pkg/ref"

	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/parse"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rancher/pkg/controllers/user/logging/deployer"
	"github.com/rancher/rancher/pkg/controllers/user/logging/utils"
	"github.com/rancher/types/apis/core/v1"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	projectv3 "github.com/rancher/types/apis/project.cattle.io/v3"
	mgmtv3client "github.com/rancher/types/client/management/v3"
	"github.com/rancher/types/config/dialer"

	"github.com/pkg/errors"
)

type Handler struct {
	dialerFactory  dialer.Factory
	testerDeployer *deployer.TesterDeployer
}

func NewHandler(dialer dialer.Factory, appsGetter projectv3.AppsGetter, projectLister mgmtv3.ProjectLister, pods v1.PodInterface, projectLoggingLister mgmtv3.ProjectLoggingLister, namespaces v1.NamespaceInterface, templateLister mgmtv3.TemplateLister) *Handler {
	testerDeployer := deployer.NewTesterDeployer(appsGetter, projectLister, pods, projectLoggingLister, namespaces, templateLister)
	return &Handler{
		dialerFactory:  dialer,
		testerDeployer: testerDeployer,
	}
}

func Formatter(apiContext *types.APIContext, resource *types.RawResource) {
	resource.AddAction(apiContext, "test")
	resource.AddAction(apiContext, "dryRun")
}

func CollectionFormatter(apiContext *types.APIContext, resource *types.GenericCollection) {
	resource.AddAction(apiContext, "test")
	resource.AddAction(apiContext, "dryRun")
}

func (h *Handler) ActionHandler(actionName string, action *types.Action, apiContext *types.APIContext) error {
	var target mgmtv3.LoggingTargets
	var clusterName, projectName, level string

	switch apiContext.Type {
	case mgmtv3client.ClusterLoggingType:
		var input mgmtv3.ClusterTestInput
		actionInput, err := parse.ReadBody(apiContext.Request)
		if err != nil {
			return err
		}

		if err = convert.ToObj(actionInput, &input); err != nil {
			return err
		}

		target = input.LoggingTargets
		clusterName = input.ClusterName
		level = loggingconfig.ClusterLevel
	case mgmtv3client.ProjectLoggingType:
		var input mgmtv3.ProjectTestInput
		actionInput, err := parse.ReadBody(apiContext.Request)
		if err != nil {
			return err
		}

		if err = convert.ToObj(actionInput, &input); err != nil {
			return err
		}

		target = input.LoggingTargets
		projectName = input.ProjectName
		clusterName, _ = ref.Parse(input.ProjectName)
		level = loggingconfig.ProjectLevel
	}

	switch actionName {
	case "test":
		if err := h.testLoggingTarget(clusterName, target); err != nil {
			return httperror.NewAPIError(httperror.ServerError, err.Error())
		}

		apiContext.Response.WriteHeader(http.StatusNoContent)
	case "dryRun":

		if err := h.dryRunLoggingTarget(level, clusterName, projectName, target); err != nil {
			return httperror.NewAPIError(httperror.ServerError, err.Error())
		}

		apiContext.Response.WriteHeader(http.StatusNoContent)
	}

	return httperror.NewAPIError(httperror.InvalidAction, "invalid action: "+actionName)

}

func (h *Handler) testLoggingTarget(clusterName string, target mgmtv3.LoggingTargets) error {
	clusterDialer, err := h.dialerFactory.ClusterDialer(clusterName)
	if err != nil {
		return errors.Wrap(err, "get cluster dialer failed")
	}

	wp := utils.NewLoggingTargetTestWrap(target.ElasticsearchConfig, target.SplunkConfig, target.SyslogConfig, target.KafkaConfig, target.FluentForwarderConfig, target.CustomTargetConfig)
	if wp == nil {
		return nil
	}

	return wp.TestReachable(clusterDialer)
}

func (h *Handler) dryRunLoggingTarget(level, clusterName, projectID string, target mgmtv3.LoggingTargets) error {
	return h.testerDeployer.Deploy(level, clusterName, projectID, target)
}
