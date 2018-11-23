package client

import (
	"github.com/rancher/norman/types"
)

const (
	ProjectMonitorGraphType                        = "projectMonitorGraph"
	ProjectMonitorGraphFieldAnnotations            = "annotations"
	ProjectMonitorGraphFieldCreated                = "created"
	ProjectMonitorGraphFieldCreatorID              = "creatorId"
	ProjectMonitorGraphFieldDescription            = "description"
	ProjectMonitorGraphFieldDetailsMetricsSelector = "detailsMetricsSelector"
	ProjectMonitorGraphFieldEnable                 = "enable"
	ProjectMonitorGraphFieldLabels                 = "labels"
	ProjectMonitorGraphFieldMetricsSelector        = "metricsSelector"
	ProjectMonitorGraphFieldName                   = "name"
	ProjectMonitorGraphFieldNamespaceId            = "namespaceId"
	ProjectMonitorGraphFieldOwnerReferences        = "ownerReferences"
	ProjectMonitorGraphFieldPriority               = "priority"
	ProjectMonitorGraphFieldRemoved                = "removed"
	ProjectMonitorGraphFieldThresholds             = "thresholds"
	ProjectMonitorGraphFieldTitle                  = "title"
	ProjectMonitorGraphFieldType                   = "type"
	ProjectMonitorGraphFieldUUID                   = "uuid"
	ProjectMonitorGraphFieldXAxis                  = "xAxis"
	ProjectMonitorGraphFieldYAxis                  = "yAxis"
)

type ProjectMonitorGraph struct {
	types.Resource
	Annotations            map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Created                string            `json:"created,omitempty" yaml:"created,omitempty"`
	CreatorID              string            `json:"creatorId,omitempty" yaml:"creatorId,omitempty"`
	Description            string            `json:"description,omitempty" yaml:"description,omitempty"`
	DetailsMetricsSelector map[string]string `json:"detailsMetricsSelector,omitempty" yaml:"detailsMetricsSelector,omitempty"`
	Enable                 bool              `json:"enable,omitempty" yaml:"enable,omitempty"`
	Labels                 map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	MetricsSelector        map[string]string `json:"metricsSelector,omitempty" yaml:"metricsSelector,omitempty"`
	Name                   string            `json:"name,omitempty" yaml:"name,omitempty"`
	NamespaceId            string            `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	OwnerReferences        []OwnerReference  `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	Priority               int64             `json:"priority,omitempty" yaml:"priority,omitempty"`
	Removed                string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	Thresholds             *float64          `json:"thresholds,omitempty" yaml:"thresholds,omitempty"`
	Title                  string            `json:"title,omitempty" yaml:"title,omitempty"`
	Type                   string            `json:"type,omitempty" yaml:"type,omitempty"`
	UUID                   string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	XAxis                  *XAxis            `json:"xAxis,omitempty" yaml:"xAxis,omitempty"`
	YAxis                  *YAxis            `json:"yAxis,omitempty" yaml:"yAxis,omitempty"`
}

type ProjectMonitorGraphCollection struct {
	types.Collection
	Data   []ProjectMonitorGraph `json:"data,omitempty"`
	client *ProjectMonitorGraphClient
}

type ProjectMonitorGraphClient struct {
	apiClient *Client
}

type ProjectMonitorGraphOperations interface {
	List(opts *types.ListOpts) (*ProjectMonitorGraphCollection, error)
	Create(opts *ProjectMonitorGraph) (*ProjectMonitorGraph, error)
	Update(existing *ProjectMonitorGraph, updates interface{}) (*ProjectMonitorGraph, error)
	Replace(existing *ProjectMonitorGraph) (*ProjectMonitorGraph, error)
	ByID(id string) (*ProjectMonitorGraph, error)
	Delete(container *ProjectMonitorGraph) error

	ActionQuery(resource *ProjectMonitorGraph, input *QueryGraphInput) (*QueryGraphOutput, error)

	CollectionActionQuery(resource *ProjectMonitorGraphCollection, input *QueryGraphInput) (*QueryGraphOutputCollection, error)
}

func newProjectMonitorGraphClient(apiClient *Client) *ProjectMonitorGraphClient {
	return &ProjectMonitorGraphClient{
		apiClient: apiClient,
	}
}

func (c *ProjectMonitorGraphClient) Create(container *ProjectMonitorGraph) (*ProjectMonitorGraph, error) {
	resp := &ProjectMonitorGraph{}
	err := c.apiClient.Ops.DoCreate(ProjectMonitorGraphType, container, resp)
	return resp, err
}

func (c *ProjectMonitorGraphClient) Update(existing *ProjectMonitorGraph, updates interface{}) (*ProjectMonitorGraph, error) {
	resp := &ProjectMonitorGraph{}
	err := c.apiClient.Ops.DoUpdate(ProjectMonitorGraphType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ProjectMonitorGraphClient) Replace(obj *ProjectMonitorGraph) (*ProjectMonitorGraph, error) {
	resp := &ProjectMonitorGraph{}
	err := c.apiClient.Ops.DoReplace(ProjectMonitorGraphType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ProjectMonitorGraphClient) List(opts *types.ListOpts) (*ProjectMonitorGraphCollection, error) {
	resp := &ProjectMonitorGraphCollection{}
	err := c.apiClient.Ops.DoList(ProjectMonitorGraphType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ProjectMonitorGraphCollection) Next() (*ProjectMonitorGraphCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ProjectMonitorGraphCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ProjectMonitorGraphClient) ByID(id string) (*ProjectMonitorGraph, error) {
	resp := &ProjectMonitorGraph{}
	err := c.apiClient.Ops.DoByID(ProjectMonitorGraphType, id, resp)
	return resp, err
}

func (c *ProjectMonitorGraphClient) Delete(container *ProjectMonitorGraph) error {
	return c.apiClient.Ops.DoResourceDelete(ProjectMonitorGraphType, &container.Resource)
}

func (c *ProjectMonitorGraphClient) ActionQuery(resource *ProjectMonitorGraph, input *QueryGraphInput) (*QueryGraphOutput, error) {
	resp := &QueryGraphOutput{}
	err := c.apiClient.Ops.DoAction(ProjectMonitorGraphType, "query", &resource.Resource, input, resp)
	return resp, err
}

func (c *ProjectMonitorGraphClient) CollectionActionQuery(resource *ProjectMonitorGraphCollection, input *QueryGraphInput) (*QueryGraphOutputCollection, error) {
	resp := &QueryGraphOutputCollection{}
	err := c.apiClient.Ops.DoCollectionAction(ProjectMonitorGraphType, "query", &resource.Collection, input, resp)
	return resp, err
}
