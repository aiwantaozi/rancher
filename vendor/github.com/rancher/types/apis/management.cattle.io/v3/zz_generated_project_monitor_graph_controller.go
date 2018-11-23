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
	ProjectMonitorGraphGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ProjectMonitorGraph",
	}
	ProjectMonitorGraphResource = metav1.APIResource{
		Name:         "projectmonitorgraphs",
		SingularName: "projectmonitorgraph",
		Namespaced:   true,

		Kind: ProjectMonitorGraphGroupVersionKind.Kind,
	}
)

type ProjectMonitorGraphList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectMonitorGraph
}

type ProjectMonitorGraphHandlerFunc func(key string, obj *ProjectMonitorGraph) (runtime.Object, error)

type ProjectMonitorGraphLister interface {
	List(namespace string, selector labels.Selector) (ret []*ProjectMonitorGraph, err error)
	Get(namespace, name string) (*ProjectMonitorGraph, error)
}

type ProjectMonitorGraphController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ProjectMonitorGraphLister
	AddHandler(ctx context.Context, name string, handler ProjectMonitorGraphHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler ProjectMonitorGraphHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ProjectMonitorGraphInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*ProjectMonitorGraph) (*ProjectMonitorGraph, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ProjectMonitorGraph, error)
	Get(name string, opts metav1.GetOptions) (*ProjectMonitorGraph, error)
	Update(*ProjectMonitorGraph) (*ProjectMonitorGraph, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ProjectMonitorGraphList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ProjectMonitorGraphController
	AddHandler(ctx context.Context, name string, sync ProjectMonitorGraphHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle ProjectMonitorGraphLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ProjectMonitorGraphHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ProjectMonitorGraphLifecycle)
}

type projectMonitorGraphLister struct {
	controller *projectMonitorGraphController
}

func (l *projectMonitorGraphLister) List(namespace string, selector labels.Selector) (ret []*ProjectMonitorGraph, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*ProjectMonitorGraph))
	})
	return
}

func (l *projectMonitorGraphLister) Get(namespace, name string) (*ProjectMonitorGraph, error) {
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
			Group:    ProjectMonitorGraphGroupVersionKind.Group,
			Resource: "projectMonitorGraph",
		}, key)
	}
	return obj.(*ProjectMonitorGraph), nil
}

type projectMonitorGraphController struct {
	controller.GenericController
}

func (c *projectMonitorGraphController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *projectMonitorGraphController) Lister() ProjectMonitorGraphLister {
	return &projectMonitorGraphLister{
		controller: c,
	}
}

func (c *projectMonitorGraphController) AddHandler(ctx context.Context, name string, handler ProjectMonitorGraphHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ProjectMonitorGraph); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *projectMonitorGraphController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler ProjectMonitorGraphHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ProjectMonitorGraph); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type projectMonitorGraphFactory struct {
}

func (c projectMonitorGraphFactory) Object() runtime.Object {
	return &ProjectMonitorGraph{}
}

func (c projectMonitorGraphFactory) List() runtime.Object {
	return &ProjectMonitorGraphList{}
}

func (s *projectMonitorGraphClient) Controller() ProjectMonitorGraphController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.projectMonitorGraphControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ProjectMonitorGraphGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &projectMonitorGraphController{
		GenericController: genericController,
	}

	s.client.projectMonitorGraphControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type projectMonitorGraphClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ProjectMonitorGraphController
}

func (s *projectMonitorGraphClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *projectMonitorGraphClient) Create(o *ProjectMonitorGraph) (*ProjectMonitorGraph, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*ProjectMonitorGraph), err
}

func (s *projectMonitorGraphClient) Get(name string, opts metav1.GetOptions) (*ProjectMonitorGraph, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*ProjectMonitorGraph), err
}

func (s *projectMonitorGraphClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ProjectMonitorGraph, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*ProjectMonitorGraph), err
}

func (s *projectMonitorGraphClient) Update(o *ProjectMonitorGraph) (*ProjectMonitorGraph, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*ProjectMonitorGraph), err
}

func (s *projectMonitorGraphClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *projectMonitorGraphClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *projectMonitorGraphClient) List(opts metav1.ListOptions) (*ProjectMonitorGraphList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ProjectMonitorGraphList), err
}

func (s *projectMonitorGraphClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *projectMonitorGraphClient) Patch(o *ProjectMonitorGraph, data []byte, subresources ...string) (*ProjectMonitorGraph, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*ProjectMonitorGraph), err
}

func (s *projectMonitorGraphClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *projectMonitorGraphClient) AddHandler(ctx context.Context, name string, sync ProjectMonitorGraphHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *projectMonitorGraphClient) AddLifecycle(ctx context.Context, name string, lifecycle ProjectMonitorGraphLifecycle) {
	sync := NewProjectMonitorGraphLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *projectMonitorGraphClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ProjectMonitorGraphHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *projectMonitorGraphClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ProjectMonitorGraphLifecycle) {
	sync := NewProjectMonitorGraphLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}
