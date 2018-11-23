package servicemonitor

// import (
// 	"context"

// 	"github.com/rancher/types/config"
// )

// func Register(ctx context.Context, workload *config.UserOnlyContext) {
// 	// c := &ServiceController{
// 	// 	pods:            workload.Core.Pods(""),
// 	// 	workloadLister:  util.NewWorkloadController(workload, nil),
// 	// 	podLister:       workload.Core.Pods("").Controller().Lister(),
// 	// 	namespaceLister: workload.Core.Namespaces("").Controller().Lister(),
// 	// 	serviceLister:   workload.Core.Services("").Controller().Lister(),
// 	// 	services:        workload.Core.Services(""),
// 	// }
// 	// p := &PodController{
// 	// 	workloadLister: util.NewWorkloadController(workload, nil),
// 	// 	pods:           workload.Core.Pods(""),
// 	// 	serviceLister:  workload.Core.Services("").Controller().Lister(),
// 	// 	services:       workload.Core.Services(""),
// 	// }
// 	// workload.Core.Services("").AddHandler("workloadServiceController", c.sync)
// 	// workload.Core.Pods("").AddHandler("podToWorkloadServiceController", p.sync)
// }
