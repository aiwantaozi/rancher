package monitoring

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/juju/errors"
	"github.com/rancher/rancher/pkg/controllers/user/helm/common"
	"github.com/rancher/rancher/pkg/monitoring"
	"github.com/rancher/rancher/pkg/project"
	"github.com/rancher/rancher/pkg/ref"
	"github.com/rancher/rancher/pkg/settings"
	appsv1beta2 "github.com/rancher/types/apis/apps/v1beta2"
	corev1 "github.com/rancher/types/apis/core/v1"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	monitoringv1 "github.com/rancher/types/apis/monitoring.cattle.io/v1"
	projectv3 "github.com/rancher/types/apis/project.cattle.io/v3"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	cattleCreatorIDAnnotationKey    = "field.cattle.io/creatorId"
	cattleProjectIDAnnotationKey    = "field.cattle.io/projectId"
	cattleMonitoringAppNameLabelKey = "monitoring.cattle.io/appName"
)

type appHandler struct {
	templateVersionClient mgmtv3.TemplateVersionInterface
	namespaceClient       corev1.NamespaceInterface
	appsGetter            projectv3.AppsGetter
	projectsGetter        mgmtv3.ProjectsGetter
}

type clusterHandler struct {
	ctx             context.Context
	clusterName     string
	clustersClient  mgmtv3.ClusterInterface
	app             *appHandler
	workloadsClient appsv1beta2.Interface
}

func (ch *clusterHandler) sync(key string, cluster *mgmtv3.Cluster) error {
	if cluster == nil || cluster.DeletionTimestamp != nil ||
		cluster.Name != ch.clusterName ||
		!mgmtv3.ClusterConditionAdditionalCRDCreated.IsTrue(cluster) {
		return nil
	}

	if cluster.Spec.EnableClusterMonitoring == nil {
		return nil
	}

	clusterTag := getClusterTag(cluster)
	src := cluster
	cpy := src.DeepCopy()

	err := ch.doSync(clusterTag, cpy)

	if !reflect.DeepEqual(cpy, src) {
		_, err := ch.clustersClient.Update(cpy)
		if err != nil {
			return errors.Annotatef(err, "failed to update cluster %s", clusterTag)
		}
	}

	return err
}

func (ch *clusterHandler) doSync(clusterTag string, cluster *mgmtv3.Cluster) error {
	enableClusterMonitoring := *cluster.Spec.EnableClusterMonitoring

	if enableClusterMonitoring {
		if !mgmtv3.ClusterConditionMonitoringEnabled.IsTrue(cluster) {
			if err := ch.app.deploySystemMonitoring(clusterTag, cluster); err != nil {
				mgmtv3.ClusterConditionMonitoringEnabled.Unknown(cluster)
				mgmtv3.ClusterConditionMonitoringEnabled.Message(cluster, err.Error())
				return err
			}

			if err := ch.detectWhileInstall(cluster); err != nil {
				mgmtv3.ClusterConditionMonitoringEnabled.Unknown(cluster)
				mgmtv3.ClusterConditionMonitoringEnabled.Message(cluster, err.Error())
				return err
			}

			mgmtv3.ClusterConditionMonitoringEnabled.True(cluster)
			mgmtv3.ClusterConditionMonitoringEnabled.Message(cluster, "")
		}
	} else {
		hasConditionMonitoringEnabled := false
		for _, cond := range cluster.Status.Conditions {
			if string(cond.Type) == string(mgmtv3.ClusterConditionMonitoringEnabled) {
				hasConditionMonitoringEnabled = true
			}
		}

		if hasConditionMonitoringEnabled && !mgmtv3.ClusterConditionMonitoringEnabled.IsFalse(cluster) {
			if err := ch.app.withdrawSystemMonitoring(clusterTag, cluster); err != nil {
				mgmtv3.ClusterConditionMonitoringEnabled.Unknown(cluster)
				mgmtv3.ClusterConditionMonitoringEnabled.Message(cluster, err.Error())
				return err
			}

			if err := ch.detectWhileUninstall(cluster); err != nil {
				mgmtv3.ClusterConditionMonitoringEnabled.Unknown(cluster)
				mgmtv3.ClusterConditionMonitoringEnabled.Message(cluster, err.Error())
				return err
			}

			mgmtv3.ClusterConditionMonitoringEnabled.False(cluster)
			mgmtv3.ClusterConditionMonitoringEnabled.Message(cluster, "")
		}
	}

	return nil
}

