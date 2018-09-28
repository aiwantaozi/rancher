package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceMonitorLifecycle interface {
	Create(obj *ServiceMonitor) (*ServiceMonitor, error)
	Remove(obj *ServiceMonitor) (*ServiceMonitor, error)
	Updated(obj *ServiceMonitor) (*ServiceMonitor, error)
}

type serviceMonitorLifecycleAdapter struct {
	lifecycle ServiceMonitorLifecycle
}

func (w *serviceMonitorLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*ServiceMonitor))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *serviceMonitorLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*ServiceMonitor))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *serviceMonitorLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*ServiceMonitor))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewServiceMonitorLifecycleAdapter(name string, clusterScoped bool, client ServiceMonitorInterface, l ServiceMonitorLifecycle) ServiceMonitorHandlerFunc {
	adapter := &serviceMonitorLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *ServiceMonitor) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
