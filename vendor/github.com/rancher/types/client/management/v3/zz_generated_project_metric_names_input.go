package client

const (
	ProjectMetricNamesInputType           = "projectMetricNamesInput"
	ProjectMetricNamesInputFieldProjectID = "projectId"
)

type ProjectMetricNamesInput struct {
	ProjectID string `json:"projectId,omitempty" yaml:"projectId,omitempty"`
}
