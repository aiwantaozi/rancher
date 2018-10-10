package watcher

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rancher/norman/controller"
	"github.com/rancher/rancher/pkg/controllers/user/alertpolicy/manager"
	"github.com/rancher/rancher/pkg/ticker"
	"github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type PodWatcher struct {
	podLister                v1.PodLister
	alertManager             *manager.AlertManager
	projectAlertPolicies     v3.ProjectAlertPolicyInterface
	projectAlertPolicyLister v3.ProjectAlertPolicyLister
	clusterName              string
	podRestartTrack          sync.Map
	clusterLister            v3.ClusterLister
}

type restartTrack struct {
	Count int32
	Time  time.Time
}

func StartPodWatcher(ctx context.Context, cluster *config.UserContext, manager *manager.AlertManager) {
	projectAlertPolicies := cluster.Management.Management.ProjectAlertPolicies("")

	podWatcher := &PodWatcher{
		podLister:                cluster.Core.Pods("").Controller().Lister(),
		projectAlertPolicies:     projectAlertPolicies,
		projectAlertPolicyLister: projectAlertPolicies.Controller().Lister(),
		alertManager:             manager,
		clusterName:              cluster.ClusterName,
		podRestartTrack:          sync.Map{},
		clusterLister:            cluster.Management.Management.Clusters("").Controller().Lister(),
	}

	projectAlertLifecycle := &ProjectAlertLifecycle{
		podWatcher: podWatcher,
	}
	projectAlertPolicies.AddClusterScopedLifecycle("pod-target-alert-watcher", cluster.ClusterName, projectAlertLifecycle)

	go podWatcher.watch(ctx, syncInterval)
}

func (w *PodWatcher) watch(ctx context.Context, interval time.Duration) {
	for range ticker.Context(ctx, interval) {
		err := w.watchRule()
		if err != nil {
			logrus.Infof("Failed to watch pod, error: %v", err)
		}
	}
}

type ProjectAlertLifecycle struct {
	podWatcher *PodWatcher
}

func (l *ProjectAlertLifecycle) Create(obj *v3.ProjectAlertPolicy) (*v3.ProjectAlertPolicy, error) {
	l.podWatcher.podRestartTrack.Store(obj.Namespace+":"+obj.Name, make([]restartTrack, 0))
	return obj, nil
}

func (l *ProjectAlertLifecycle) Updated(obj *v3.ProjectAlertPolicy) (*v3.ProjectAlertPolicy, error) {
	return obj, nil
}

func (l *ProjectAlertLifecycle) Remove(obj *v3.ProjectAlertPolicy) (*v3.ProjectAlertPolicy, error) {
	l.podWatcher.podRestartTrack.Delete(obj.Namespace + ":" + obj.Name)
	return obj, nil
}

func (w *PodWatcher) watchRule() error {
	if w.alertManager.IsDeploy == false {
		return nil
	}

	projectAlerts, err := w.projectAlertPolicyLister.List("", labels.NewSelector())
	if err != nil {
		return err
	}

	pAlerts := []*v3.ProjectAlertPolicy{}
	for _, alert := range projectAlerts {
		if controller.ObjectInCluster(w.clusterName, alert) {
			pAlerts = append(pAlerts, alert)
		}
	}

	for _, alert := range pAlerts {
		if alert.Status.AlertState == "inactive" {
			continue
		}

		for _, targetPod := range alert.Spec.TargetPods {
			parts := strings.Split(targetPod.PodName, ":")
			ns := parts[0]
			podID := parts[1]
			newPod, err := w.podLister.Get(ns, podID)
			if err != nil {
				//TODO: what to do when pod not found
				if kerrors.IsNotFound(err) || newPod == nil {
					if err = w.projectAlertPolicies.DeleteNamespaced(alert.Namespace, alert.Name, &metav1.DeleteOptions{}); err != nil {
						return err
					}
				}
				logrus.Debugf("Failed to get pod %s: %v", podID, err)

				continue
			}

			switch targetPod.Condition {
			case "notrunning":
				w.checkPodRunning(alert.Name, alert.Namespace, newPod, &targetPod)
			case "notscheduled":
				w.checkPodScheduled(alert.Name, alert.Namespace, newPod, &targetPod)
			case "restarts":
				w.checkPodRestarts(alert.Name, alert.Namespace, newPod, &targetPod)
			}
		}
	}

	return nil
}

