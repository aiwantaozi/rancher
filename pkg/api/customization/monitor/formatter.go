package monitor

import (
	"github.com/rancher/norman/types"
)

const (
	ResourceNode              = "node"
	ResourceCluster           = "cluster"
	ResourceEtcd              = "etcd"
	ResourceAPIServer         = "apiserver"
	ResourceScheduler         = "scheduler"
	ResourceControllerManager = "controllermanager"
	ResourceIngressController = "ingressController"
	ResourceFluentd           = "fluentd"
	ResourceWorkload          = "workload"
	ResourcePod               = "pod"
	ResourceContainer         = "container"
)

const (
	queryAction           = "query"
	listMetricNameAction  = "listmetricname"
	querycluster          = "querycluster"
	queryproject          = "queryproject"
	listclustermetricname = "listclustermetricname"
	listprojectmetricname = "listprojectmetricname"
)

func MonitorGraphFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	resource.AddAction(apiContext, queryAction)
}

func QueryGraphCollectionFormatter(apiContext *types.APIContext, collection *types.GenericCollection) {
	collection.AddAction(apiContext, queryAction)
}

func MetricFormatter(apiContext *types.APIContext, resource *types.RawResource) {
	resource.AddAction(apiContext, queryAction)
	resource.AddAction(apiContext, listMetricNameAction)
}

func MetricCollectionFormatter(apiContext *types.APIContext, collection *types.GenericCollection) {
	collection.AddAction(apiContext, queryAction)
	collection.AddAction(apiContext, listMetricNameAction)
}
