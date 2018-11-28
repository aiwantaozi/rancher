package monitoring

import (
	"fmt"

	"github.com/rancher/norman/types"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Deprecated: use SystemMonitoringInfo() or ClusterMonitoringInfo() instead.
	CattleNamespaceName          = cattleNamespaceName
	CattleCreatorIDAnnotationKey = "field.cattle.io/creatorId"
	CattleProjectIDAnnotationKey = "field.cattle.io/projectId"
	CattleProjectIDLabelKey      = "field.cattle.io/projectId"
	ClusterAppName               = "cluster-monitoring"
	ProjectAppName               = "project-monitoring"
)

type Level string

const (
	SystemLevel  Level = "system"
	ClusterLevel Level = "cluster"
	ProjectLevel Level = "project"
)

const (
	cattleNamespaceName = "cattle-prometheus"

	// The label info of Namespace
	CattleMonitoringLabelKey = "monitoring.coreos.com"

	// The label info of App, RoleBinding
	appNameLabelKey            = CattleMonitoringLabelKey + "/appName"
	appTargetNamespaceLabelKey = CattleMonitoringLabelKey + "/appTargetNamespace"
	levelLabelKey              = CattleMonitoringLabelKey + "/level"

	// The names of App
	systemAppName           = "system-monitoring"
	metricExpressionAppName = "metric-expression"

	// The headless service name of Prometheus
	prometheusHeadlessServiceName = "prometheus-operated"

	// The label info of PrometheusRule
	CattlePrometheusRuleLabelKey           = "source"
	CattleAlertingPrometheusRuleLabelValue = "rancher-alert"
)

var (
	APIVersion = types.APIVersion{
		Version: "v1",
		Group:   "monitoring.coreos.com",
		Path:    "/v3/project",
	}
)

func OwnedAppListOptions(appName, appTargetNamespace string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s, %s=%s", appNameLabelKey, appName, appTargetNamespaceLabelKey, appTargetNamespace),
	}
}

func OwnedLabels(appName, appTargetNamespace string, level Level) map[string]string {
	return map[string]string{
		appNameLabelKey:            appName,
		appTargetNamespaceLabelKey: appTargetNamespace,
		levelLabelKey:              string(level),
	}
}

func OwnedProjectListOptions(projectName string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", CattleProjectIDAnnotationKey, projectName),
	}
}

func SystemMonitoringInfo() (appName, appTargetNamespace string) {
	return systemAppName, cattleNamespaceName
}

func ClusterMonitoringInfo() (appName, appTargetNamespace string) {
	return ClusterAppName, cattleNamespaceName
}

func ClusterMonitoringMetricsInfo() (appName, appTargetNamespace string) {
	return metricExpressionAppName, cattleNamespaceName
}

func ProjectMonitoringInfo(project *mgmtv3.Project) (appName, appTargetNamespace string) {
	return ProjectAppName, fmt.Sprintf("%s-%s", cattleNamespaceName, project.Name)
}

func ClusterPrometheusEndpoint() (headlessServiceName, namespace, port string) {
	return prometheusHeadlessServiceName, cattleNamespaceName, "9090"
}

func ProjectPrometheusEndpoint(project *mgmtv3.Project) (headlessServiceName, namespace string, port string) {
	return prometheusHeadlessServiceName, fmt.Sprintf("%s-%s", cattleNamespaceName, project.Name), "9090"
}
