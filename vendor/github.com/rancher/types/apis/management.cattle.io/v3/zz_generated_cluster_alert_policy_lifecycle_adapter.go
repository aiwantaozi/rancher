package v3

import (
	"github.com/rancher/norman/lifecycle"
	"k8s.io/apimachinery/pkg/runtime"
)

type ClusterAlertPolicyLifecycle interface {
	Create(obj *ClusterAlertPolicy) (*ClusterAlertPolicy, error)
	Remove(obj *ClusterAlertPolicy) (*ClusterAlertPolicy, error)
	Updated(obj *ClusterAlertPolicy) (*ClusterAlertPolicy, error)
}

type clusterAlertPolicyLifecycleAdapter struct {
	lifecycle ClusterAlertPolicyLifecycle
}

func (w *clusterAlertPolicyLifecycleAdapter) Create(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Create(obj.(*ClusterAlertPolicy))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *clusterAlertPolicyLifecycleAdapter) Finalize(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Remove(obj.(*ClusterAlertPolicy))
	if o == nil {
		return nil, err
	}
	return o, err
}

func (w *clusterAlertPolicyLifecycleAdapter) Updated(obj runtime.Object) (runtime.Object, error) {
	o, err := w.lifecycle.Updated(obj.(*ClusterAlertPolicy))
	if o == nil {
		return nil, err
	}
	return o, err
}

func NewClusterAlertPolicyLifecycleAdapter(name string, clusterScoped bool, client ClusterAlertPolicyInterface, l ClusterAlertPolicyLifecycle) ClusterAlertPolicyHandlerFunc {
	adapter := &clusterAlertPolicyLifecycleAdapter{lifecycle: l}
	syncFn := lifecycle.NewObjectLifecycleAdapter(name, clusterScoped, adapter, client.ObjectClient())
	return func(key string, obj *ClusterAlertPolicy) error {
		if obj == nil {
			return syncFn(key, nil)
		}
		return syncFn(key, obj)
	}
}
