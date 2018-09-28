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
	ProjectAlertPolicyGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "ProjectAlertPolicy",
	}
	ProjectAlertPolicyResource = metav1.APIResource{
		Name:         "projectalertpolicies",
		SingularName: "projectalertpolicy",
		Namespaced:   true,

		Kind: ProjectAlertPolicyGroupVersionKind.Kind,
	}
)

type ProjectAlertPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectAlertPolicy
}

type ProjectAlertPolicyHandlerFunc func(key string, obj *ProjectAlertPolicy) error

type ProjectAlertPolicyLister interface {
	List(namespace string, selector labels.Selector) (ret []*ProjectAlertPolicy, err error)
	Get(namespace, name string) (*ProjectAlertPolicy, error)
}

type ProjectAlertPolicyController interface {
	Generic() controller.GenericController
	Informer() cache.SharedIndexInformer
	Lister() ProjectAlertPolicyLister
	AddHandler(name string, handler ProjectAlertPolicyHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler ProjectAlertPolicyHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type ProjectAlertPolicyInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*ProjectAlertPolicy) (*ProjectAlertPolicy, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ProjectAlertPolicy, error)
	Get(name string, opts metav1.GetOptions) (*ProjectAlertPolicy, error)
	Update(*ProjectAlertPolicy) (*ProjectAlertPolicy, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*ProjectAlertPolicyList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() ProjectAlertPolicyController
	AddHandler(name string, sync ProjectAlertPolicyHandlerFunc)
	AddLifecycle(name string, lifecycle ProjectAlertPolicyLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync ProjectAlertPolicyHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle ProjectAlertPolicyLifecycle)
}

type projectAlertPolicyLister struct {
	controller *projectAlertPolicyController
}

func (l *projectAlertPolicyLister) List(namespace string, selector labels.Selector) (ret []*ProjectAlertPolicy, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*ProjectAlertPolicy))
	})
	return
}

func (l *projectAlertPolicyLister) Get(namespace, name string) (*ProjectAlertPolicy, error) {
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
			Group:    ProjectAlertPolicyGroupVersionKind.Group,
			Resource: "projectAlertPolicy",
		}, key)
	}
	return obj.(*ProjectAlertPolicy), nil
}

type projectAlertPolicyController struct {
	controller.GenericController
}

func (c *projectAlertPolicyController) Generic() controller.GenericController {
	return c.GenericController
}

func (c *projectAlertPolicyController) Lister() ProjectAlertPolicyLister {
	return &projectAlertPolicyLister{
		controller: c,
	}
}

func (c *projectAlertPolicyController) AddHandler(name string, handler ProjectAlertPolicyHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*ProjectAlertPolicy))
	})
}

func (c *projectAlertPolicyController) AddClusterScopedHandler(name, cluster string, handler ProjectAlertPolicyHandlerFunc) {
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

		return handler(key, obj.(*ProjectAlertPolicy))
	})
}

type projectAlertPolicyFactory struct {
}

func (c projectAlertPolicyFactory) Object() runtime.Object {
	return &ProjectAlertPolicy{}
}

func (c projectAlertPolicyFactory) List() runtime.Object {
	return &ProjectAlertPolicyList{}
}

func (s *projectAlertPolicyClient) Controller() ProjectAlertPolicyController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.projectAlertPolicyControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(ProjectAlertPolicyGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &projectAlertPolicyController{
		GenericController: genericController,
	}

	s.client.projectAlertPolicyControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type projectAlertPolicyClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   ProjectAlertPolicyController
}

func (s *projectAlertPolicyClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *projectAlertPolicyClient) Create(o *ProjectAlertPolicy) (*ProjectAlertPolicy, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*ProjectAlertPolicy), err
}

func (s *projectAlertPolicyClient) Get(name string, opts metav1.GetOptions) (*ProjectAlertPolicy, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*ProjectAlertPolicy), err
}

func (s *projectAlertPolicyClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*ProjectAlertPolicy, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*ProjectAlertPolicy), err
}

func (s *projectAlertPolicyClient) Update(o *ProjectAlertPolicy) (*ProjectAlertPolicy, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*ProjectAlertPolicy), err
}

func (s *projectAlertPolicyClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *projectAlertPolicyClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *projectAlertPolicyClient) List(opts metav1.ListOptions) (*ProjectAlertPolicyList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*ProjectAlertPolicyList), err
}

func (s *projectAlertPolicyClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *projectAlertPolicyClient) Patch(o *ProjectAlertPolicy, data []byte, subresources ...string) (*ProjectAlertPolicy, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*ProjectAlertPolicy), err
}

func (s *projectAlertPolicyClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *projectAlertPolicyClient) AddHandler(name string, sync ProjectAlertPolicyHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *projectAlertPolicyClient) AddLifecycle(name string, lifecycle ProjectAlertPolicyLifecycle) {
	sync := NewProjectAlertPolicyLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *projectAlertPolicyClient) AddClusterScopedHandler(name, clusterName string, sync ProjectAlertPolicyHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *projectAlertPolicyClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle ProjectAlertPolicyLifecycle) {
	sync := NewProjectAlertPolicyLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
