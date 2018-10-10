package monitoring

import (
	"fmt"

	"github.com/juju/errors"
	"github.com/rancher/rancher/pkg/monitoring"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

var (
	ConditionGrafanaDeployed           = condition(mgmtv3.MonitoringConditionGrafanaDeployed)
	ConditionNodeExporterDeployed      = condition(mgmtv3.MonitoringConditionNodeExporterDeployed)
	ConditionKubeStateExporterDeployed = condition(mgmtv3.MonitoringConditionKubeStateExporterDeployed)
	ConditionOperatorDeployed          = condition(mgmtv3.MonitoringConditionOperatorDeployed)
	ConditionPrometheusDeployed        = condition(mgmtv3.MonitoringConditionPrometheusDeployed)
	ConditionAlertmanagerDeployed      = condition(mgmtv3.MonitoringConditionAlertmaanagerDeployed)
)

func (ch *clusterHandler) isGrafanaDeployed(monitoringStatus *mgmtv3.MonitoringStatus, clusterName string) error {
	_, err := ConditionGrafanaDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := ch.workloadsClient.Deployments(monitoring.SystemMonitoringNamespaceName).Get("grafana-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Grafana Deployment isn't deployed")
			}

			return nil, errors.Annotate(err, "failed to get Grafana Deployment information")
		}

		status := obj.Status
		if status.Replicas != (status.AvailableReplicas - status.UnavailableReplicas) {
			return nil, errors.New("Grafana Deployment is deploying")
		}

		monitoringStatus.GrafanaEndpoint = fmt.Sprintf("/k8s/clusters/%s/api/v1/namespaces/%s/services/http:grafana-nginx-%s:80/proxy/", clusterName, monitoring.SystemMonitoringNamespaceName, monitoring.SystemMonitoringAppName)

		return monitoringStatus, nil
	})

	return err
}

func (ch *clusterHandler) isGrafanaWithdrew(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionGrafanaDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := ch.workloadsClient.Deployments(monitoring.SystemMonitoringNamespaceName).Get("grafana-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				monitoringStatus.GrafanaEndpoint = ""
				return monitoringStatus, nil
			}

			return nil, errors.Annotate(err, "failed to get Grafana Deployment information")
		}

		return nil, errors.New("Grafana Deployment is withdrawing")
	})

	return err
}

func (ch *clusterHandler) isNodeExporterDeployed(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionNodeExporterDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := ch.workloadsClient.DaemonSets(monitoring.SystemMonitoringNamespaceName).Get("exporter-node-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Node Exporter DaemonSet isn't deployed")
			}

			return nil, errors.Annotate(err, "failed to get Node Exporter DaemonSet information")
		}

		if obj.Status.DesiredNumberScheduled != obj.Status.CurrentNumberScheduled {
			return nil, errors.New("Node Exporter DaemonSet is deploying")
		}

		return monitoringStatus, nil
	})

	return err
}

func (ch *clusterHandler) isNodeExporterWithdrew(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionNodeExporterDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := ch.workloadsClient.DaemonSets(monitoring.SystemMonitoringNamespaceName).Get("exporter-node-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return monitoringStatus, nil
			}

			return nil, errors.Annotate(err, "failed to get Node Exporter DaemonSet information")
		}

		return nil, errors.New("Node Exporter DaemonSet is withdrawing")
	})

	return err
}

func (ch *clusterHandler) isKubeStateExporterDeployed(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionKubeStateExporterDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := ch.workloadsClient.Deployments(monitoring.SystemMonitoringNamespaceName).Get("exporter-kube-state-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Kube State Exporter Deployment isn't deployed")
			}

			return nil, errors.Annotate(err, "failed to get Kube State Exporter Deployment information")
		}

		status := obj.Status
		if status.Replicas != (status.AvailableReplicas - status.UnavailableReplicas) {
			return nil, errors.New("Kube State Exporter Deployment is deploying")
		}

		return monitoringStatus, nil
	})

	return err
}

