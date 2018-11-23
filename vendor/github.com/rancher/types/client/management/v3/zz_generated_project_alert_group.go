package client

import (
	"github.com/rancher/norman/types"
)

const (
	ProjectAlertGroupType                       = "projectAlertGroup"
	ProjectAlertGroupFieldAPIVersion            = "apiVersion"
	ProjectAlertGroupFieldAnnotations           = "annotations"
	ProjectAlertGroupFieldCode                  = "code"
	ProjectAlertGroupFieldCreated               = "created"
	ProjectAlertGroupFieldCreatorID             = "creatorId"
	ProjectAlertGroupFieldDescription           = "description"
	ProjectAlertGroupFieldDetails               = "details"
	ProjectAlertGroupFieldGroupIntervalSeconds  = "groupIntervalSeconds"
	ProjectAlertGroupFieldGroupWaitSeconds      = "groupWaitSeconds"
	ProjectAlertGroupFieldKind                  = "kind"
	ProjectAlertGroupFieldLabels                = "labels"
	ProjectAlertGroupFieldListMeta              = "metadata"
	ProjectAlertGroupFieldMessage               = "message"
	ProjectAlertGroupFieldName                  = "name"
	ProjectAlertGroupFieldNamespaceId           = "namespaceId"
	ProjectAlertGroupFieldOwnerReferences       = "ownerReferences"
	ProjectAlertGroupFieldProjectID             = "projectId"
	ProjectAlertGroupFieldProjectRules          = "projectRule"
	ProjectAlertGroupFieldReason                = "reason"
	ProjectAlertGroupFieldRecipients            = "recipients"
	ProjectAlertGroupFieldRemoved               = "removed"
	ProjectAlertGroupFieldRepeatIntervalSeconds = "repeatIntervalSeconds"
	ProjectAlertGroupFieldState                 = "state"
	ProjectAlertGroupFieldStatus                = "status"
	ProjectAlertGroupFieldTransitioning         = "transitioning"
	ProjectAlertGroupFieldTransitioningMessage  = "transitioningMessage"
	ProjectAlertGroupFieldUUID                  = "uuid"
)