func (ch *clusterHandler) detectWhileInstall(cluster *mgmtv3.Cluster) error {
	time.Sleep(5 * time.Second)

	if cluster.Status.MonitoringStatus == nil {
		cluster.Status.MonitoringStatus = &mgmtv3.MonitoringStatus{
			// in case of races
			Conditions: []mgmtv3.MonitoringCondition{
				{Type: mgmtv3.ClusterConditionType(ConditionGrafanaDeployed), Status: k8scorev1.ConditionFalse},
				{Type: mgmtv3.ClusterConditionType(ConditionNodeExporterDeployed), Status: k8scorev1.ConditionFalse},
				{Type: mgmtv3.ClusterConditionType(ConditionKubeStateExporterDeployed), Status: k8scorev1.ConditionFalse},
				{Type: mgmtv3.ClusterConditionType(ConditionOperatorDeployed), Status: k8scorev1.ConditionFalse},
				{Type: mgmtv3.ClusterConditionType(ConditionPrometheusDeployed), Status: k8scorev1.ConditionFalse},
				// without Alertmanager
				// {Type: mgmtv3.ClusterConditionType(ConditionAlertmanagerDeployed), Status: k8scorev1.ConditionFalse},
			},
		}
	}

	monitoringStatus := cluster.Status.MonitoringStatus

	return stream(
		func() error {
			return ch.isGrafanaDeployed(monitoringStatus, cluster.Name)
		},
		func() error {
			return ch.isNodeExporterDeployed(monitoringStatus)
		},
		func() error {
			return ch.isKubeStateExporterDeployed(monitoringStatus)
		},
		func() error {
			return ch.isOperatorDeployed(monitoringStatus)
		},
		func() error {
			return ch.isPrometheusDeployed(monitoringStatus)
		},
		// without Alertmanager
		// func() error {
		// 	return ch.isAlertmanagerDeployed(monitoringStatus)
		// },
	)
}

func (ch *clusterHandler) detectWhileUninstall(cluster *mgmtv3.Cluster) error {
	if cluster.Status.MonitoringStatus == nil {
		return nil
	}

	time.Sleep(5 * time.Second)

	monitoringStatus := cluster.Status.MonitoringStatus

	return stream(
		func() error {
			return ch.isPrometheusWithdrew(monitoringStatus)
		},
		func() error {
			return ch.isOperatorWithdrew(monitoringStatus)
		},
		func() error {
			return ch.isKubeStateExporterWithdrew(monitoringStatus)
		},
		func() error {
			return ch.isNodeExporterWithdrew(monitoringStatus)
		},
		func() error {
			return ch.isGrafanaWithdrew(monitoringStatus)
		},
		// without Alertmanager
		// func() error {
		// 	return ch.isAlertmanagerWithdrew(monitoringStatus)
		// },
	)
}

