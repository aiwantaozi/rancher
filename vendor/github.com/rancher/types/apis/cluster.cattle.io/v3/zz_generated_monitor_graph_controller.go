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
	MonitorGraphGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "MonitorGraph",
	}
	MonitorGraphResource = metav1.APIResource{
		Name:         "monitorgraphs",
		SingularName: "monitorgraph",
		Namespaced:   true,

		Kind: MonitorGraphGroupVersionKind.Kind,
	}
)

type MonitorGraphList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MonitorGraph
}

type MonitorGraphHandlerFunc func(key string, obj *MonitorGraph) (runtime.Object, error)

type MonitorGraphLister interface {
	List(namespace string, selector labels.Selector) (ret []*MonitorGraph, err error)
	Get(namespace, name string) (*MonitorGraph, error)
}

type MonitorGraphController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() MonitorGraphLister
	AddHandler(ctx context.Context, name string, handler MonitorGraphHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler MonitorGraphHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type MonitorGraphInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*MonitorGraph) (*MonitorGraph, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*MonitorGraph, error)
	Get(name string, opts metav1.GetOptions) (*MonitorGraph, error)
	Update(*MonitorGraph) (*MonitorGraph, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*MonitorGraphList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() MonitorGraphController
	AddHandler(ctx context.Context, name string, sync MonitorGraphHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle MonitorGraphLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync MonitorGraphHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle MonitorGraphLifecycle)
}

type monitorGraphLister struct {
	controller *monitorGraphController
}

func (l *monitorGraphLister) List(namespace string, selector labels.Selector) (ret []*MonitorGraph, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*MonitorGraph))
	})
	return
}

func (l *monitorGraphLister) Get(namespace, name string) (*MonitorGraph, error) {
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
			Group:    MonitorGraphGroupVersionKind.Group,
			Resource: "monitorGraph",
		}, key)
	}
	return obj.(*MonitorGraph), nil
}

type monitorGraphController struct {
	controller.GenericController
}

func (c *monitorGraphController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *monitorGraphController) Lister() MonitorGraphLister {
	return &monitorGraphLister{
		controller: c,
	}
}

func (c *monitorGraphController) AddHandler(ctx context.Context, name string, handler MonitorGraphHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*MonitorGraph); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *monitorGraphController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler MonitorGraphHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*MonitorGraph); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type monitorGraphFactory struct {
}

func (c monitorGraphFactory) Object() runtime.Object {
	return &MonitorGraph{}
}

func (c monitorGraphFactory) List() runtime.Object {
	return &MonitorGraphList{}
}

func (s *monitorGraphClient) Controller() MonitorGraphController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.monitorGraphControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(MonitorGraphGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &monitorGraphController{
		GenericController: genericController,
	}

	s.client.monitorGraphControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type monitorGraphClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   MonitorGraphController
}

func (s *monitorGraphClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *monitorGraphClient) Create(o *MonitorGraph) (*MonitorGraph, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*MonitorGraph), err
}

func (s *monitorGraphClient) Get(name string, opts metav1.GetOptions) (*MonitorGraph, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*MonitorGraph), err
}

func (s *monitorGraphClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*MonitorGraph, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*MonitorGraph), err
}

func (s *monitorGraphClient) Update(o *MonitorGraph) (*MonitorGraph, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*MonitorGraph), err
}

func (s *monitorGraphClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *monitorGraphClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *monitorGraphClient) List(opts metav1.ListOptions) (*MonitorGraphList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*MonitorGraphList), err
}

func (s *monitorGraphClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *monitorGraphClient) Patch(o *MonitorGraph, data []byte, subresources ...string) (*MonitorGraph, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*MonitorGraph), err
}

func (s *monitorGraphClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *monitorGraphClient) AddHandler(ctx context.Context, name string, sync MonitorGraphHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *monitorGraphClient) AddLifecycle(ctx context.Context, name string, lifecycle MonitorGraphLifecycle) {
	sync := NewMonitorGraphLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *monitorGraphClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync MonitorGraphHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *monitorGraphClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle MonitorGraphLifecycle) {
	sync := NewMonitorGraphLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}