func (ch *clusterHandler) isKubeStateExporterWithdrew(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionKubeStateExporterDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := ch.workloadsClient.Deployments(monitoring.SystemMonitoringNamespaceName).Get("exporter-kube-state-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return monitoringStatus, nil
			}

			return nil, errors.Annotate(err, "failed to get Kube State Exporter Deployment information")
		}

		return nil, errors.New("Kube State Exporter Deployment is withdrawing")
	})

	return err
}

func (ch *clusterHandler) isOperatorDeployed(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionOperatorDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := ch.workloadsClient.Deployments(monitoring.SystemMonitoringNamespaceName).Get("prometheus-operator-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Prometheus Operator Deployment isn't deployed")
			}

			return nil, errors.Annotate(err, "failed to get Prometheus Operator Deployment information")
		}

		status := obj.Status
		if status.Replicas != (status.AvailableReplicas - status.UnavailableReplicas) {
			return nil, errors.New("Prometheus Operator Deployment is deploying")
		}

		return monitoringStatus, nil
	})

	return err
}

func (ch *clusterHandler) isOperatorWithdrew(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionOperatorDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := ch.workloadsClient.Deployments(monitoring.SystemMonitoringNamespaceName).Get("prometheus-operator-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return monitoringStatus, nil
			}

			return nil, errors.Annotate(err, "failed to get Prometheus Operator Deployment information")
		}

		return nil, errors.New("Prometheus Operator Deployment is withdrawing")
	})

	return err
}

func (ch *clusterHandler) isPrometheusDeployed(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionPrometheusDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := ch.workloadsClient.StatefulSets(monitoring.SystemMonitoringNamespaceName).Get("prometheus-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Prometheus StatefulSet isn't deployed")
			}

			return nil, errors.Annotate(err, "failed to get Prometheus StatefulSet information")
		}

		if obj.Status.Replicas != obj.Status.CurrentReplicas {
			return nil, errors.New("Prometheus StatefulSet is deploying")
		}

		return monitoringStatus, nil
	})

	return err
}

func (ch *clusterHandler) isPrometheusWithdrew(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionPrometheusDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := ch.workloadsClient.StatefulSets(monitoring.SystemMonitoringNamespaceName).Get("prometheus-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return monitoringStatus, nil
			}

			return nil, errors.Annotate(err, "failed to get Prometheus StatefulSet information")
		}

		return nil, errors.New("Prometheus StatefulSet is withdrawing")
	})

	return err
}

func (ch *clusterHandler) isAlertmanagerDeployed(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionAlertmanagerDeployed.DoUntilTrue(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		obj, err := ch.workloadsClient.StatefulSets(monitoring.SystemMonitoringNamespaceName).Get("alertmanager-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, errors.New("Alertmanager StatefulSet isn't deployed")
			}

			return nil, errors.Annotate(err, "failed to get Alertmanager StatefulSet information")
		}

		if obj.Status.Replicas != obj.Status.CurrentReplicas {
			return nil, errors.New("Alertmanager StatefulSet is deploying")
		}

		return monitoringStatus, nil
	})

	return err
}

func (ch *clusterHandler) isAlertmanagerWithdrew(monitoringStatus *mgmtv3.MonitoringStatus) error {
	_, err := ConditionAlertmanagerDeployed.DoUntilFalse(monitoringStatus, func() (*mgmtv3.MonitoringStatus, error) {
		_, err := ch.workloadsClient.StatefulSets(monitoring.SystemMonitoringNamespaceName).Get("alertmanager-"+monitoring.SystemMonitoringAppName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return monitoringStatus, nil
			}

			return nil, errors.Annotate(err, "failed to get Alertmanager StatefulSet information")
		}

		return nil, errors.New("Alertmanager StatefulSet is withdrawing")
	})

	return err
}

func stream(funcs ...func() error) error {
	return utilerrors.AggregateGoroutines(funcs...)
}
