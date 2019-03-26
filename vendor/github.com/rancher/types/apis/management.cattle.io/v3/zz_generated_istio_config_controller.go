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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	IstioConfigGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "IstioConfig",
	}
	IstioConfigResource = metav1.APIResource{
		Name:         "istioconfigs",
		SingularName: "istioconfig",
		Namespaced:   true,

		Kind: IstioConfigGroupVersionKind.Kind,
	}
)

func NewIstioConfig(namespace, name string, obj IstioConfig) *IstioConfig {
	obj.APIVersion, obj.Kind = IstioConfigGroupVersionKind.ToAPIVersionAndKind()
	obj.Name = name
	obj.Namespace = namespace
	return &obj
}

type IstioConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IstioConfig
}

type IstioConfigHandlerFunc func(key string, obj *IstioConfig) (runtime.Object, error)

type IstioConfigChangeHandlerFunc func(obj *IstioConfig) (runtime.Object, error)

type IstioConfigLister interface {
	List(namespace string, selector labels.Selector) (ret []*IstioConfig, err error)
	Get(namespace, name string) (*IstioConfig, error)
}

type IstioConfigController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() IstioConfigLister
	AddHandler(ctx context.Context, name string, handler IstioConfigHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler IstioConfigHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type IstioConfigInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*IstioConfig) (*IstioConfig, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*IstioConfig, error)
	Get(name string, opts metav1.GetOptions) (*IstioConfig, error)
	Update(*IstioConfig) (*IstioConfig, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*IstioConfigList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() IstioConfigController
	AddHandler(ctx context.Context, name string, sync IstioConfigHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle IstioConfigLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync IstioConfigHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle IstioConfigLifecycle)
}

type istioConfigLister struct {
	controller *istioConfigController
}

func (l *istioConfigLister) List(namespace string, selector labels.Selector) (ret []*IstioConfig, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*IstioConfig))
	})
	return
}

func (l *istioConfigLister) Get(namespace, name string) (*IstioConfig, error) {
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
			Group:    IstioConfigGroupVersionKind.Group,
			Resource: "istioConfig",
		}, key)
	}
	return obj.(*IstioConfig), nil
}

type istioConfigController struct {
	controller.GenericController
}

func (c *istioConfigController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *istioConfigController) Lister() IstioConfigLister {
	return &istioConfigLister{
		controller: c,
	}
}

func (c *istioConfigController) AddHandler(ctx context.Context, name string, handler IstioConfigHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*IstioConfig); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *istioConfigController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler IstioConfigHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*IstioConfig); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type istioConfigFactory struct {
}

func (c istioConfigFactory) Object() runtime.Object {
	return &IstioConfig{}
}

func (c istioConfigFactory) List() runtime.Object {
	return &IstioConfigList{}
}

func (s *istioConfigClient) Controller() IstioConfigController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.istioConfigControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(IstioConfigGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &istioConfigController{
		GenericController: genericController,
	}

	s.client.istioConfigControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type istioConfigClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   IstioConfigController
}

func (s *istioConfigClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *istioConfigClient) Create(o *IstioConfig) (*IstioConfig, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*IstioConfig), err
}

func (s *istioConfigClient) Get(name string, opts metav1.GetOptions) (*IstioConfig, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*IstioConfig), err
}

func (s *istioConfigClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*IstioConfig, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*IstioConfig), err
}

func (s *istioConfigClient) Update(o *IstioConfig) (*IstioConfig, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*IstioConfig), err
}

func (s *istioConfigClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *istioConfigClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *istioConfigClient) List(opts metav1.ListOptions) (*IstioConfigList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*IstioConfigList), err
}

func (s *istioConfigClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *istioConfigClient) Patch(o *IstioConfig, patchType types.PatchType, data []byte, subresources ...string) (*IstioConfig, error) {
	obj, err := s.objectClient.Patch(o.Name, o, patchType, data, subresources...)
	return obj.(*IstioConfig), err
}

