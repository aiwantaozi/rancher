package watcher

import (
	"fmt"
	"strconv"

	"github.com/rancher/rancher/pkg/controllers/user/alertpolicy/manager"
	"github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type EventWatcher struct {
	eventLister              v1.EventLister
	clusterAlertPolicyLister v3.ClusterAlertPolicyLister
	alertManager             *manager.AlertManager
	clusterName              string
	clusterLister            v3.ClusterLister
}

func StartEventWatcher(cluster *config.UserContext, manager *manager.AlertManager) {
	events := cluster.Core.Events("")
	eventWatcher := &EventWatcher{
		eventLister:              events.Controller().Lister(),
		clusterAlertPolicyLister: cluster.Management.Management.ClusterAlertPolicies(cluster.ClusterName).Controller().Lister(),
		alertManager:             manager,
		clusterName:              cluster.ClusterName,
		clusterLister:            cluster.Management.Management.Clusters("").Controller().Lister(),
	}

	events.AddHandler("cluster-event-alert-watcher", eventWatcher.Sync)
}

func (l *EventWatcher) Sync(key string, obj *corev1.Event) error {
	if l.alertManager.IsDeploy == false {
		return nil
	}

	if obj == nil {
		return nil
	}

	clusterAlerts, err := l.clusterAlertPolicyLister.List("", labels.NewSelector())
	if err != nil {
		return err
	}

	for _, alert := range clusterAlerts {
		if alert.Status.AlertState == "inactive" || alert.Status.AlertState == "muted" {
			continue
		}
		alertID := alert.Namespace + "-" + alert.Name
		targets := alert.Spec.TargetEvents
		for _, target := range targets {
			if target.EventType == obj.Type && target.ResourceKind == obj.InvolvedObject.Kind {

				clusterDisplayName := l.clusterName
				cluster, err := l.clusterLister.Get("", l.clusterName)
				if err != nil {
					logrus.Warnf("Failed to get cluster for %s: %v", l.clusterName, err)
				} else {
					clusterDisplayName = cluster.Spec.DisplayName
				}

				data := map[string]string{}
				data["alert_type"] = "event"
				data["group_id"] = alertID
				data["group_name"] = alert.Name
				data["event_type"] = target.EventType
				data["resource_kind"] = target.ResourceKind
				data["severity"] = target.Severity
				data["alert_name"] = alert.Spec.DisplayName
				data["cluster_name"] = clusterDisplayName
				data["target_name"] = obj.InvolvedObject.Name
				data["event_count"] = strconv.Itoa(int(obj.Count))
				data["event_message"] = obj.Message
				data["event_firstseen"] = fmt.Sprintf("%s", obj.FirstTimestamp)
				data["event_lastseen"] = fmt.Sprintf("%s", obj.LastTimestamp)

				if err := l.alertManager.SendAlert(data); err != nil {
					logrus.Debugf("Failed to send alert: %v", err)
				}
			}
		}

	}

	return nil
}
