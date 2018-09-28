package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type PrometheusLifecycle interface {
	Create(obj *Prometheus) (*Prometheus, error)
	Remove(obj *Prometheus) (*Prometheus, error)
	Updated(obj *Prometheus) (*Prometheus, error)
}

type prometheusLifecycleAdapter struct {
	lifecycle PrometheusLifecycle
}

func (w *prometheusLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*Prometheus))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *prometheusLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*Prometheus))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *prometheusLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*Prometheus))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewPrometheusLifecycleAdapter(name string, clusterScoped bool, client PrometheusInterface, l PrometheusLifecycle) PrometheusHandlerFunc {
	adapter := &prometheusLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *Prometheus) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
