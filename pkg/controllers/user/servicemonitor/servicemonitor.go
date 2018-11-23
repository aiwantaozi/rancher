package servicemonitor

// import (
// 	"context"

// 	util "github.com/rancher/rancher/pkg/controllers/user/workload"
// 	"github.com/rancher/types/apis/core/v1"
// 	managementv3 "github.com/rancher/types/apis/management.cattle.io/v3"
// 	projectv3 "github.com/rancher/types/apis/project.cattle.io/v3"
// 	"github.com/rancher/types/config"
// )

// type ServiceController struct {
// 	serviceLister        v1.ServiceLister
// 	serviceClient        v1.ServiceInterface
// 	serviceMonitorLister projectv3.ServiceMonitorLister
// 	serviceMonitorClient projectv3.ServiceMonitorInterface
// 	nsLister             v1.NamespaceLister
// 	projectLister        managementv3.ProjectLister
// }

// type PodController struct {
// 	pods           v1.PodInterface
// 	workloadLister util.CommonController
// 	serviceLister  v1.ServiceLister
// 	services       v1.ServiceInterface
// }

// func Register(ctx context.Context, workload *config.UserOnlyContext) {
// 	c := &ServiceController{
// 		pods:            workload.Core.Pods(""),
// 		workloadLister:  util.NewWorkloadController(workload, nil),
// 		podLister:       workload.Core.Pods("").Controller().Lister(),
// 		namespaceLister: workload.Core.Namespaces("").Controller().Lister(),
// 		serviceLister:   workload.Core.Services("").Controller().Lister(),
// 		services:        workload.Core.Services(""),
// 	}
// 	p := &PodController{
// 		workloadLister: util.NewWorkloadController(workload, nil),
// 		pods:           workload.Core.Pods(""),
// 		serviceLister:  workload.Core.Services("").Controller().Lister(),
// 		services:       workload.Core.Services(""),
// 	}
// 	workload.Core.Services("").AddHandler("workloadServiceController", c.sync)
// 	workload.Core.Pods("").AddHandler("podToWorkloadServiceController", p.sync)
// }
