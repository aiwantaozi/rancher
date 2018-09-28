package v1

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type AlertmanagerLifecycle interface {
	Create(obj *Alertmanager) (*Alertmanager, error)
	Remove(obj *Alertmanager) (*Alertmanager, error)
	Updated(obj *Alertmanager) (*Alertmanager, error)
}

type alertmanagerLifecycleAdapter struct {
	lifecycle AlertmanagerLifecycle
}

func (w *alertmanagerLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*Alertmanager))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *alertmanagerLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*Alertmanager))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *alertmanagerLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*Alertmanager))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewAlertmanagerLifecycleAdapter(name string, clusterScoped bool, client AlertmanagerInterface, l AlertmanagerLifecycle) AlertmanagerHandlerFunc {
	adapter := &alertmanagerLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *Alertmanager) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
