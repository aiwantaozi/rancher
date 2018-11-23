package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type MonitorGraphLifecycle interface {
	Create(obj *MonitorGraph) (runtime.Object, error)
	Remove(obj *MonitorGraph) (runtime.Object, error)
	Updated(obj *MonitorGraph) (runtime.Object, error)
}

type monitorGraphLifecycleAdapter struct {
	lifecycle MonitorGraphLifecycle
}

func (w *monitorGraphLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*MonitorGraph))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *monitorGraphLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*MonitorGraph))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *monitorGraphLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*MonitorGraph))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewMonitorGraphLifecycleAdapter(name string, clusterScoped bool, client MonitorGraphInterface, l MonitorGraphLifecycle) MonitorGraphHandlerFunc {
	adapter := &monitorGraphLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *MonitorGraph) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
