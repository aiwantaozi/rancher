package watcher

import (
	"context"
	"strings"
	"time"

	"github.com/rancher/rancher/pkg/controllers/user/alertpolicy/manager"
	"github.com/rancher/rancher/pkg/ticker"
	"github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SysComponentWatcher struct {
	componentStatuses        v1.ComponentStatusInterface
	clusterAlertPolicyLister v3.ClusterAlertPolicyLister
	alertManager             *manager.AlertManager
	clusterName              string
	clusterLister            v3.ClusterLister
}

func StartSysComponentWatcher(ctx context.Context, cluster *config.UserContext, manager *manager.AlertManager) {

	s := &SysComponentWatcher{
		componentStatuses:        cluster.Core.ComponentStatuses(""),
		clusterAlertPolicyLister: cluster.Management.Management.ClusterAlertPolicies(cluster.ClusterName).Controller().Lister(),
		alertManager:             manager,
		clusterName:              cluster.ClusterName,
		clusterLister:            cluster.Management.Management.Clusters("").Controller().Lister(),
	}
	go s.watch(ctx, syncInterval)
}

func (w *SysComponentWatcher) watch(ctx context.Context, interval time.Duration) {
	for range ticker.Context(ctx, interval) {
		err := w.watchRule()
		if err != nil {
			logrus.Infof("Failed to watch system component, error: %v", err)
		}
	}
}

func (w *SysComponentWatcher) watchRule() error {
	if w.alertManager.IsDeploy == false {
		return nil
	}

	clusterAlerts, err := w.clusterAlertPolicyLister.List("", labels.NewSelector())
	if err != nil {
		return err
	}

	statuses, err := w.componentStatuses.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, group := range clusterAlerts {
		if group.Status.AlertState == "inactive" {
			continue
		}
		if group.Spec.TargetSystemServices != nil {
			for _, alert := range group.Spec.TargetSystemServices {
				w.checkComponentHealthy(group.Name, group.Namespace, statuses, &alert)
			}
		}
	}
	return nil
}

func (w *SysComponentWatcher) checkComponentHealthy(name, namespace string, statuses *v1.ComponentStatusList, alert *v3.TargetSystemService) {
	groupID := namespace + "-" + name
	for _, cs := range statuses.Items {
		if strings.HasPrefix(cs.Name, alert.Condition) {
			for _, cond := range cs.Conditions {
				if cond.Type == corev1.ComponentHealthy {
					if cond.Status == corev1.ConditionFalse {

						clusterDisplayName := w.clusterName
						cluster, err := w.clusterLister.Get("", w.clusterName)
						if err != nil {
							logrus.Warnf("Failed to get cluster for %s: %v", w.clusterName, err)
						} else {
							clusterDisplayName = cluster.Spec.DisplayName
						}

						data := map[string]string{}
						data["alert_type"] = "systemService"
						data["group_id"] = groupID
						data["group_name"] = name
						data["severity"] = alert.Severity
						// data["alert_name"] = alert.Spec.DisplayName //todo
						data["cluster_name"] = clusterDisplayName
						data["component_name"] = alert.Condition

						if cond.Message != "" {
							data["logs"] = cond.Message
						}
						if err := w.alertManager.SendAlert(data); err != nil {
							logrus.Debugf("Failed to send alert: %v", err)
						}
						return
					}
				}
			}
		}
	}

}
