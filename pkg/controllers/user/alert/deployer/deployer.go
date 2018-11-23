package deployer

import (
	"github.com/pkg/errors"
	"github.com/rancher/norman/controller"
	"github.com/rancher/rancher/pkg/controllers/user/alert/manager"
	"github.com/rancher/rancher/pkg/image"
	monitorutil "github.com/rancher/rancher/pkg/monitoring"
	"github.com/rancher/types/apis/apps/v1beta2"
	"github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	appName    = "alertmanager"
	appSvcName = "alertmanager-svc"
)

func NewDeployer(cluster *config.UserContext, manager *manager.AlertManager) *Deployer {
	return &Deployer{
		nsClient:                cluster.Core.Namespaces(""),
		appsClient:              cluster.Apps.Deployments(""),
		secretClient:            cluster.Core.Secrets(""),
		svcClient:               cluster.Core.Services(""),
		clusterAlertGroupLister: cluster.Management.Management.ClusterAlertGroups(cluster.ClusterName).Controller().Lister(),
		projectAlertGroupLister: cluster.Management.Management.ProjectAlertGroups("").Controller().Lister(),
		notifierLister:          cluster.Management.Management.Notifiers(cluster.ClusterName).Controller().Lister(),
		alertManager:            manager,
		clusterName:             cluster.ClusterName,
	}
}

type Deployer struct {
	nsClient                v1.NamespaceInterface
	appsClient              v1beta2.DeploymentInterface
	secretClient            v1.SecretInterface
	svcClient               v1.ServiceInterface
	projectAlertGroupLister v3.ProjectAlertGroupLister
	clusterAlertGroupLister v3.ClusterAlertGroupLister
	notifierLister          v3.NotifierLister
	alertManager            *manager.AlertManager
	clusterName             string
}

func (d *Deployer) ProjectSync(key string, alert *v3.ProjectAlertGroup) (runtime.Object, error) {
	return nil, d.sync()
}

func (d *Deployer) ClusterSync(key string, alert *v3.ClusterAlertGroup) (runtime.Object, error) {
	return nil, d.sync()
}

//deploy or clean up resources(alertmanager deployment, service, namespace) required by alerting.
func (d *Deployer) sync() error {
	needDeploy, err := d.needDeploy()
	if err != nil {
		return errors.Wrapf(err, "Check alertmanager deployment")
	}

	if needDeploy {
		return d.deploy()
	}

	return d.cleanup()
}

//only deploy the alertmanager when notifier is configured and alert is using it.
func (d *Deployer) needDeploy() (bool, error) {
	notifiers, err := d.notifierLister.List("", labels.NewSelector())
	if err != nil {
		return false, err
	}

	if len(notifiers) == 0 {
		return false, err
	}

	clusterAlerts, err := d.clusterAlertGroupLister.List("", labels.NewSelector())
	if err != nil {
		return false, err
	}

	for _, alert := range clusterAlerts {
		if len(alert.Spec.Recipients) > 0 {
			return true, nil
		}
	}

	projectAlerts, err := d.projectAlertGroupLister.List("", labels.NewSelector())
	if err != nil {
		return false, nil
	}

	for _, alert := range projectAlerts {
		if controller.ObjectInCluster(d.clusterName, alert) {
			if len(alert.Spec.Recipients) > 0 {
				return true, nil
			}
		}
	}

	return false, nil

}

func (d *Deployer) cleanup() error {
	deleteOp := metav1.DeletePropagationBackground
	var errgrp errgroup.Group

	errgrp.Go(func() error {
		return d.svcClient.DeleteNamespaced(monitorutil.CattleNamespaceName, appName, &metav1.DeleteOptions{})
	})

	errgrp.Go(func() error {
		return d.appsClient.DeleteNamespaced(monitorutil.CattleNamespaceName, appSvcName, &metav1.DeleteOptions{})
	})

	errgrp.Go(func() error {
		return d.secretClient.DeleteNamespaced(monitorutil.CattleNamespaceName, appName, &metav1.DeleteOptions{PropagationPolicy: &deleteOp})
	})

	if err := errgrp.Wait(); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	d.alertManager.IsDeploy = false
	return nil
}

func (d *Deployer) deploy() error {

	//TODO: cleanup resources while there is not any alert configured
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: monitorutil.CattleNamespaceName,
		},
	}
	if _, err := d.nsClient.Create(ns); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "Creating ns")
	}

	secret := d.getSecret()
	if _, err := d.secretClient.Create(secret); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "Creating secret")
	}

	deployment := GetDeployment()
	if _, err := d.appsClient.Create(deployment); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "Creating deployment")
	}

	service := getService()
	if _, err := d.svcClient.Create(service); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "Creating service")
	}

	d.alertManager.IsDeploy = true

	return nil
}

func (d *Deployer) getSecret() *corev1.Secret {
	cfg := d.alertManager.GetDefaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: monitorutil.CattleNamespaceName,
			Name:      appName,
		},
		Data: map[string][]byte{
			"config.yml":        data,
			"notification.tmpl": []byte(NotificationTmpl),
		},
	}
}

func getService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: monitorutil.CattleNamespaceName,
			Name:      appSvcName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": appName,
			},
			Ports: []corev1.ServicePort{
				{
					Name: appName,
					Port: 9093,
				},
			},
		},
	}
}

func GetDeployment() *appsv1beta2.Deployment {
	replicas := int32(1)
	return &appsv1beta2.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: monitorutil.CattleNamespaceName,
			Name:      appName,
		},
		Spec: appsv1beta2.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": appName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": appName},
					Name:   appName,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  appName,
							Image: image.Resolve(v3.ToolsSystemImages.AlertSystemImages.AlertManager),
							Args:  []string{"--config.file=/etc/alertmanager/config.yml", "--storage.path=/alertmanager"},
							Ports: []corev1.ContainerPort{
								{
									Name:          appName,
									ContainerPort: 9093,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      appName,
									MountPath: "/alertmanager",
								},
								{
									Name:      "config-volume",
									MountPath: "/etc/alertmanager",
								},
							},
						},
						{
							Name:    "alertmanager-helper",
							Image:   image.Resolve(v3.ToolsSystemImages.AlertSystemImages.AlertManagerHelper),
							Command: []string{"alertmanager-helper"},
							Args:    []string{"--watched-file-list", "/etc/alertmanager"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config-volume",
									MountPath: "/etc/alertmanager",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: appName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: appName,
								},
							},
						},
					},
				},
			},
		},
	}
}