func (ah *appHandler) deploySystemMonitoring(clusterTag string, cluster *mgmtv3.Cluster) error {
	clusterName := cluster.Name
	clusterCreatorID := cluster.Annotations[cattleCreatorIDAnnotationKey]

	// check system monitoring app
	systemMonitoringApps, err := ah.appsGetter.Apps(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", cattleMonitoringAppNameLabelKey, monitoring.SystemMonitoringAppName),
	})
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Annotatef(err, "failed to find %q app of cluster %s", monitoring.SystemMonitoringAppName, clusterTag)
	}

	systemMonitoringApps = systemMonitoringApps.DeepCopy()
	for _, app := range systemMonitoringApps.Items {
		if app.Name == monitoring.SystemMonitoringAppName {
			if app.DeletionTimestamp != nil {
				return errors.Annotatef(err, "stale %q app of cluster %s is still on terminating", monitoring.SystemMonitoringAppName, clusterTag)
			}

			return nil
		}
	}

	// check monitoring namespace
	namespace := &k8scorev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: monitoring.SystemMonitoringNamespaceName,
		},
	}
	if _, err := ah.namespaceClient.Create(namespace); err != nil && !k8serrors.IsAlreadyExists(err) {
		return errors.Annotatef(err, "failed to create %q namespace for cluster %s", monitoring.SystemMonitoringNamespaceName, clusterTag)
	}

	deployNamespace, err := ah.namespaceClient.Get(monitoring.SystemMonitoringNamespaceName, metav1.GetOptions{})
	if err != nil {
		return errors.Annotatef(err, "failed to find %q namespace of cluster %s", monitoring.SystemMonitoringNamespaceName, clusterTag)
	}
	deployNamespace = deployNamespace.DeepCopy()

	deployProjectName := ""
	if projectName, ok := deployNamespace.Annotations[cattleProjectIDAnnotationKey]; ok {
		deployProjectName = projectName
	} else {
		// take system project
		defaultSystemProjects, _ := ah.projectsGetter.Projects(clusterName).List(metav1.ListOptions{
			LabelSelector: "authz.management.cattle.io/system-project=true",
		})

		var deployProject *mgmtv3.Project
		defaultSystemProjects = defaultSystemProjects.DeepCopy()
		for _, defaultProject := range defaultSystemProjects.Items {
			deployProject = &defaultProject

			if defaultProject.Spec.DisplayName == project.System {
				break
			}
		}
		if deployProject == nil {
			return fmt.Errorf("failed to find any cattle system project of cluster %s", clusterTag)
		}

		deployProjectName = fmt.Sprintf("%s:%s", clusterName, deployProject.Name)

		deployNamespace.Annotations[cattleProjectIDAnnotationKey] = deployProjectName
		_, err := ah.namespaceClient.Update(deployNamespace)
		if err != nil {
			return errors.Annotatef(err, "failed to move namespace %s to project %s", monitoring.SystemMonitoringNamespaceName, deployProject.Spec.DisplayName)
		}
	}

	// check monitoring
	_, projectID := ref.Parse(deployProjectName)
	catalogID := settings.SystemMonitoringCatalogID.Get()
	templateVersionID, err := common.ParseExternalID(catalogID)
	if err != nil {
		return errors.Annotatef(err, "failed to parse catalog ID %q", catalogID)
	}

	if _, err := ah.templateVersionClient.Get(templateVersionID, metav1.GetOptions{}); err != nil {
		return errors.Annotatef(err, "failed to find catalog by ID %q", catalogID)
	}

	app := &projectv3.App{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				cattleCreatorIDAnnotationKey: clusterCreatorID,
			},
			Labels: map[string]string{
				cattleMonitoringAppNameLabelKey: monitoring.SystemMonitoringAppName,
			},
			Name:      monitoring.SystemMonitoringAppName,
			Namespace: projectID,
		},
		Spec: projectv3.AppSpec{
			Answers: map[string]string{
				"apiGroup":                                          monitoringv1.GroupName,
				"alertmanager.enabled":                              "false",
				"alertmanager.enabledDefaultPrometheusRules":        "false",
				"alertmanager.apiGroup":                             monitoringv1.GroupName,
				"alertmanager.ingress.enabled":                      "false",
				"alertmanager.ingress.hosts[0]":                     "xip.io",
				"alertmanager.persistence.enabled":                  "false",
				"alertmanager.persistence.size":                     "10Gi",
				"alertmanager.persistence.storageClass":             "",
				"exporter-kube-state.apiGroup":                      monitoringv1.GroupName,
				"exporter-kube-state.enabledDefaultPrometheusRules": "false",
				"exporter-kubelets.apiGroup":                        monitoringv1.GroupName,
				"exporter-kubelets.enabledDefaultPrometheusRules":   "false",
				"exporter-kubelets.https":                           "false",
				"exporter-kubernetes.apiGroup":                      monitoringv1.GroupName,
				"exporter-kubernetes.enabledDefaultPrometheusRules": "false",
				"exporter-node.apiGroup":                            monitoringv1.GroupName,
				"exporter-node.enabledDefaultPrometheusRules":       "false",
				"grafana.apiGroup":                                  monitoringv1.GroupName,
				"grafana.ingress.enabled":                           "false",
				"grafana.ingress.hosts[0]":                          "xip.io",
				"grafana.persistence.enabled":                       "false",
				"grafana.persistence.size":                          "10Gi",
				"grafana.persistence.storageClass":                  "",
				"grafana.rancherClusterId":                          clusterName,
				"prometheus.apiGroup":                               monitoringv1.GroupName,
				"prometheus.enabledDefaultPrometheusRules":          "false",
				"prometheus.ingress.enabled":                        "false",
				"prometheus.ingress.hosts[0]":                       "xip.io",
				"prometheus.service.type":                           "ClusterIP",
				"prometheus.service.nodePort":                       "",
				"prometheus.persistence.enabled":                    "false",
				"prometheus.persistence.size":                       "10Gi",
				"prometheus.persistence.storageClass":               "",
				"prometheus.alertingEndpoints[0].name":              "alertmanager",
				"prometheus.alertingEndpoints[0].namespace":         monitoring.SystemMonitoringNamespaceName,
				"prometheus.alertingEndpoints[0].port":              "alertmanager",
				"prometheus.rulesSelector.matchLabels[0].source":    "rancher-alert",
			},
			Description:     "System Monitoring",
			ExternalID:      catalogID,
			ProjectName:     deployProjectName,
			TargetNamespace: monitoring.SystemMonitoringNamespaceName,
		},
	}

	if _, err := ah.appsGetter.Apps(projectID).Create(app); err != nil && !k8serrors.IsAlreadyExists(err) {
		return errors.Annotatef(err, "failed to create %q app for cluster %s", monitoring.SystemMonitoringAppName, clusterTag)
	}

	return nil
}

func (ah *appHandler) withdrawSystemMonitoring(clusterTag string, cluster *mgmtv3.Cluster) error {
	// check system monitoring app
	systemMonitoringApps, err := ah.appsGetter.Apps(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", cattleMonitoringAppNameLabelKey, monitoring.SystemMonitoringAppName),
	})
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Annotatef(err, "failed to find %q app of cluster %s", monitoring.SystemMonitoringAppName, clusterTag)
	}

	systemMonitoringApps = systemMonitoringApps.DeepCopy()
	for _, app := range systemMonitoringApps.Items {
		if app.Name == monitoring.SystemMonitoringAppName && app.DeletionTimestamp == nil {
			if err := ah.appsGetter.Apps(app.Namespace).Delete(monitoring.SystemMonitoringAppName, &metav1.DeleteOptions{}); err != nil {
				return errors.Annotatef(err, "failed to remove %q app of cluster %s", monitoring.SystemMonitoringAppName, clusterTag)
			}

			return nil
		}
	}

	return nil
}

func getClusterTag(cluster *mgmtv3.Cluster) string {
	return fmt.Sprintf("%s(%s)", cluster.Spec.DisplayName, cluster.Name)
}
