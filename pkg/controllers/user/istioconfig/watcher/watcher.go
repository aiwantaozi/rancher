package watcher

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/rancher/norman/condition"
	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/pkg/controllers/user/app"
	istiocommon "github.com/rancher/rancher/pkg/controllers/user/istioconfig/common"
	"github.com/rancher/rancher/pkg/project"
	"github.com/rancher/rancher/pkg/ticker"
	mgmtv3 "github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	deployedSuccess = "success"
)

type appWatcher struct {
	appDeployer   *app.AppDeployer
	clusterName   string
	projectLister mgmtv3.ProjectLister
	istioConfigs  mgmtv3.IstioConfigInterface
}

func StartAppWatcher(ctx context.Context, cluster *config.UserContext) {
	appDeployer := app.NewAppDeployer(cluster)
	s := &appWatcher{
		appDeployer:   appDeployer,
		clusterName:   cluster.ClusterName,
		istioConfigs:  cluster.Management.Management.IstioConfigs(cluster.ClusterName),
		projectLister: cluster.Management.Management.Projects(metav1.NamespaceAll).Controller().Lister(),
	}
	go s.watch(ctx, 10*time.Second)
}

func (e *appWatcher) watch(ctx context.Context, interval time.Duration) {
	for range ticker.Context(ctx, interval) {
		istioConfigs, err := e.istioConfigs.Controller().Lister().List(e.clusterName, labels.NewSelector())
		if err != nil {
			logrus.Errorf("watcher list istioConfigs in cluster %s failed, %v", e.clusterName, err)
			return
		}

		if len(istioConfigs) == 0 {
			continue
		}

		if len(istioConfigs) > 1 {
			logrus.Errorf("watcher get %v istioConfigs, not equal to 1", fmt.Sprint(len(istioConfigs)))
		}

		istioConfig := istioConfigs[0].DeepCopy()
		istioConfig.Status = mgmtv3.IstioConfigStatus{} //unset status because old status may include previous installed status, last time certmanager is installed, but not installed this time

		if istioConfig.Spec.Enable {
			appErr := e.checkApp(istioConfig)
			podsErr := e.checkPods(istioConfig)
			if appErr == nil && podsErr == nil {
				istioConfig.Status.AppliedSpec = istioConfig.Spec
			} else {
				istioConfig.Status.AppliedSpec.Enable = false
			}
		} else {
			istioConfig.Status.AppliedSpec = istioConfig.Spec
		}

		current, err := e.istioConfigs.Get(istioConfig.Name, metav1.GetOptions{})
		if err != nil {
			logrus.Errorf("watcher list istioConfig in cluster %s failed, %v", e.clusterName, err)
			return
		}

		istioConfig.Status.Conditions = copyConditions(current.Status.Conditions, istioConfig.Status.Conditions)
		if reflect.DeepEqual(current.Status, istioConfig.Status) {
			continue
		}

		cpIstioConfig := current.DeepCopy()
		cpIstioConfig.Status = istioConfig.Status
		if _, err = e.istioConfigs.Update(cpIstioConfig); err != nil {
			logrus.Errorf("watcher update istioConfig %s:%s failed, %v", cpIstioConfig.Namespace, cpIstioConfig.Name, err)
		}
	}
}

func (e *appWatcher) checkApp(obj *mgmtv3.IstioConfig) error {
	systemProject, err := project.GetSystemProject(e.clusterName, e.projectLister)
	if err != nil {
		mgmtv3.IstioConditionAppInstalled.False(obj)
		mgmtv3.IstioConditionAppInstalled.Message(obj, err.Error())
		return err
	}

	if err := e.appDeployer.IsAppDeploySuccess(istiocommon.AppName, systemProject.Name); err != nil {
		mgmtv3.IstioConditionAppInstalled.False(obj)
		mgmtv3.IstioConditionAppInstalled.Message(obj, err.Error())
		return err
	}

	mgmtv3.IstioConditionAppInstalled.True(obj)
	mgmtv3.IstioConditionAppInstalled.Message(obj, "")
	return nil
}

func (e *appWatcher) checkPods(obj *mgmtv3.IstioConfig) error {
	wg := sync.WaitGroup{}
	deployedStatusMap := sync.Map{}
	namespace := istiocommon.IstioDeployedNamespace

	componentSelectorMap := buildComponentSelectorMap(obj.Spec.AppAnswers)
	componentSelectorMap.Range(func(k, v interface{}) bool {
		wg.Add(1)

		condStr := fmt.Sprintf("%v", k)
		componentName := strings.TrimSuffix(condStr, "Deployed")
		selector, ok := v.(map[string]string)
		if !ok {
			return false
		}

		go func() {
			defer wg.Done()
			err := e.appDeployer.IsPodDeploySuccess(namespace, selector)
			if err != nil {
				errMsg := fmt.Sprintf("checking %s deployed status failed, %v", componentName, err)
				deployedStatusMap.Store(k, errMsg)
			} else {
				deployedStatusMap.Store(k, deployedSuccess)
			}
		}()
		return true
	})

	wg.Wait()

	return e.updateIstioConditions(&deployedStatusMap, obj)
}

func (e *appWatcher) updateIstioConditions(condistionStatus *sync.Map, obj *mgmtv3.IstioConfig) error {
	multiErrs := types.Errors{}
	condistionStatus.Range(func(k, v interface{}) bool {
		cond, ok := k.(condition.Cond)
		if !ok {
			return false
		}

		message, ok := v.(string)
		if !ok {
			return false
		}

		if message == deployedSuccess {
			cond.True(obj)
			cond.Message(obj, "")
			return true
		}

		cond.False(obj)
		cond.Message(obj, message)
		multiErrs.Add(errors.New(message))
		return true
	})

	return multiErrs.Err()
}

func buildComponentSelectorMap(answers map[string]string) *sync.Map {
	componentSelectorMap := &sync.Map{}
	for _, v := range istiocommon.AllIstioComponents {
		answerKey := fmt.Sprintf("%s.enabled", v.Name)
		if enabled, ok := answers[answerKey]; ok && enabled == "true" {
			componentSelectorMap.Store(v.Cond, v.Selector)
		}
	}
	return componentSelectorMap
}

func copyConditions(current []mgmtv3.Condition, new []mgmtv3.Condition) []mgmtv3.Condition {
	var cpConditions []mgmtv3.Condition
	for _, n := range new {
		cond := n
		for _, c := range current {
			if c.Type == n.Type {
				cond.LastUpdateTime = c.LastUpdateTime
				cond.LastTransitionTime = c.LastTransitionTime
			}
		}
		cpConditions = append(cpConditions, cond)
	}

	return cpConditions
}
