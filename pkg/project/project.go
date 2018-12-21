package project

const (
	System  = "System"
	Default = "Default"
)

const (
	SystemImageVersionAnn = "field.cattle.io/systemImageVersion"
	ProjectIDAnn          = "field.cattle.io/projectId"
)

var (
	SystemProjectLabel = map[string]string{"authz.management.cattle.io/system-project": "true"}
)
