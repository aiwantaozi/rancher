package alert

import (
	"context"
	"fmt"

	"github.com/rancher/rancher/pkg/controllers/user/alert/configsyncer"
	"github.com/rancher/rancher/pkg/controllers/user/alert/deployer"
	"github.com/rancher/rancher/pkg/controllers/user/alert/manager"
	"github.com/rancher/rancher/pkg/controllers/user/alert/watcher"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	SeverityInfo       = "info"
	SeverityCritical   = "critical"
	SeverityWarning    = "warning"
	defaultTimingField = v3.TimingField{
		GroupWaitSeconds:      180,
		GroupIntervalSeconds:  180,
		RepeatIntervalSeconds: 3600,
	}
)

var (
	ComparisonEqual          = "equal"
	ComparisonNotEqual       = "not-equal"
	ComparisonGreaterThan    = "greater-than"
	ComparisonLessThan       = "less-than"
	ComparisonGreaterOrEqual = "greater-or-equal"
	ComparisonLessOrEqual    = "less-or-equal"
)

func Register(ctx context.Context, cluster *config.UserContext) {
	alertmanager := manager.NewAlertManager(cluster)

	prometheusCRDManager := manager.NewPrometheusCRDManager(ctx, cluster)

	clusterAlertRules := cluster.Management.Management.ClusterAlertRules(cluster.ClusterName)
	projectAlertRules := cluster.Management.Management.ProjectAlertRules("")

	clusterAlertGroups := cluster.Management.Management.ClusterAlertGroups(cluster.ClusterName)
	projectAlertGroups := cluster.Management.Management.ProjectAlertGroups("")

	deploy := deployer.NewDeployer(cluster, alertmanager)
	clusterAlertGroups.AddClusterScopedHandler(ctx, "cluster-alert-deployer", cluster.ClusterName, deploy.ClusterSync)
	projectAlertGroups.AddClusterScopedHandler(ctx, "project-alert-deployer", cluster.ClusterName, deploy.ProjectSync)

	configSyncer := configsyncer.NewConfigSyncer(ctx, cluster, alertmanager, prometheusCRDManager)
	clusterAlertGroups.AddClusterScopedHandler(ctx, "cluster-alert-group-controller", cluster.ClusterName, configSyncer.ClusterGroupSync)
	projectAlertGroups.AddClusterScopedHandler(ctx, "project-alert-group-controller", cluster.ClusterName, configSyncer.ProjectGroupSync)

	clusterAlertRules.AddClusterScopedHandler(ctx, "cluster-alert-rule-controller", cluster.ClusterName, configSyncer.ClusterRuleSync)
	projectAlertRules.AddClusterScopedHandler(ctx, "project-alert-rule-controller", cluster.ClusterName, configSyncer.ProjectRuleSync)

	projects := cluster.Management.Management.Projects("")
	projectLifecycle := &ProjectLifecycle{
		projectAlertRules:  projectAlertRules,
		projectAlertGroups: projectAlertGroups,
		clusterName:        cluster.ClusterName,
	}
	projects.AddClusterScopedLifecycle(ctx, "project-precan-alertpoicy-controller", cluster.ClusterName, projectLifecycle)
	initClusterPreCanAlerts(clusterAlertGroups, clusterAlertRules, cluster.ClusterName)

	watcher.StartEventWatcher(ctx, cluster, alertmanager)
	watcher.StartSysComponentWatcher(ctx, cluster, alertmanager)
	watcher.StartPodWatcher(ctx, cluster, alertmanager)
	watcher.StartWorkloadWatcher(ctx, cluster, alertmanager)
	watcher.StartNodeWatcher(ctx, cluster, alertmanager)

	cleaner := &alertGroupCleaner{
		clusterAlertRules: clusterAlertRules,
		projectAlertRules: projectAlertRules,
	}

	cl := &clusterAlertGroupLifecycle{cleaner: cleaner}
	pl := &projectAlertGroupLifecycle{cleaner: cleaner}
	clusterAlertGroups.AddClusterScopedLifecycle(ctx, "cluster-alert-group-lifecycle", cluster.ClusterName, cl)
	projectAlertGroups.AddClusterScopedLifecycle(ctx, "project-alert-group-lifecycle", cluster.ClusterName, pl)

}

type clusterAlertGroupLifecycle struct {
	cleaner *alertGroupCleaner
}

type projectAlertGroupLifecycle struct {
	cleaner *alertGroupCleaner
}

type alertGroupCleaner struct {
	clusterAlertRules v3.ClusterAlertRuleInterface
	projectAlertRules v3.ProjectAlertRuleInterface
}

func (l *projectAlertGroupLifecycle) Create(obj *v3.ProjectAlertGroup) (runtime.Object, error) {
	return obj, nil
}

func (l *projectAlertGroupLifecycle) Updated(obj *v3.ProjectAlertGroup) (runtime.Object, error) {
	return obj, nil
}

func (l *projectAlertGroupLifecycle) Remove(obj *v3.ProjectAlertGroup) (runtime.Object, error) {
	l.cleaner.Clean(nil, obj)
	return obj, nil
}

func (l *clusterAlertGroupLifecycle) Create(obj *v3.ClusterAlertGroup) (runtime.Object, error) {
	return obj, nil
}

func (l *clusterAlertGroupLifecycle) Updated(obj *v3.ClusterAlertGroup) (runtime.Object, error) {
	return obj, nil
}

func (l *clusterAlertGroupLifecycle) Remove(obj *v3.ClusterAlertGroup) (runtime.Object, error) {
	l.cleaner.Clean(obj, nil)
	return obj, nil
}

func (l *alertGroupCleaner) Clean(clusterGroup *v3.ClusterAlertGroup, projectGroup *v3.ProjectAlertGroup) error {
	if clusterGroup != nil {
		return l.clusterAlertRules.DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.groupName=%s:%s", clusterGroup.Namespace, clusterGroup.Name),
		})
	}

	if projectGroup != nil {
		return l.projectAlertRules.DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.namespace=%s,spec.groupName=%s:%s", projectGroup.Namespace, projectGroup.Namespace, projectGroup.Name),
		})
	}
	return nil
}
