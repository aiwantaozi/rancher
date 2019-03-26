package deployer

import (
	"fmt"
	"reflect"

	"github.com/rancher/rancher/pkg/controllers/user/app"
	istiocommon "github.com/rancher/rancher/pkg/controllers/user/istioconfig/common"
	"github.com/rancher/rancher/pkg/namespace"
	"github.com/rancher/rancher/pkg/project"
	"github.com/rancher/rancher/pkg/ref"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	projectv3 "github.com/rancher/types/apis/project.cattle.io/v3"
	"github.com/rancher/types/config"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Deployer struct {
	clusterName    string
	istioConfigs   mgmtv3.IstioConfigInterface
	projectLister  mgmtv3.ProjectLister
	templateLister mgmtv3.CatalogTemplateLister
	appDeployer    *app.AppDeployer
}

func NewDeployer(cluster *config.UserContext) *Deployer {
	clusterName := cluster.ClusterName
	appDeployer := app.NewAppDeployer(cluster)

	return &Deployer{
		clusterName:    clusterName,
		istioConfigs:   cluster.Management.Management.IstioConfigs(clusterName),
		projectLister:  cluster.Management.Management.Projects(metav1.NamespaceAll).Controller().Lister(),
		templateLister: cluster.Management.Management.CatalogTemplates(metav1.NamespaceAll).Controller().Lister(),
		appDeployer:    appDeployer,
	}
}

func (d *Deployer) IstioConfigSync(key string, obj *mgmtv3.IstioConfig) (runtime.Object, error) {
	return obj, d.sync(obj)
}

func (d *Deployer) sync(obj *mgmtv3.IstioConfig) error {
	systemProject, err := project.GetSystemProject(d.clusterName, d.projectLister)
	if err != nil {
		return err
	}

	systemProjectCreator := systemProject.Annotations[app.CreatorIDAnn]
	systemProjectID := ref.Ref(systemProject)
	systemProjectName := systemProject.Name

	if obj == nil || obj.DeletionTimestamp != nil || obj.Spec.Enable == false {
		return d.cleanup(systemProjectName)
	}

	if reflect.DeepEqual(obj.Spec, obj.Status.AppliedSpec) {
		return nil
	}

	istioConfigObj := obj.DeepCopy()

	if istioConfigObj != nil && istioConfigObj.Spec.Enable {
		if err := d.deploy(systemProjectID, systemProjectCreator, istioConfigObj); err != nil {
			return err
		}

		mgmtv3.IstioConditionAppInstalled.True(istioConfigObj)
	}

	istioConfigs, err := d.istioConfigs.List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "list istioConfig %s:%s failed", d.clusterName, d.clusterName)
	}

	if len(istioConfigs.Items) != 1 {
		return errors.New("get " + fmt.Sprint(len(istioConfigs.Items)) + " istioConfigs, not equal to 1")
	}

	istioConfig := istioConfigs.Items[0]
	cpIstioConfig := istioConfig.DeepCopy()

	if !reflect.DeepEqual(istioConfigObj, cpIstioConfig) {
		cpIstioConfig.Spec = istioConfigObj.Spec
		cpIstioConfig.Status = istioConfigObj.Status
		if _, err = d.istioConfigs.Update(cpIstioConfig); err != nil {
			return errors.Wrapf(err, "update istioConfig %s:%s failed", cpIstioConfig.Namespace, cpIstioConfig.Name)
		}
	}

	return nil
}

func (d *Deployer) deploy(systemProjectID, systemProjectCreator string, obj *mgmtv3.IstioConfig) error {
	return d.deployIstio(systemProjectID, systemProjectCreator, obj.Spec.AppAnswers)
}

func (d *Deployer) cleanup(projectName string) error {
	appName := istiocommon.AppName
	return d.appDeployer.Cleanup(appName, projectName)
}

func (d *Deployer) deployIstio(systemProjectID, systemProjectCreator string, answers map[string]string) error {
	templateName := app.GetTemplateName(app.SystemCatalogName, istiocommon.TemplateName)
	template, err := d.templateLister.Get(namespace.GlobalNamespace, templateName)
	if err != nil {
		return errors.Wrapf(err, "get template %s:%s failed", namespace.GlobalNamespace, templateName)
	}

	externalID := app.GetExternalID(app.SystemCatalogName, istiocommon.TemplateName, template.Spec.DefaultVersion)
	app := istioApp(systemProjectCreator, systemProjectID, externalID, answers)
	return d.appDeployer.Deploy(app)
}

func istioApp(systemProjectCreator, projectID, externalID string, answers map[string]string) *projectv3.App {
	appName := istiocommon.AppName
	namepspace := istiocommon.IstioDeployedNamespace
	_, projectName := ref.Parse(projectID)

	return &projectv3.App{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				app.CreatorIDAnn: systemProjectCreator,
			},
			Name:      appName,
			Namespace: projectName,
		},
		Spec: projectv3.AppSpec{
			Answers:         answers,
			Description:     "Istio for connect, secure, control, and observe services.",
			ExternalID:      externalID,
			ProjectName:     projectID,
			TargetNamespace: namepspace,
		},
	}
}
