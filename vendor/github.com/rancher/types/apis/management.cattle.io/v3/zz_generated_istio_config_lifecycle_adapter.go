package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type IstioConfigLifecycle interface {
	Create(obj *IstioConfig) (runtime.Object, error)
	Remove(obj *IstioConfig) (runtime.Object, error)
	Updated(obj *IstioConfig) (runtime.Object, error)
}

type istioConfigLifecycleAdapter struct {
	lifecycle IstioConfigLifecycle
}

func (w *istioConfigLifecycleAdapter) HasCreate() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasCreate()
}

func (w *istioConfigLifecycleAdapter) HasFinalize() bool {
	o, ok := w.lifecycle.(lifecycle.ObjectLifecycleCondition)
	return !ok || o.HasFinalize()
}

func (w *istioConfigLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*IstioConfig))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *istioConfigLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*IstioConfig))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *istioConfigLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*IstioConfig))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewIstioConfigLifecycleAdapter(name string, clusterScoped bool, client IstioConfigInterface, l IstioConfigLifecycle) IstioConfigHandlerFunc {
	adapter := &istioConfigLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *IstioConfig) (runtime.Object, error) {
		newObj, err := syncFn(key, obj)
		if o, ok := newObj.(runtime.Object); ok {
			return o, err
		}
		return nil, err
	}
}
