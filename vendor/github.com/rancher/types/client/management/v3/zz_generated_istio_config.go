package client

import (
	"github.com/rancher/norman/types"
)

const (
	IstioConfigType                      = "istioConfig"
	IstioConfigFieldAnnotations          = "annotations"
	IstioConfigFieldAppAnswers           = "answers"
	IstioConfigFieldClusterID            = "clusterId"
	IstioConfigFieldCreated              = "created"
	IstioConfigFieldCreatorID            = "creatorId"
	IstioConfigFieldEnable               = "enable"
	IstioConfigFieldLabels               = "labels"
	IstioConfigFieldName                 = "name"
	IstioConfigFieldNamespaceId          = "namespaceId"
	IstioConfigFieldOwnerReferences      = "ownerReferences"
	IstioConfigFieldRemoved              = "removed"
	IstioConfigFieldState                = "state"
	IstioConfigFieldStatus               = "status"
	IstioConfigFieldTransitioning        = "transitioning"
	IstioConfigFieldTransitioningMessage = "transitioningMessage"
	IstioConfigFieldUUID                 = "uuid"
)

type IstioConfig struct {
	types.Resource
	Annotations          map[string]string  `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	AppAnswers           map[string]string  `json:"answers,omitempty" yaml:"answers,omitempty"`
	ClusterID            string             `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	Created              string             `json:"created,omitempty" yaml:"created,omitempty"`
	CreatorID            string             `json:"creatorId,omitempty" yaml:"creatorId,omitempty"`
	Enable               bool               `json:"enable,omitempty" yaml:"enable,omitempty"`
	Labels               map[string]string  `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name                 string             `json:"name,omitempty" yaml:"name,omitempty"`
	NamespaceId          string             `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	OwnerReferences      []OwnerReference   `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	Removed              string             `json:"removed,omitempty" yaml:"removed,omitempty"`
	State                string             `json:"state,omitempty" yaml:"state,omitempty"`
	Status               *IstioConfigStatus `json:"status,omitempty" yaml:"status,omitempty"`
	Transitioning        string             `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string             `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                 string             `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type IstioConfigCollection struct {
	types.Collection
	Data   []IstioConfig `json:"data,omitempty"`
	client *IstioConfigClient
}

type IstioConfigClient struct {
	apiClient *Client
}

type IstioConfigOperations interface {
	List(opts *types.ListOpts) (*IstioConfigCollection, error)
	Create(opts *IstioConfig) (*IstioConfig, error)
	Update(existing *IstioConfig, updates interface{}) (*IstioConfig, error)
	Replace(existing *IstioConfig) (*IstioConfig, error)
	ByID(id string) (*IstioConfig, error)
	Delete(container *IstioConfig) error
}

func newIstioConfigClient(apiClient *Client) *IstioConfigClient {
	return &IstioConfigClient{
		apiClient: apiClient,
	}
}

func (c *IstioConfigClient) Create(container *IstioConfig) (*IstioConfig, error) {
	resp := &IstioConfig{}
	err := c.apiClient.Ops.DoCreate(IstioConfigType, container, resp)
	return resp, err
}

func (c *IstioConfigClient) Update(existing *IstioConfig, updates interface{}) (*IstioConfig, error) {
	resp := &IstioConfig{}
	err := c.apiClient.Ops.DoUpdate(IstioConfigType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *IstioConfigClient) Replace(obj *IstioConfig) (*IstioConfig, error) {
	resp := &IstioConfig{}
	err := c.apiClient.Ops.DoReplace(IstioConfigType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *IstioConfigClient) List(opts *types.ListOpts) (*IstioConfigCollection, error) {
	resp := &IstioConfigCollection{}
	err := c.apiClient.Ops.DoList(IstioConfigType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *IstioConfigCollection) Next() (*IstioConfigCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &IstioConfigCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *IstioConfigClient) ByID(id string) (*IstioConfig, error) {
	resp := &IstioConfig{}
	err := c.apiClient.Ops.DoByID(IstioConfigType, id, resp)
	return resp, err
}

func (c *IstioConfigClient) Delete(container *IstioConfig) error {
	return c.apiClient.Ops.DoResourceDelete(IstioConfigType, &container.Resource)
}
