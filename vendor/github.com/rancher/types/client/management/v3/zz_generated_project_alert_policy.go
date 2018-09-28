package client

import (
	"github.com/rancher/norman/types"
)

const (
	ProjectAlertPolicyType                      = "projectAlertPolicy"
	ProjectAlertPolicyFieldAnnotations          = "annotations"
	ProjectAlertPolicyFieldCreated              = "created"
	ProjectAlertPolicyFieldCreatorID            = "creatorId"
	ProjectAlertPolicyFieldDescription          = "description"
	ProjectAlertPolicyFieldDisplayName          = "displayName"
	ProjectAlertPolicyFieldLabels               = "labels"
	ProjectAlertPolicyFieldMetrics              = "metrics"
	ProjectAlertPolicyFieldName                 = "name"
	ProjectAlertPolicyFieldNamespaceId          = "namespaceId"
	ProjectAlertPolicyFieldOwnerReferences      = "ownerReferences"
	ProjectAlertPolicyFieldProjectID            = "projectId"
	ProjectAlertPolicyFieldRecipients           = "recipients"
	ProjectAlertPolicyFieldRemoved              = "removed"
	ProjectAlertPolicyFieldState                = "state"
	ProjectAlertPolicyFieldStatus               = "status"
	ProjectAlertPolicyFieldTargetPods           = "targetPods"
	ProjectAlertPolicyFieldTargetWorkloads      = "targetWorkloads"
	ProjectAlertPolicyFieldTransitioning        = "transitioning"
	ProjectAlertPolicyFieldTransitioningMessage = "transitioningMessage"
	ProjectAlertPolicyFieldUUID                 = "uuid"
)

type ProjectAlertPolicy struct {
	types.Resource
	Annotations          map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Created              string            `json:"created,omitempty" yaml:"created,omitempty"`
	CreatorID            string            `json:"creatorId,omitempty" yaml:"creatorId,omitempty"`
	Description          string            `json:"description,omitempty" yaml:"description,omitempty"`
	DisplayName          string            `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Metrics              []Metric          `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	Name                 string            `json:"name,omitempty" yaml:"name,omitempty"`
	NamespaceId          string            `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	OwnerReferences      []OwnerReference  `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	ProjectID            string            `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	Recipients           []Recipient       `json:"recipients,omitempty" yaml:"recipients,omitempty"`
	Removed              string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	State                string            `json:"state,omitempty" yaml:"state,omitempty"`
	Status               *AlertStatus      `json:"status,omitempty" yaml:"status,omitempty"`
	TargetPods           []TargetPod       `json:"targetPods,omitempty" yaml:"targetPods,omitempty"`
	TargetWorkloads      []TargetWorkload  `json:"targetWorkloads,omitempty" yaml:"targetWorkloads,omitempty"`
	Transitioning        string            `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string            `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                 string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ProjectAlertPolicyCollection struct {
	types.Collection
	Data   []ProjectAlertPolicy `json:"data,omitempty"`
	client *ProjectAlertPolicyClient
}

type ProjectAlertPolicyClient struct {
	apiClient *Client
}

type ProjectAlertPolicyOperations interface {
	List(opts *types.ListOpts) (*ProjectAlertPolicyCollection, error)
	Create(opts *ProjectAlertPolicy) (*ProjectAlertPolicy, error)
	Update(existing *ProjectAlertPolicy, updates interface{}) (*ProjectAlertPolicy, error)
	Replace(existing *ProjectAlertPolicy) (*ProjectAlertPolicy, error)
	ByID(id string) (*ProjectAlertPolicy, error)
	Delete(container *ProjectAlertPolicy) error

	ActionActivate(resource *ProjectAlertPolicy) error

	ActionDeactivate(resource *ProjectAlertPolicy) error

	ActionMute(resource *ProjectAlertPolicy) error

	ActionUnmute(resource *ProjectAlertPolicy) error
}

func newProjectAlertPolicyClient(apiClient *Client) *ProjectAlertPolicyClient {
	return &ProjectAlertPolicyClient{
		apiClient: apiClient,
	}
}

func (c *ProjectAlertPolicyClient) Create(container *ProjectAlertPolicy) (*ProjectAlertPolicy, error) {
	resp := &ProjectAlertPolicy{}
	err := c.apiClient.Ops.DoCreate(ProjectAlertPolicyType, container, resp)
	return resp, err
}

func (c *ProjectAlertPolicyClient) Update(existing *ProjectAlertPolicy, updates interface{}) (*ProjectAlertPolicy, error) {
	resp := &ProjectAlertPolicy{}
	err := c.apiClient.Ops.DoUpdate(ProjectAlertPolicyType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ProjectAlertPolicyClient) Replace(obj *ProjectAlertPolicy) (*ProjectAlertPolicy, error) {
	resp := &ProjectAlertPolicy{}
	err := c.apiClient.Ops.DoReplace(ProjectAlertPolicyType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ProjectAlertPolicyClient) List(opts *types.ListOpts) (*ProjectAlertPolicyCollection, error) {
	resp := &ProjectAlertPolicyCollection{}
	err := c.apiClient.Ops.DoList(ProjectAlertPolicyType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ProjectAlertPolicyCollection) Next() (*ProjectAlertPolicyCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ProjectAlertPolicyCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ProjectAlertPolicyClient) ByID(id string) (*ProjectAlertPolicy, error) {
	resp := &ProjectAlertPolicy{}
	err := c.apiClient.Ops.DoByID(ProjectAlertPolicyType, id, resp)
	return resp, err
}

func (c *ProjectAlertPolicyClient) Delete(container *ProjectAlertPolicy) error {
	return c.apiClient.Ops.DoResourceDelete(ProjectAlertPolicyType, &container.Resource)
}

func (c *ProjectAlertPolicyClient) ActionActivate(resource *ProjectAlertPolicy) error {
	err := c.apiClient.Ops.DoAction(ProjectAlertPolicyType, "activate", &resource.Resource, nil, nil)
	return err
}

func (c *ProjectAlertPolicyClient) ActionDeactivate(resource *ProjectAlertPolicy) error {
	err := c.apiClient.Ops.DoAction(ProjectAlertPolicyType, "deactivate", &resource.Resource, nil, nil)
	return err
}

func (c *ProjectAlertPolicyClient) ActionMute(resource *ProjectAlertPolicy) error {
	err := c.apiClient.Ops.DoAction(ProjectAlertPolicyType, "mute", &resource.Resource, nil, nil)
	return err
}

func (c *ProjectAlertPolicyClient) ActionUnmute(resource *ProjectAlertPolicy) error {
	err := c.apiClient.Ops.DoAction(ProjectAlertPolicyType, "unmute", &resource.Resource, nil, nil)
	return err
}
