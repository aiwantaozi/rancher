package app

import (
	"fmt"
	"reflect"

	"github.com/rancher/rancher/pkg/namespace"
	"github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	projectv3 "github.com/rancher/types/apis/project.cattle.io/v3"
	"github.com/rancher/types/config"

	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// This controller is responsible for deploy app
const (
	CreatorIDAnn      = "field.cattle.io/creatorId"
	SystemCatalogName = "system-library"
)

type AppDeployer struct {
	AppsGetter     projectv3.AppsGetter
	namespaces     v1.NamespaceInterface
	pods           v1.PodInterface
	templateLister v3.CatalogTemplateLister
}

func NewAppDeployer(cluster *config.UserContext) *AppDeployer {
	return &AppDeployer{
		AppsGetter:     cluster.Management.Project,
		namespaces:     cluster.Core.Namespaces(metav1.NamespaceAll),
		pods:           cluster.Core.Pods(metav1.NamespaceAll),
		templateLister: cluster.Management.Management.CatalogTemplates(namespace.GlobalNamespace).Controller().Lister(),
	}
}

func (d *AppDeployer) Deploy(app *projectv3.App) error {
	if err := d.initNamespace(app.Spec.TargetNamespace); err != nil {
		return err
	}

	current, err := d.AppsGetter.Apps(app.Namespace).GetNamespaced(app.Namespace, app.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			if _, err := d.AppsGetter.Apps(app.Namespace).Create(app); err != nil && !apierrors.IsAlreadyExists(err) {
				return errors.Wrapf(err, "create app %s:%s failed", app.Namespace, app.Name)
			}
			return nil
		}
		return errors.Wrapf(err, "get app %s:%s failed", app.Namespace, app.Name)
	}

	if current.DeletionTimestamp != nil {
		return errors.New("stale app " + app.Namespace + ":" + app.Name + " still on terminating")
	}

	if reflect.DeepEqual(current, app) {
		return nil
	}

	new := current.DeepCopy()
	new.Spec = app.Spec
	if _, err := d.AppsGetter.Apps(new.Namespace).Update(new); err != nil {
		return errors.Wrapf(err, "update app %s:%s failed", new.Namespace, new.Name)
	}

	return nil
}

func (d *AppDeployer) Cleanup(appName, projectName string) error {
	if err := d.AppsGetter.Apps(projectName).DeleteNamespaced(projectName, appName, &metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "delete app %s:%s failed", projectName, appName)
	}

	return nil
}

func (d *AppDeployer) IsPodDeploySuccess(targetNamespace string, selector map[string]string) error {
	opt := metav1.ListOptions{
		LabelSelector: labels.Set(selector).String(),
		FieldSelector: fields.Set{"metadata.namespace": targetNamespace}.String(),
	}
	pods, err := d.pods.List(opt)
	if err != nil {
		return errors.Wrapf(err, "'list pods in namespace %s with selector %+v failed", targetNamespace, selector)
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("couldn't find pods in namespace %s with selector %+v", targetNamespace, selector)
	}

	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case k8scorev1.PodFailed:
			return errors.New("get failed status from pod " + pod.Namespace + ":" + pod.Name + " , please the check logs")
		case k8scorev1.PodRunning, k8scorev1.PodSucceeded:
			return nil
		}
	}
	return nil
}

func (d *AppDeployer) IsAppDeploySuccess(appName, projectName string) error {
	app, err := d.AppsGetter.Apps(projectName).GetNamespaced(projectName, appName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "get app %s:%s failed", projectName, appName)
	}

	if !projectv3.AppConditionInstalled.IsTrue(app) {
		errMsg := fmt.Sprintf("app %s:%s installed condition is not true", projectName, appName)
		if condMsg := projectv3.AppConditionInstalled.GetMessage(app); condMsg != "" {
			return errors.New(errMsg + ", message: " + condMsg)
		}
	}

	return nil
}

func (d *AppDeployer) Upgrade(currentVersion, appName, templateName, projectName string) (string, error) {
	template, err := d.templateLister.Get(namespace.GlobalNamespace, templateName)
	if err != nil {
		return "", errors.Wrapf(err, "get template %s:%s failed", namespace.GlobalNamespace, templateName)
	}

	newFullVersion := fmt.Sprintf("%s-%s", templateName, template.Spec.DefaultVersion)
	if currentVersion == newFullVersion {
		return currentVersion, nil
	}

	//upgrade old app
	app, err := d.AppsGetter.Apps(projectName).GetNamespaced(projectName, appName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return newFullVersion, nil
		}
		return "", errors.Wrapf(err, "get app %s:%s failed", projectName, appName)
	}
	newApp := app.DeepCopy()
	newApp.Spec.ExternalID = GetExternalID(SystemCatalogName, templateName, template.Spec.DefaultVersion)

	if !reflect.DeepEqual(newApp, app) {
		if _, err = d.AppsGetter.Apps(projectName).Update(newApp); err != nil {
			return "", errors.Wrapf(err, "update app %s:%s failed", app.Namespace, app.Name)
		}
	}
	return newFullVersion, nil
}

func (d *AppDeployer) initNamespace(name string) error {
	if _, err := d.namespaces.Controller().Lister().Get(metav1.NamespaceAll, name); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		initNamespace := k8scorev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}

		if _, err := d.namespaces.Create(&initNamespace); err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func GetExternalID(catalogName, templateName, version string) string {
	return fmt.Sprintf("catalog://?catalog=%s&template=%s&version=%s", catalogName, templateName, version)
}

func GetTemplateVersion(catalogName, templateName, version string) string {
	return fmt.Sprintf("%s-%s-%s", catalogName, templateName, version)
}

func GetTemplateName(catalogName, templateName string) string {
	return fmt.Sprintf("%s-%s", catalogName, templateName)
}