func (s *istioConfigClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *istioConfigClient) AddHandler(ctx context.Context, name string, sync IstioConfigHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *istioConfigClient) AddLifecycle(ctx context.Context, name string, lifecycle IstioConfigLifecycle) {
	sync := NewIstioConfigLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *istioConfigClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync IstioConfigHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *istioConfigClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle IstioConfigLifecycle) {
	sync := NewIstioConfigLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

type IstioConfigIndexer func(obj *IstioConfig) ([]string, error)

type IstioConfigClientCache interface {
	Get(namespace, name string) (*IstioConfig, error)
	List(namespace string, selector labels.Selector) ([]*IstioConfig, error)

	Index(name string, indexer IstioConfigIndexer)
	GetIndexed(name, key string) ([]*IstioConfig, error)
}

type IstioConfigClient interface {
	Create(*IstioConfig) (*IstioConfig, error)
	Get(namespace, name string, opts metav1.GetOptions) (*IstioConfig, error)
	Update(*IstioConfig) (*IstioConfig, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	List(namespace string, opts metav1.ListOptions) (*IstioConfigList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)

	Cache() IstioConfigClientCache

	OnCreate(ctx context.Context, name string, sync IstioConfigChangeHandlerFunc)
	OnChange(ctx context.Context, name string, sync IstioConfigChangeHandlerFunc)
	OnRemove(ctx context.Context, name string, sync IstioConfigChangeHandlerFunc)
	Enqueue(namespace, name string)

	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
	Interface() IstioConfigInterface
}

type istioConfigClientCache struct {
	client *istioConfigClient2
}

type istioConfigClient2 struct {
	iface      IstioConfigInterface
	controller IstioConfigController
}

func (n *istioConfigClient2) Interface() IstioConfigInterface {
	return n.iface
}

func (n *istioConfigClient2) Generic() controller.GenericController {
	return n.iface.Controller().Generic()
}

func (n *istioConfigClient2) ObjectClient() *objectclient.ObjectClient {
	return n.Interface().ObjectClient()
}

func (n *istioConfigClient2) Enqueue(namespace, name string) {
	n.iface.Controller().Enqueue(namespace, name)
}

func (n *istioConfigClient2) Create(obj *IstioConfig) (*IstioConfig, error) {
	return n.iface.Create(obj)
}

func (n *istioConfigClient2) Get(namespace, name string, opts metav1.GetOptions) (*IstioConfig, error) {
	return n.iface.GetNamespaced(namespace, name, opts)
}

func (n *istioConfigClient2) Update(obj *IstioConfig) (*IstioConfig, error) {
	return n.iface.Update(obj)
}

func (n *istioConfigClient2) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	return n.iface.DeleteNamespaced(namespace, name, options)
}

func (n *istioConfigClient2) List(namespace string, opts metav1.ListOptions) (*IstioConfigList, error) {
	return n.iface.List(opts)
}

func (n *istioConfigClient2) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return n.iface.Watch(opts)
}

func (n *istioConfigClientCache) Get(namespace, name string) (*IstioConfig, error) {
	return n.client.controller.Lister().Get(namespace, name)
}

func (n *istioConfigClientCache) List(namespace string, selector labels.Selector) ([]*IstioConfig, error) {
	return n.client.controller.Lister().List(namespace, selector)
}

func (n *istioConfigClient2) Cache() IstioConfigClientCache {
	n.loadController()
	return &istioConfigClientCache{
		client: n,
	}
}

func (n *istioConfigClient2) OnCreate(ctx context.Context, name string, sync IstioConfigChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-create", &istioConfigLifecycleDelegate{create: sync})
}

func (n *istioConfigClient2) OnChange(ctx context.Context, name string, sync IstioConfigChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name+"-change", &istioConfigLifecycleDelegate{update: sync})
}

func (n *istioConfigClient2) OnRemove(ctx context.Context, name string, sync IstioConfigChangeHandlerFunc) {
	n.loadController()
	n.iface.AddLifecycle(ctx, name, &istioConfigLifecycleDelegate{remove: sync})
}

func (n *istioConfigClientCache) Index(name string, indexer IstioConfigIndexer) {
	err := n.client.controller.Informer().GetIndexer().AddIndexers(map[string]cache.IndexFunc{
		name: func(obj interface{}) ([]string, error) {
			if v, ok := obj.(*IstioConfig); ok {
				return indexer(v)
			}
			return nil, nil
		},
	})

	if err != nil {
		panic(err)
	}
}

func (n *istioConfigClientCache) GetIndexed(name, key string) ([]*IstioConfig, error) {
	var result []*IstioConfig
	objs, err := n.client.controller.Informer().GetIndexer().ByIndex(name, key)
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if v, ok := obj.(*IstioConfig); ok {
			result = append(result, v)
		}
	}

	return result, nil
}

func (n *istioConfigClient2) loadController() {
	if n.controller == nil {
		n.controller = n.iface.Controller()
	}
}

type istioConfigLifecycleDelegate struct {
	create IstioConfigChangeHandlerFunc
	update IstioConfigChangeHandlerFunc
	remove IstioConfigChangeHandlerFunc
}

func (n *istioConfigLifecycleDelegate) HasCreate() bool {
	return n.create != nil
}

func (n *istioConfigLifecycleDelegate) Create(obj *IstioConfig) (runtime.Object, error) {
	if n.create == nil {
		return obj, nil
	}
	return n.create(obj)
}

func (n *istioConfigLifecycleDelegate) HasFinalize() bool {
	return n.remove != nil
}

func (n *istioConfigLifecycleDelegate) Remove(obj *IstioConfig) (runtime.Object, error) {
	if n.remove == nil {
		return obj, nil
	}
	return n.remove(obj)
}

func (n *istioConfigLifecycleDelegate) Updated(obj *IstioConfig) (runtime.Object, error) {
	if n.update == nil {
		return obj, nil
	}
	return n.update(obj)
}
