package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type PrometheusRuleLifecycle interface {
	Create(obj *PrometheusRule) (*PrometheusRule, error)
	Remove(obj *PrometheusRule) (*PrometheusRule, error)
	Updated(obj *PrometheusRule) (*PrometheusRule, error)
}

type prometheusRuleLifecycleAdapter struct {
	lifecycle PrometheusRuleLifecycle
}

func (w *prometheusRuleLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*PrometheusRule))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *prometheusRuleLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*PrometheusRule))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *prometheusRuleLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*PrometheusRule))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewPrometheusRuleLifecycleAdapter(name string, clusterScoped bool, client PrometheusRuleInterface, l PrometheusRuleLifecycle) PrometheusRuleHandlerFunc {
	adapter := &prometheusRuleLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *PrometheusRule) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
