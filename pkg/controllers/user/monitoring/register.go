package monitoring

import (
	"context"

	"github.com/rancher/types/config"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Register initializes the controllers and registers
func Register(ctx context.Context, userContext *config.UserContext) error {
	clusterName := userContext.ClusterName

	logrus.Info("Registering cluster monitoring")

	mgmtContext := userContext.Management
	clustersClient := mgmtContext.Management.Clusters(metav1.NamespaceAll)
	clusterHandler := &clusterHandler{
		ctx:             ctx,
		clusterName:     clusterName,
		clustersClient:  clustersClient,
		workloadsClient: userContext.Apps,
		app: &appHandler{
			templateVersionClient: mgmtContext.Management.TemplateVersions(metav1.NamespaceAll),
			namespaceClient:       userContext.Core.Namespaces(metav1.NamespaceAll),
			appsGetter:            mgmtContext.Project,
			projectsGetter:        mgmtContext.Management,
		},
	}
	clustersClient.AddHandler("user-cluster-monitoring", clusterHandler.sync)

	return nil
}
