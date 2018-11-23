package client

import (
	"github.com/rancher/norman/types"
)

const (
	MonitorGraphType                        = "monitorGraph"
	MonitorGraphFieldAnnotations            = "annotations"
	MonitorGraphFieldCreated                = "created"
	MonitorGraphFieldCreatorID              = "creatorId"
	MonitorGraphFieldDescription            = "description"
	MonitorGraphFieldDetailsMetricsSelector = "detailsMetricsSelector"
	MonitorGraphFieldEnable                 = "enable"
	MonitorGraphFieldLabels                 = "labels"
	MonitorGraphFieldMetricsSelector        = "metricsSelector"
	MonitorGraphFieldName                   = "name"
	MonitorGraphFieldNamespaceId            = "namespaceId"
	MonitorGraphFieldOwnerReferences        = "ownerReferences"
	MonitorGraphFieldPriority               = "priority"
	MonitorGraphFieldRemoved                = "removed"
	MonitorGraphFieldState                  = "state"
	MonitorGraphFieldStatus                 = "status"
	MonitorGraphFieldThresholds             = "thresholds"
	MonitorGraphFieldTitle                  = "title"
	MonitorGraphFieldTransitioning          = "transitioning"
	MonitorGraphFieldTransitioningMessage   = "transitioningMessage"
	MonitorGraphFieldUUID                   = "uuid"
	MonitorGraphFieldXAxis                  = "xAxis"
	MonitorGraphFieldYAxis                  = "yAxis"
)

type MonitorGraph struct {
	types.Resource
	Annotations            map[string]string   `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Created                string              `json:"created,omitempty" yaml:"created,omitempty"`
	CreatorID              string              `json:"creatorId,omitempty" yaml:"creatorId,omitempty"`
	Description            string              `json:"description,omitempty" yaml:"description,omitempty"`
	DetailsMetricsSelector map[string]string   `json:"detailsMetricsSelector,omitempty" yaml:"detailsMetricsSelector,omitempty"`
	Enable                 bool                `json:"enable,omitempty" yaml:"enable,omitempty"`
	Labels                 map[string]string   `json:"labels,omitempty" yaml:"labels,omitempty"`
	MetricsSelector        map[string]string   `json:"metricsSelector,omitempty" yaml:"metricsSelector,omitempty"`
	Name                   string              `json:"name,omitempty" yaml:"name,omitempty"`
	NamespaceId            string              `json:"namespaceId,omitempty" yaml:"namespaceId,omitempty"`
	OwnerReferences        []OwnerReference    `json:"ownerReferences,omitempty" yaml:"ownerReferences,omitempty"`
	Priority               int64               `json:"priority,omitempty" yaml:"priority,omitempty"`
	Removed                string              `json:"removed,omitempty" yaml:"removed,omitempty"`
	State                  string              `json:"state,omitempty" yaml:"state,omitempty"`
	Status                 *MonitorGraphStatus `json:"status,omitempty" yaml:"status,omitempty"`
	Thresholds             *float64            `json:"thresholds,omitempty" yaml:"thresholds,omitempty"`
	Title                  string              `json:"title,omitempty" yaml:"title,omitempty"`
	Transitioning          string              `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage   string              `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                   string              `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	XAxis                  *XAxis              `json:"xAxis,omitempty" yaml:"xAxis,omitempty"`
	YAxis                  *YAxis              `json:"yAxis,omitempty" yaml:"yAxis,omitempty"`
}

type MonitorGraphCollection struct {
	types.Collection
	Data   []MonitorGraph `json:"data,omitempty"`
	client *MonitorGraphClient
}

type MonitorGraphClient struct {
	apiClient *Client
}

type MonitorGraphOperations interface {
	List(opts *types.ListOpts) (*MonitorGraphCollection, error)
	Create(opts *MonitorGraph) (*MonitorGraph, error)
	Update(existing *MonitorGraph, updates interface{}) (*MonitorGraph, error)
	Replace(existing *MonitorGraph) (*MonitorGraph, error)
	ByID(id string) (*MonitorGraph, error)
	Delete(container *MonitorGraph) error
}

func newMonitorGraphClient(apiClient *Client) *MonitorGraphClient {
	return &MonitorGraphClient{
		apiClient: apiClient,
	}
}

func (c *MonitorGraphClient) Create(container *MonitorGraph) (*MonitorGraph, error) {
	resp := &MonitorGraph{}
	err := c.apiClient.Ops.DoCreate(MonitorGraphType, container, resp)
	return resp, err
}

func (c *MonitorGraphClient) Update(existing *MonitorGraph, updates interface{}) (*MonitorGraph, error) {
	resp := &MonitorGraph{}
	err := c.apiClient.Ops.DoUpdate(MonitorGraphType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *MonitorGraphClient) Replace(obj *MonitorGraph) (*MonitorGraph, error) {
	resp := &MonitorGraph{}
	err := c.apiClient.Ops.DoReplace(MonitorGraphType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *MonitorGraphClient) List(opts *types.ListOpts) (*MonitorGraphCollection, error) {
	resp := &MonitorGraphCollection{}
	err := c.apiClient.Ops.DoList(MonitorGraphType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *MonitorGraphCollection) Next() (*MonitorGraphCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &MonitorGraphCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *MonitorGraphClient) ByID(id string) (*MonitorGraph, error) {
	resp := &MonitorGraph{}
	err := c.apiClient.Ops.DoByID(MonitorGraphType, id, resp)
	return resp, err
}

func (c *MonitorGraphClient) Delete(container *MonitorGraph) error {
	return c.apiClient.Ops.DoResourceDelete(MonitorGraphType, &container.Resource)
}
