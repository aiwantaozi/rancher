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
	ClusterAlertRuleGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ClusterAlertRule",
	}
	ClusterAlertRuleResource = metav1.APIResource{
		Name:         "clusteralertrules",
		SingularName: "clusteralertrule",
		Namespaced:   true,

		Kind: ClusterAlertRuleGroupVersionKind.Kind,
	}
)

type ClusterAlertRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAlertRule
}

type ClusterAlertRuleHandlerFunc func(key string, obj *ClusterAlertRule) (runtime.Object, error)

type ClusterAlertRuleLister interface {
	List(namespace string, selector labels.Selector) (ret []*ClusterAlertRule, err error)
	Get(namespace, name string) (*ClusterAlertRule, error)
}

type ClusterAlertRuleController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ClusterAlertRuleLister
	AddHandler(ctx context.Context, name string, handler ClusterAlertRuleHandlerFunc)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, handler ClusterAlertRuleHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ClusterAlertRuleInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*ClusterAlertRule) (*ClusterAlertRule, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ClusterAlertRule, error)
	Get(name string, opts metav1.GetOptions) (*ClusterAlertRule, error)
	Update(*ClusterAlertRule) (*ClusterAlertRule, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ClusterAlertRuleList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ClusterAlertRuleController
	AddHandler(ctx context.Context, name string, sync ClusterAlertRuleHandlerFunc)
	AddLifecycle(ctx context.Context, name string, lifecycle ClusterAlertRuleLifecycle)
	AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ClusterAlertRuleHandlerFunc)
	AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ClusterAlertRuleLifecycle)
}

type clusterAlertRuleLister struct {
	controller *clusterAlertRuleController
}

func (l *clusterAlertRuleLister) List(namespace string, selector labels.Selector) (ret []*ClusterAlertRule, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*ClusterAlertRule))
	})
	return
}

func (l *clusterAlertRuleLister) Get(namespace, name string) (*ClusterAlertRule, error) {
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
			Group:    ClusterAlertRuleGroupVersionKind.Group,
			Resource: "clusterAlertRule",
		}, key)
	}
	return obj.(*ClusterAlertRule), nil
}

type clusterAlertRuleController struct {
	controller.GenericController
}

func (c *clusterAlertRuleController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *clusterAlertRuleController) Lister() ClusterAlertRuleLister {
	return &clusterAlertRuleLister{
		controller: c,
	}
}

func (c *clusterAlertRuleController) AddHandler(ctx context.Context, name string, handler ClusterAlertRuleHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ClusterAlertRule); ok {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

func (c *clusterAlertRuleController) AddClusterScopedHandler(ctx context.Context, name, cluster string, handler ClusterAlertRuleHandlerFunc) {
	c.GenericController.AddHandler(ctx, name, func(key string, obj interface{}) (interface{}, error) {
		if obj == nil {
			return handler(key, nil)
		} else if v, ok := obj.(*ClusterAlertRule); ok && controller.ObjectInCluster(cluster, obj) {
			return handler(key, v)
		} else {
			return nil, nil
		}
	})
}

type clusterAlertRuleFactory struct {
}

func (c clusterAlertRuleFactory) Object() runtime.Object {
	return &ClusterAlertRule{}
}

func (c clusterAlertRuleFactory) List() runtime.Object {
	return &ClusterAlertRuleList{}
}

func (s *clusterAlertRuleClient) Controller() ClusterAlertRuleController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.clusterAlertRuleControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ClusterAlertRuleGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &clusterAlertRuleController{
		GenericController: genericController,
	}

	s.client.clusterAlertRuleControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type clusterAlertRuleClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ClusterAlertRuleController
}

func (s *clusterAlertRuleClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *clusterAlertRuleClient) Create(o *ClusterAlertRule) (*ClusterAlertRule, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*ClusterAlertRule), err
}

func (s *clusterAlertRuleClient) Get(name string, opts metav1.GetOptions) (*ClusterAlertRule, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*ClusterAlertRule), err
}

func (s *clusterAlertRuleClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ClusterAlertRule, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*ClusterAlertRule), err
}

func (s *clusterAlertRuleClient) Update(o *ClusterAlertRule) (*ClusterAlertRule, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*ClusterAlertRule), err
}

func (s *clusterAlertRuleClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *clusterAlertRuleClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *clusterAlertRuleClient) List(opts metav1.ListOptions) (*ClusterAlertRuleList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ClusterAlertRuleList), err
}

func (s *clusterAlertRuleClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *clusterAlertRuleClient) Patch(o *ClusterAlertRule, data []byte, subresources ...string) (*ClusterAlertRule, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*ClusterAlertRule), err
}

func (s *clusterAlertRuleClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *clusterAlertRuleClient) AddHandler(ctx context.Context, name string, sync ClusterAlertRuleHandlerFunc) {
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *clusterAlertRuleClient) AddLifecycle(ctx context.Context, name string, lifecycle ClusterAlertRuleLifecycle) {
	sync := NewClusterAlertRuleLifecycleAdapter(name, false, s, lifecycle)
	s.Controller().AddHandler(ctx, name, sync)
}

func (s *clusterAlertRuleClient) AddClusterScopedHandler(ctx context.Context, name, clusterName string, sync ClusterAlertRuleHandlerFunc) {
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}

func (s *clusterAlertRuleClient) AddClusterScopedLifecycle(ctx context.Context, name, clusterName string, lifecycle ClusterAlertRuleLifecycle) {
	sync := NewClusterAlertRuleLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.Controller().AddClusterScopedHandler(ctx, name, clusterName, sync)
}
