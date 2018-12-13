package watcher

import (
	"context"
	"time"

	"github.com/rancher/rancher/pkg/controllers/user/logging/utils"
	"github.com/rancher/rancher/pkg/ticker"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	"github.com/rancher/types/config/dialer"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type endpointWatcher struct {
	dialerFactory   dialer.Factory
	clusterName     string
	clusterLoggings mgmtv3.ClusterLoggingInterface
	projectLoggings mgmtv3.ProjectLoggingInterface
}

func StartEndpointWatcher(ctx context.Context, cluster *config.UserContext) {
	s := &endpointWatcher{
		dialerFactory:   cluster.Management.Dialer,
		clusterName:     cluster.ClusterName,
		clusterLoggings: cluster.Management.Management.ClusterLoggings(cluster.ClusterName),
		projectLoggings: cluster.Management.Management.ProjectLoggings(metav1.NamespaceAll),
	}
	go s.watch(ctx, 10*time.Second)
}

func (e *endpointWatcher) watch(ctx context.Context, interval time.Duration) {
	for range ticker.Context(ctx, interval) {
		if err := e.checkTarget(); err != nil {
			logrus.Error(err)
		}
	}
}

func (e *endpointWatcher) checkTarget() error {
	cls, err := e.clusterLoggings.Controller().Lister().List(e.clusterName, labels.NewSelector())
	if err != nil {
		return errors.Wrapf(err, "list clusterlogging fail in endpoint watcher")
	}
	if len(cls) == 0 {
		return nil
	}
	obj := cls[0]

	clusterDialer, err := e.dialerFactory.ClusterDialer(obj.Spec.ClusterName)
	if err != nil {
		return errors.Wrapf(err, "get cluster dailer %s failed", obj.Spec.ClusterName)
	}

	wl := utils.NewWrapLogging(obj.Spec.ElasticsearchConfig, obj.Spec.SplunkConfig, obj.Spec.SyslogConfig, obj.Spec.KafkaConfig, obj.Spec.FluentForwarderConfig)
	err = wl.GetLoggingTarget().TestReachable(clusterDialer)
	updatedObj := setClusterLoggingErrMsg(obj, err)

	_, updateErr := e.clusterLoggings.Update(updatedObj)
	if updateErr != errors.Wrapf(updateErr, "set clusterlogging fail in watch endpoint") {
		return updateErr
	}

	pls, err := e.projectLoggings.Controller().Lister().List(metav1.NamespaceAll, labels.NewSelector())
	if err != nil {
		return errors.Wrapf(err, "list clusterlogging fail in endpoint watcher")
	}

	for _, v := range pls {
		wp := utils.NewWrapLogging(v.Spec.ElasticsearchConfig, v.Spec.SplunkConfig, v.Spec.SyslogConfig, v.Spec.KafkaConfig, v.Spec.FluentForwarderConfig)
		err = wp.GetLoggingTarget().TestReachable(clusterDialer)
		updatedObj := setProjectLoggingErrMsg(v, err)
		_, updateErr := e.projectLoggings.Update(updatedObj)
		if updateErr != errors.Wrapf(updateErr, "set project fail in watch endpoint") {
			return updateErr
		}
	}

	return nil
}

func setProjectLoggingErrMsg(obj *mgmtv3.ProjectLogging, err error) *mgmtv3.ProjectLogging {
	updatedObj := obj.DeepCopy()
	if err != nil {
		mgmtv3.LoggingConditionUpdated.False(updatedObj)
		mgmtv3.LoggingConditionUpdated.Message(updatedObj, err.Error())
		return updatedObj
	}

	mgmtv3.LoggingConditionUpdated.True(updatedObj)
	mgmtv3.LoggingConditionUpdated.Message(updatedObj, "")
	return updatedObj
}

func setClusterLoggingErrMsg(obj *mgmtv3.ClusterLogging, err error) *mgmtv3.ClusterLogging {
	updatedObj := obj.DeepCopy()
	if err != nil {
		updatedObj.Status.FailedSpec = &obj.Spec
		mgmtv3.LoggingConditionUpdated.False(updatedObj)
		mgmtv3.LoggingConditionUpdated.Message(updatedObj, err.Error())
		return updatedObj
	}

	updatedObj.Status.FailedSpec = nil
	updatedObj.Status.AppliedSpec = obj.Spec

	mgmtv3.LoggingConditionUpdated.True(updatedObj)
	mgmtv3.LoggingConditionUpdated.Message(updatedObj, "")
	return updatedObj
}
