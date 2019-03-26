package deployer

import (
	"context"

	"github.com/rancher/rancher/pkg/controllers/user/app"
	istiocommom "github.com/rancher/rancher/pkg/controllers/user/istioconfig/common"
	"github.com/rancher/rancher/pkg/controllers/user/systemimage"
	"github.com/rancher/rancher/pkg/project"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	serviceName = "istio"
)

type istioService struct {
	clusterName   string
	projectLister v3.ProjectLister
	appDeployer   *app.AppDeployer
}

func init() {
	systemimage.RegisterSystemService(serviceName, &istioService{})
}

func (l *istioService) Init(ctx context.Context, cluster *config.UserContext) {
	ad := app.NewAppDeployer(cluster)
	l.clusterName = cluster.ClusterName
	l.projectLister = cluster.Management.Management.Projects(metav1.NamespaceAll).Controller().Lister()
	l.appDeployer = ad
}

func (l *istioService) Version() (string, error) {
	return app.GetTemplateVersion(app.SystemCatalogName, istiocommom.TemplateName, istiocommom.AppInitVersion), nil
}

func (l *istioService) Upgrade(currentVersion string) (string, error) {
	appName := istiocommom.AppName
	templateName := app.GetTemplateName(app.SystemCatalogName, istiocommom.TemplateName)

	//upgrade old app
	systemProject, err := project.GetSystemProject(l.clusterName, l.projectLister)
	if err != nil {
		return "", err
	}

	return l.appDeployer.Upgrade(currentVersion, appName, templateName, systemProject.Name)
}
