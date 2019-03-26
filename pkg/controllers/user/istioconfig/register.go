package istioconfig

import (
	"context"

	"github.com/rancher/rancher/pkg/controllers/user/istioconfig/deployer"
	"github.com/rancher/rancher/pkg/controllers/user/istioconfig/watcher"
	"github.com/rancher/types/config"
)

/*
The deployer watching istioConfig crd and control the istio app deployment.
*/

func Register(ctx context.Context, cluster *config.UserContext) {
	clusterName := cluster.ClusterName
	istioConfig := cluster.Management.Management.IstioConfigs(clusterName)

	deployer := deployer.NewDeployer(cluster)
	istioConfig.AddClusterScopedHandler(ctx, "istio-deployer", cluster.ClusterName, deployer.IstioConfigSync)

	watcher.StartAppWatcher(ctx, cluster)
}
