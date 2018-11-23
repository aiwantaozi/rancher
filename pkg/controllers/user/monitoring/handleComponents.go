// All component names base on rancher-monitoring chart
package monitoring

import (
	"fmt"

	"github.com/pkg/errors"
	appsv1beta2 "github.com/rancher/types/apis/apps/v1beta2"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ConditionGrafanaDeployed           = condition(mgmtv3.MonitoringConditionGrafanaDeployed)
	ConditionNodeExporterDeployed      = condition(mgmtv3.MonitoringConditionNodeExporterDeployed)
	ConditionKubeStateExporterDeployed = condition(mgmtv3.MonitoringConditionKubeStateExporterDeployed)
	ConditionPrometheusDeployed        = condition(mgmtv3.MonitoringConditionPrometheusDeployed)
)

func isGrafanaDeployed(workloadsClient appsv1beta2.Interface, appNamespace, appNameSuffix string, monitoringStatus *mgmtv3.MonitoringStatus, clusterName string) error {
	_, err := ConditionGrafanaDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := workloadsClient.Deployments(appNamespace).Get("grafana-"+appNameSuffix, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Grafana Deployment isn't deployed")
			}

			return nil, errors.Wrap(err, "failed to get Grafana Deployment information")
		}

		status := obj.Status
		if status.Replicas != (status.AvailableReplicas - status.UnavailableReplicas) {
			return nil, errors.New("Grafana Deployment is deploying")
		}

		monitoringStatus.GrafanaEndpoint = fmt.Sprintf("/k8s/clusters/%s/api/v1/namespaces/%s/services/http:grafana-%s:80/proxy/", clusterName, appNamespace, appNameSuffix)

		return monitoringStatus, nil
	})

	return err
}

func isGrafanaWithdrew(workloadsClient appsv1beta2.Interface, appNamespace, appNameSuffix string, monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionGrafanaDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := workloadsClient.Deployments(appNamespace).Get("grafana-"+appNameSuffix, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				monitoringStatus.GrafanaEndpoint = ""
				return monitoringStatus, nil
			}

			return nil, errors.Wrap(err, "failed to get Grafana Deployment information")
		}

		return nil, errors.New("Grafana Deployment is withdrawing")
	})

	return err
}

func isNodeExporterDeployed(workloadsClient appsv1beta2.Interface, appNamespace, appNameSuffix string, monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionNodeExporterDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := workloadsClient.DaemonSets(appNamespace).Get("exporter-node-"+appNameSuffix, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Node Exporter DaemonSet isn't deployed")
			}

			return nil, errors.Wrap(err, "failed to get Node Exporter DaemonSet information")
		}

		if obj.Status.DesiredNumberScheduled != obj.Status.CurrentNumberScheduled {
			return nil, errors.New("Node Exporter DaemonSet is deploying")
		}

		return monitoringStatus, nil
	})

	return err
}

func isNodeExporterWithdrew(workloadsClient appsv1beta2.Interface, appNamespace, appNameSuffix string, monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionNodeExporterDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := workloadsClient.DaemonSets(appNamespace).Get("exporter-node-"+appNameSuffix, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return monitoringStatus, nil
			}

			return nil, errors.Wrap(err, "failed to get Node Exporter DaemonSet information")
		}

		return nil, errors.New("Node Exporter DaemonSet is withdrawing")
	})

	return err
}

func isKubeStateExporterDeployed(workloadsClient appsv1beta2.Interface, appNamespace, appNameSuffix string, monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionKubeStateExporterDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := workloadsClient.Deployments(appNamespace).Get("exporter-kube-state-"+appNameSuffix, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Kube State Exporter Deployment isn't deployed")
			}

			return nil, errors.Wrap(err, "failed to get Kube State Exporter Deployment information")
		}

		status := obj.Status
		if status.Replicas != (status.AvailableReplicas - status.UnavailableReplicas) {
			return nil, errors.New("Kube State Exporter Deployment is deploying")
		}

		return monitoringStatus, nil
	})

	return err
}

func isKubeStateExporterWithdrew(workloadsClient appsv1beta2.Interface, appNamespace, appNameSuffix string, monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionKubeStateExporterDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := workloadsClient.Deployments(appNamespace).Get("exporter-kube-state-"+appNameSuffix, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return monitoringStatus, nil
			}

			return nil, errors.Wrap(err, "failed to get Kube State Exporter Deployment information")
		}

		return nil, errors.New("Kube State Exporter Deployment is withdrawing")
	})

	return err
}

func isPrometheusDeployed(workloadsClient appsv1beta2.Interface, appNamespace, appNameSuffix string, monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionPrometheusDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := workloadsClient.StatefulSets(appNamespace).Get("prometheus-"+appNameSuffix, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Prometheus StatefulSet isn't deployed")
			}

			return nil, errors.Wrap(err, "failed to get Prometheus StatefulSet information")
		}

		if obj.Status.Replicas != obj.Status.CurrentReplicas {
			return nil, errors.New("Prometheus StatefulSet is deploying")
		}

		return monitoringStatus, nil
	})

	return err
}

func isPrometheusWithdrew(workloadsClient appsv1beta2.Interface, appNamespace, appNameSuffix string, monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionPrometheusDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := workloadsClient.StatefulSets(appNamespace).Get("prometheus-"+appNameSuffix, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return monitoringStatus, nil
			}

			return nil, errors.Wrap(err, "failed to get Prometheus StatefulSet information")
		}

		return nil, errors.New("Prometheus StatefulSet is withdrawing")
	})

	return err
}