type ProjectAlertGroup struct {
	types.Resource
	APIVersion            string             `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Annotations           map[string]string  `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Code                  int64              `json:"code,omitempty" yaml:"code,omitempty"`
	Created               string             `json:"created,omitempty" yaml:"created,omitempty"`
	CreatorID             string             `json:"creatorId,omitempty" yaml:"creatorId,omitempty"`
	Description           string             `json:"description,omitempty" yaml:"description,omitempty"`
	Details               *StatusDetails     `json:"details,omitempty" yaml:"details,omitempty"`
	GroupIntervalSeconds  int64              `json:"groupIntervalSeconds,omitempty" yaml:"groupIntervalSeconds,omitempty"`
	GroupWaitSeconds      int64              `json:"groupWaitSeconds,omitempty" yaml:"groupWaitSeconds,omitempty"`
	Kind                  string             `json:"kind,omitempty" yaml:"kind,omitempty"`
	Labels                map[string]string  `json:"labels,omitempty" yaml:"labels,omitempty"`
	ListMeta              *ListMeta          `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Message               string             `json:"message,omitempty" yaml:"message,omitempty"`
	Name                  string             `json:"name,omitempty" yaml:"name,omitempty"`
	NamespaceId           string             `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	OwnerReferences       []OwnerReference   `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	ProjectID             string             `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	ProjectRules          []ProjectAlertRule `json:"projectRule,omitempty" yaml:"projectRule,omitempty"`
	Reason                string             `json:"reason,omitempty" yaml:"reason,omitempty"`
	Recipients            []Recipient        `json:"recipients,omitempty" yaml:"recipients,omitempty"`
	Removed               string             `json:"removed,omitempty" yaml:"removed,omitempty"`
	RepeatIntervalSeconds int64              `json:"repeatIntervalSeconds,omitempty" yaml:"repeatIntervalSeconds,omitempty"`
	State                 string             `json:"state,omitempty" yaml:"state,omitempty"`
	Status                string             `json:"status,omitempty" yaml:"status,omitempty"`
	Transitioning         string             `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage  string             `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                  string             `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ProjectAlertGroupCollection struct {
	types.Collection
	Data   []ProjectAlertGroup `json:"data,omitempty"`
	client *ProjectAlertGroupClient
}

type ProjectAlertGroupClient struct {
	apiClient *Client
}

type ProjectAlertGroupOperations interface {
	List(opts *types.ListOpts) (*ProjectAlertGroupCollection, error)
	Create(opts *ProjectAlertGroup) (*ProjectAlertGroup, error)
	Update(existing *ProjectAlertGroup, updates interface{}) (*ProjectAlertGroup, error)
	Replace(existing *ProjectAlertGroup) (*ProjectAlertGroup, error)
	ByID(id string) (*ProjectAlertGroup, error)
	Delete(container *ProjectAlertGroup) error

	ActionActivate(resource *ProjectAlertGroup) error

	ActionDeactivate(resource *ProjectAlertGroup) error

	ActionMute(resource *ProjectAlertGroup) error

	ActionUnmute(resource *ProjectAlertGroup) error
}

func newProjectAlertGroupClient(apiClient *Client) *ProjectAlertGroupClient {
	return &ProjectAlertGroupClient{
		apiClient: apiClient,
	}
}

func (c *ProjectAlertGroupClient) Create(container *ProjectAlertGroup) (*ProjectAlertGroup, error) {
	resp := &ProjectAlertGroup{}
	err := c.apiClient.Ops.DoCreate(ProjectAlertGroupType, container, resp)
	return resp, err
}

func (c *ProjectAlertGroupClient) Update(existing *ProjectAlertGroup, updates interface{}) (*ProjectAlertGroup, error) {
	resp := &ProjectAlertGroup{}
	err := c.apiClient.Ops.DoUpdate(ProjectAlertGroupType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ProjectAlertGroupClient) Replace(obj *ProjectAlertGroup) (*ProjectAlertGroup, error) {
	resp := &ProjectAlertGroup{}
	err := c.apiClient.Ops.DoReplace(ProjectAlertGroupType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ProjectAlertGroupClient) List(opts *types.ListOpts) (*ProjectAlertGroupCollection, error) {
	resp := &ProjectAlertGroupCollection{}
	err := c.apiClient.Ops.DoList(ProjectAlertGroupType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ProjectAlertGroupCollection) Next() (*ProjectAlertGroupCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ProjectAlertGroupCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ProjectAlertGroupClient) ByID(id string) (*ProjectAlertGroup, error) {
	resp := &ProjectAlertGroup{}
	err := c.apiClient.Ops.DoByID(ProjectAlertGroupType, id, resp)
	return resp, err
}

func (c *ProjectAlertGroupClient) Delete(container *ProjectAlertGroup) error {
	return c.apiClient.Ops.DoResourceDelete(ProjectAlertGroupType, &container.Resource)
}

func (c *ProjectAlertGroupClient) ActionActivate(resource *ProjectAlertGroup) error {
	err := c.apiClient.Ops.DoAction(ProjectAlertGroupType, "activate", &resource.Resource, nil, nil)
	return err
}

func (c *ProjectAlertGroupClient) ActionDeactivate(resource *ProjectAlertGroup) error {
	err := c.apiClient.Ops.DoAction(ProjectAlertGroupType, "deactivate", &resource.Resource, nil, nil)
	return err
}

func (c *ProjectAlertGroupClient) ActionMute(resource *ProjectAlertGroup) error {
	err := c.apiClient.Ops.DoAction(ProjectAlertGroupType, "mute", &resource.Resource, nil, nil)
	return err
}

func (c *ProjectAlertGroupClient) ActionUnmute(resource *ProjectAlertGroup) error {
	err := c.apiClient.Ops.DoAction(ProjectAlertGroupType, "unmute", &resource.Resource, nil, nil)
	return err
}
