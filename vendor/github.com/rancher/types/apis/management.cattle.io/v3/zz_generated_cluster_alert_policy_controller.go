package v3

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	ClusterAlertPolicyGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ClusterAlertPolicy",
	}
	ClusterAlertPolicyResource = metav1.APIResource{
		Name:         "clusteralertpolicies",
		SingularName: "clusteralertpolicy",
		Namespaced:   true,

		Kind: ClusterAlertPolicyGroupVersionKind.Kind,
	}
)

type ClusterAlertPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAlertPolicy
}

type ClusterAlertPolicyHandlerFunc func(key string, obj *ClusterAlertPolicy) error

type ClusterAlertPolicyLister interface {
	List(namespace string, selector labels.Selector) (ret []*ClusterAlertPolicy, err error)
	Get(namespace, name string) (*ClusterAlertPolicy, error)
}

type ClusterAlertPolicyController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ClusterAlertPolicyLister
	AddHandler(name string, handler ClusterAlertPolicyHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler ClusterAlertPolicyHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ClusterAlertPolicyInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*ClusterAlertPolicy) (*ClusterAlertPolicy, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ClusterAlertPolicy, error)
	Get(name string, opts metav1.GetOptions) (*ClusterAlertPolicy, error)
	Update(*ClusterAlertPolicy) (*ClusterAlertPolicy, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ClusterAlertPolicyList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ClusterAlertPolicyController
	AddHandler(name string, sync ClusterAlertPolicyHandlerFunc)
	AddLifecycle(name string, lifecycle ClusterAlertPolicyLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync ClusterAlertPolicyHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle ClusterAlertPolicyLifecycle)
}

type clusterAlertPolicyLister struct {
	controller *clusterAlertPolicyController
}

func (l *clusterAlertPolicyLister) List(namespace string, selector labels.Selector) (ret []*ClusterAlertPolicy, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*ClusterAlertPolicy))
	})
	return
}

func (l *clusterAlertPolicyLister) Get(namespace, name string) (*ClusterAlertPolicy, error) {
	var key string
	if namespace != "" {
		key = namespace + "/" + name
	} else {
		key = name
	}
	obj, exists, err := l.controller.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    ClusterAlertPolicyGroupVersionKind.Group,
			Resource: "clusterAlertPolicy",
		}, key)
	}
	return obj.(*ClusterAlertPolicy), nil
}

type clusterAlertPolicyController struct {
	controller.GenericController
}

func (c *clusterAlertPolicyController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *clusterAlertPolicyController) Lister() ClusterAlertPolicyLister {
	return &clusterAlertPolicyLister{
		controller: c,
	}
}

func (c *clusterAlertPolicyController) AddHandler(name string, handler ClusterAlertPolicyHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*ClusterAlertPolicy))
	})
}

func (c *clusterAlertPolicyController) AddClusterScopedHandler(name, cluster string, handler ClusterAlertPolicyHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}

		if !controller.ObjectInCluster(cluster, obj) {
			return nil
		}

		return handler(key, obj.(*ClusterAlertPolicy))
	})
}

type clusterAlertPolicyFactory struct {
}

func (c clusterAlertPolicyFactory) Object() runtime.Object {
	return &ClusterAlertPolicy{}
}

func (c clusterAlertPolicyFactory) List() runtime.Object {
	return &ClusterAlertPolicyList{}
}

func (s *clusterAlertPolicyClient) Controller() ClusterAlertPolicyController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.clusterAlertPolicyControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ClusterAlertPolicyGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &clusterAlertPolicyController{
		GenericController: genericController,
	}

	s.client.clusterAlertPolicyControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type clusterAlertPolicyClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ClusterAlertPolicyController
}

func (s *clusterAlertPolicyClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *clusterAlertPolicyClient) Create(o *ClusterAlertPolicy) (*ClusterAlertPolicy, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*ClusterAlertPolicy), err
}

func (s *clusterAlertPolicyClient) Get(name string, opts metav1.GetOptions) (*ClusterAlertPolicy, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*ClusterAlertPolicy), err
}

func (s *clusterAlertPolicyClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ClusterAlertPolicy, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*ClusterAlertPolicy), err
}

func (s *clusterAlertPolicyClient) Update(o *ClusterAlertPolicy) (*ClusterAlertPolicy, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*ClusterAlertPolicy), err
}

func (s *clusterAlertPolicyClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *clusterAlertPolicyClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *clusterAlertPolicyClient) List(opts metav1.ListOptions) (*ClusterAlertPolicyList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ClusterAlertPolicyList), err
}

func (s *clusterAlertPolicyClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *clusterAlertPolicyClient) Patch(o *ClusterAlertPolicy, data []byte, subresources ...string) (*ClusterAlertPolicy, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*ClusterAlertPolicy), err
}

func (s *clusterAlertPolicyClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *clusterAlertPolicyClient) AddHandler(name string, sync ClusterAlertPolicyHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *clusterAlertPolicyClient) AddLifecycle(name string, lifecycle ClusterAlertPolicyLifecycle) {
	sync := NewClusterAlertPolicyLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *clusterAlertPolicyClient) AddClusterScopedHandler(name, clusterName string, sync ClusterAlertPolicyHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *clusterAlertPolicyClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle ClusterAlertPolicyLifecycle) {
	sync := NewClusterAlertPolicyLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