func (w *PodWatcher) checkPodRestarts(name, namespace string, pod *corev1.Pod, alert *v3.TargetPod) {

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Running == nil {
			curCount := containerStatus.RestartCount
			preCount := w.getRestartTimeFromTrack(name, namespace, alert, curCount)

			if curCount-preCount >= int32(alert.RestartTimes) {
				alertID := namespace + "-" + name
				details := ""
				if containerStatus.State.Waiting != nil {
					details = containerStatus.State.Waiting.Message
				}

				clusterDisplayName := w.clusterName
				cluster, err := w.clusterLister.Get("", w.clusterName)
				if err != nil {
					logrus.Warnf("Failed to get cluster for %s: %v", w.clusterName, err)
				} else {
					clusterDisplayName = cluster.Spec.DisplayName
				}

				data := map[string]string{}
				data["alert_type"] = "podRestarts"
				data["alert_id"] = alertID
				data["severity"] = alert.Severity
				// data["alert_name"] = alert.Spec.DisplayName
				data["cluster_name"] = clusterDisplayName
				data["namespace"] = pod.Namespace
				data["pod_name"] = pod.Name
				data["container_name"] = containerStatus.Name
				data["restart_times"] = strconv.Itoa(alert.RestartTimes)
				data["restart_interval"] = strconv.Itoa(alert.RestartIntervalSeconds)

				if details != "" {
					data["logs"] = details
				}

				if err := w.alertManager.SendAlert(data); err != nil {
					logrus.Debugf("Error occurred while getting pod %s: %v", alert.PodName, err)
				}
			}

			return
		}
	}

}

func (w *PodWatcher) getRestartTimeFromTrack(name, namespace string, alert *v3.TargetPod, curCount int32) int32 {

	obj, ok := w.podRestartTrack.Load(namespace + ":" + name)
	if !ok {
		return curCount
	}
	tracks := obj.([]restartTrack)

	now := time.Now()

	if len(tracks) == 0 {
		tracks = append(tracks, restartTrack{Count: curCount, Time: now})
		w.podRestartTrack.Store(namespace+":"+name, tracks)
		return curCount
	}

	for i, track := range tracks {
		if now.Sub(track.Time).Seconds() < float64(alert.RestartIntervalSeconds) {
			tracks = tracks[i:]
			tracks = append(tracks, restartTrack{Count: curCount, Time: now})
			w.podRestartTrack.Store(namespace+":"+name, tracks)
			return track.Count
		}
	}

	w.podRestartTrack.Store(namespace+":"+name, []restartTrack{})
	return curCount
}

func (w *PodWatcher) checkPodRunning(name, namespace string, pod *corev1.Pod, alert *v3.TargetPod) {
	if !w.checkPodScheduled(name, namespace, pod, alert) {
		return
	}

	groupID := namespace + "-" + name
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Running == nil {
			//TODO: need to consider all the cases
			details := ""
			if containerStatus.State.Waiting != nil {
				details = containerStatus.State.Waiting.Message
			}

			if containerStatus.State.Terminated != nil {
				details = containerStatus.State.Terminated.Message
			}

			clusterDisplayName := w.clusterName
			cluster, err := w.clusterLister.Get("", w.clusterName)
			if err != nil {
				logrus.Warnf("Failed to get cluster for %s: %v", w.clusterName, err)
			} else {
				clusterDisplayName = cluster.Spec.DisplayName
			}

			data := map[string]string{}
			data["alert_type"] = "podNotRunning"
			data["group_id"] = groupID
			data["severity"] = alert.Severity
			// data["alert_name"] = alert.Spec.DisplayName //todo
			data["cluster_name"] = clusterDisplayName
			data["namespace"] = pod.Namespace
			data["pod_name"] = pod.Name
			data["container_name"] = containerStatus.Name

			if details != "" {
				data["logs"] = details
			}

			if err := w.alertManager.SendAlert(data); err != nil {
				logrus.Debugf("Error occurred while send alert %s: %v", alert.PodName, err)
			}
			return
		}
	}
}

func (w *PodWatcher) checkPodScheduled(name, namespace string, pod *corev1.Pod, alert *v3.TargetPod) bool {

	groupID := namespace + "-" + name
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
			details := condition.Message

			clusterDisplayName := w.clusterName
			cluster, err := w.clusterLister.Get("", w.clusterName)
			if err != nil {
				logrus.Warnf("Failed to get cluster for %s: %v", w.clusterName, err)
			} else {
				clusterDisplayName = cluster.Spec.DisplayName
			}

			data := map[string]string{}
			data["alert_type"] = "podNotScheduled"
			data["group_id"] = groupID
			data["severity"] = alert.Severity
			// data["alert_name"] = alert.Spec.DisplayName
			data["cluster_name"] = clusterDisplayName
			data["namespace"] = pod.Namespace
			data["pod_name"] = pod.Name

			if details != "" {
				data["logs"] = details
			}

			if err := w.alertManager.SendAlert(data); err != nil {
				logrus.Debugf("Error occurred while getting pod %s: %v", alert.PodName, err)
			}
			return false
		}
	}

	return true

}
