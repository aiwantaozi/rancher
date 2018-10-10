package alertpolicy

import (
	"context"

	"github.com/rancher/rancher/pkg/controllers/user/alertpolicy/configsyncer"
	"github.com/rancher/rancher/pkg/controllers/user/alertpolicy/deployer"
	"github.com/rancher/rancher/pkg/controllers/user/alertpolicy/manager"
	"github.com/rancher/rancher/pkg/controllers/user/alertpolicy/watcher"
	"github.com/rancher/types/config"
)

var (
	PromCRDNamespace = "cattle-test"
)

func Register(ctx context.Context, cluster *config.UserContext) {
	alertmanager := manager.NewAlertManager(cluster)
	prometheusCRDManager := manager.NewPrometheusCRDManager(cluster)
	clusterAlertPolicies := cluster.Management.Management.ClusterAlertPolicies(cluster.ClusterName)
	projectAlertPolicies := cluster.Management.Management.ProjectAlertPolicies("")

	deployer := deployer.NewDeployer(cluster, alertmanager)
	clusterAlertPolicies.AddClusterScopedHandler("cluster-alert-policies-deployer", cluster.ClusterName, deployer.ClusterSync)
	projectAlertPolicies.AddClusterScopedHandler("project-alert-policies-deployer", cluster.ClusterName, deployer.ProjectSync)

	configSyncer := configsyner.NewConfigSyncer(ctx, cluster, alertmanager, prometheusCRDManager)

	clusterAlertPolicies.AddClusterScopedHandler("cluster-alert-policies-controller", cluster.ClusterName, configSyncer.ClusterSync)
	projectAlertPolicies.AddClusterScopedHandler("project-alert-policies-controller", cluster.ClusterName, configSyncer.ProjectSync)

	watcher.StartEventWatcher(cluster, alertmanager)
	watcher.StartSysComponentWatcher(ctx, cluster, alertmanager)
	watcher.StartPodWatcher(ctx, cluster, alertmanager)
	watcher.StartWorkloadWatcher(ctx, cluster, alertmanager)
}
