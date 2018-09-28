package client

import (
	"github.com/rancher/norman/types"
)

const (
	ClusterAlertPolicyType                      = "clusterAlertPolicy"
	ClusterAlertPolicyFieldAlertState           = "alertState"
	ClusterAlertPolicyFieldAnnotations          = "annotations"
	ClusterAlertPolicyFieldClusterID            = "clusterId"
	ClusterAlertPolicyFieldCreated              = "created"
	ClusterAlertPolicyFieldCreatorID            = "creatorId"
	ClusterAlertPolicyFieldDescription          = "description"
	ClusterAlertPolicyFieldLabels               = "labels"
	ClusterAlertPolicyFieldMetrics              = "metrics"
	ClusterAlertPolicyFieldName                 = "name"
	ClusterAlertPolicyFieldNamespaceId          = "namespaceId"
	ClusterAlertPolicyFieldOwnerReferences      = "ownerReferences"
	ClusterAlertPolicyFieldRecipients           = "recipients"
	ClusterAlertPolicyFieldRemoved              = "removed"
	ClusterAlertPolicyFieldState                = "state"
	ClusterAlertPolicyFieldTargetEvents         = "targetEvents"
	ClusterAlertPolicyFieldTargetSystemServices = "targetSystemServices"
	ClusterAlertPolicyFieldTransitioning        = "transitioning"
	ClusterAlertPolicyFieldTransitioningMessage = "transitioningMessage"
	ClusterAlertPolicyFieldUUID                 = "uuid"
)

type ClusterAlertPolicy struct {
	types.Resource
	AlertState           string                `json:"alertState,omitempty" yaml:"alertState,omitempty"`
	Annotations          map[string]string     `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	ClusterID            string                `json:"clusterId,omitempty" yaml:"clusterId,omitempty"`
	Created              string                `json:"created,omitempty" yaml:"created,omitempty"`
	CreatorID            string                `json:"creatorId,omitempty" yaml:"creatorId,omitempty"`
	Description          string                `json:"description,omitempty" yaml:"description,omitempty"`
	Labels               map[string]string     `json:"labels,omitempty" yaml:"labels,omitempty"`
	Metrics              []Metric              `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	Name                 string                `json:"name,omitempty" yaml:"name,omitempty"`
	NamespaceId          string                `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	OwnerReferences      []OwnerReference      `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	Recipients           []Recipient           `json:"recipients,omitempty" yaml:"recipients,omitempty"`
	Removed              string                `json:"removed,omitempty" yaml:"removed,omitempty"`
	State                string                `json:"state,omitempty" yaml:"state,omitempty"`
	TargetEvents         []TargetEvent         `json:"targetEvents,omitempty" yaml:"targetEvents,omitempty"`
	TargetSystemServices []TargetSystemService `json:"targetSystemServices,omitempty" yaml:"targetSystemServices,omitempty"`
	Transitioning        string                `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string                `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                 string                `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ClusterAlertPolicyCollection struct {
	types.Collection
	Data   []ClusterAlertPolicy `json:"data,omitempty"`
	client *ClusterAlertPolicyClient
}

type ClusterAlertPolicyClient struct {
	apiClient *Client
}

type ClusterAlertPolicyOperations interface {
	List(opts *types.ListOpts) (*ClusterAlertPolicyCollection, error)
	Create(opts *ClusterAlertPolicy) (*ClusterAlertPolicy, error)
	Update(existing *ClusterAlertPolicy, updates interface{}) (*ClusterAlertPolicy, error)
	Replace(existing *ClusterAlertPolicy) (*ClusterAlertPolicy, error)
	ByID(id string) (*ClusterAlertPolicy, error)
	Delete(container *ClusterAlertPolicy) error

	ActionActivate(resource *ClusterAlertPolicy) error

	ActionDeactivate(resource *ClusterAlertPolicy) error

	ActionMute(resource *ClusterAlertPolicy) error

	ActionUnmute(resource *ClusterAlertPolicy) error
}

func newClusterAlertPolicyClient(apiClient *Client) *ClusterAlertPolicyClient {
	return &ClusterAlertPolicyClient{
		apiClient: apiClient,
	}
}

func (c *ClusterAlertPolicyClient) Create(container *ClusterAlertPolicy) (*ClusterAlertPolicy, error) {
	resp := &ClusterAlertPolicy{}
	err := c.apiClient.Ops.DoCreate(ClusterAlertPolicyType, container, resp)
	return resp, err
}

func (c *ClusterAlertPolicyClient) Update(existing *ClusterAlertPolicy, updates interface{}) (*ClusterAlertPolicy, error) {
	resp := &ClusterAlertPolicy{}
	err := c.apiClient.Ops.DoUpdate(ClusterAlertPolicyType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ClusterAlertPolicyClient) Replace(obj *ClusterAlertPolicy) (*ClusterAlertPolicy, error) {
	resp := &ClusterAlertPolicy{}
	err := c.apiClient.Ops.DoReplace(ClusterAlertPolicyType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ClusterAlertPolicyClient) List(opts *types.ListOpts) (*ClusterAlertPolicyCollection, error) {
	resp := &ClusterAlertPolicyCollection{}
	err := c.apiClient.Ops.DoList(ClusterAlertPolicyType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ClusterAlertPolicyCollection) Next() (*ClusterAlertPolicyCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ClusterAlertPolicyCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ClusterAlertPolicyClient) ByID(id string) (*ClusterAlertPolicy, error) {
	resp := &ClusterAlertPolicy{}
	err := c.apiClient.Ops.DoByID(ClusterAlertPolicyType, id, resp)
	return resp, err
}

func (c *ClusterAlertPolicyClient) Delete(container *ClusterAlertPolicy) error {
	return c.apiClient.Ops.DoResourceDelete(ClusterAlertPolicyType, &container.Resource)
}

func (c *ClusterAlertPolicyClient) ActionActivate(resource *ClusterAlertPolicy) error {
	err := c.apiClient.Ops.DoAction(ClusterAlertPolicyType, "activate", &resource.Resource, nil, nil)
	return err
}

func (c *ClusterAlertPolicyClient) ActionDeactivate(resource *ClusterAlertPolicy) error {
	err := c.apiClient.Ops.DoAction(ClusterAlertPolicyType, "deactivate", &resource.Resource, nil, nil)
	return err
}

func (c *ClusterAlertPolicyClient) ActionMute(resource *ClusterAlertPolicy) error {
	err := c.apiClient.Ops.DoAction(ClusterAlertPolicyType, "mute", &resource.Resource, nil, nil)
	return err
}

func (c *ClusterAlertPolicyClient) ActionUnmute(resource *ClusterAlertPolicy) error {
	err := c.apiClient.Ops.DoAction(ClusterAlertPolicyType, "unmute", &resource.Resource, nil, nil)
	return err
}
