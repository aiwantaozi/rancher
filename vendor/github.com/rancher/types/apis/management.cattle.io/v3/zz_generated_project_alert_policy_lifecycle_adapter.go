package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ProjectAlertPolicyLifecycle interface {
	Create(obj *ProjectAlertPolicy) (*ProjectAlertPolicy, error)
	Remove(obj *ProjectAlertPolicy) (*ProjectAlertPolicy, error)
	Updated(obj *ProjectAlertPolicy) (*ProjectAlertPolicy, error)
}

type projectAlertPolicyLifecycleAdapter struct {
	lifecycle ProjectAlertPolicyLifecycle
}

func (w *projectAlertPolicyLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*ProjectAlertPolicy))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *projectAlertPolicyLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*ProjectAlertPolicy))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *projectAlertPolicyLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*ProjectAlertPolicy))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewProjectAlertPolicyLifecycleAdapter(name string, clusterScoped bool, client ProjectAlertPolicyInterface, l ProjectAlertPolicyLifecycle) ProjectAlertPolicyHandlerFunc {
	adapter := &projectAlertPolicyLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *ProjectAlertPolicy) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
