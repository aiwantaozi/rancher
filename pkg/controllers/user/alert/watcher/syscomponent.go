package watcher

import (
	"context"
	"strings"
	"time"

	"github.com/rancher/rancher/pkg/controllers/user/alert/manager"
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
	componentStatuses      v1.ComponentStatusInterface
	clusterAlertRuleLister v3.ClusterAlertRuleLister
	alertManager           *manager.AlertManager
	clusterName            string
	clusterLister          v3.ClusterLister
}

func StartSysComponentWatcher(ctx context.Context, cluster *config.UserContext, manager *manager.AlertManager) {

	s := &SysComponentWatcher{
		componentStatuses:      cluster.Core.ComponentStatuses(""),
		clusterAlertRuleLister: cluster.Management.Management.ClusterAlertRules(cluster.ClusterName).Controller().Lister(),
		alertManager:           manager,
		clusterName:            cluster.ClusterName,
		clusterLister:          cluster.Management.Management.Clusters("").Controller().Lister(),
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

	clusterAlerts, err := w.clusterAlertRuleLister.List("", labels.NewSelector())
	if err != nil {
		return err
	}

	statuses, err := w.componentStatuses.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, rule := range clusterAlerts {
		if rule.Status.State == "inactive" || rule.Spec.SystemServiceRule == nil {
			continue
		}
		if rule.Spec.SystemServiceRule != nil {
			w.checkComponentHealthy(rule.Name, rule.Namespace, statuses, rule)
		}
	}
	return nil
}

func (w *SysComponentWatcher) checkComponentHealthy(name, namespace string, statuses *v1.ComponentStatusList, alert *v3.ClusterAlertRule) {
	groupID := namespace + "-" + name
	for _, cs := range statuses.Items {
		if strings.HasPrefix(cs.Name, alert.Spec.SystemServiceRule.Condition) {
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
						data["severity"] = alert.Spec.Severity
						data["rule_id"] = alert.Name
						data["cluster_name"] = clusterDisplayName
						data["component_name"] = alert.Spec.SystemServiceRule.Condition

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
