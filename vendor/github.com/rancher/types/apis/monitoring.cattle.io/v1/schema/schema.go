package schema

import (
	"github.com/rancher/norman/types"
	m "github.com/rancher/norman/types/mapper"
	monitoringv1 "github.com/rancher/types/apis/monitoring.cattle.io/v1"
	"github.com/rancher/types/factory"
)

var (
	Version = types.APIVersion{
		Version: "v1",
		Group:   "monitoring.cattle.io",
		Path:    "/v1",
	}

	Schemas = factory.Schemas(&Version).
		Init(Constructor(&Version, false))
)

func Constructor(version *types.APIVersion, withoutController bool) func(schemas *types.Schemas) *types.Schemas {
	return func(schemas *types.Schemas) *types.Schemas {
		return schemas.
			AddMapperForType(version, monitoringv1.StorageSpec{},
				&m.Drop{Field: "class"},
				&m.Drop{Field: "selector"},
				&m.Drop{Field: "resources"},
			).
			AddMapperForType(version, monitoringv1.Prometheus{},
				&m.Drop{Field: "status"},
				&m.AnnotationField{Field: "description"},
			).
			AddMapperForType(version, monitoringv1.PrometheusSpec{},
				&m.Drop{Field: "thanos"},
				&m.Drop{Field: "apiserverConfig"},
				&m.Drop{Field: "serviceMonitorNamespaceSelector"},
				&m.Drop{Field: "ruleNamespaceSelector"},
				&m.Drop{Field: "paused"},
				&m.Enum{
					Field: "logLevel",
					Options: []string{
						"all",
						"debug",
						"info",
						"warn",
						"error",
						"none",
					},
				},
			).
			AddMapperForType(version, monitoringv1.Alertmanager{},
				&m.Drop{Field: "status"},
			).
			MustImportAndCustomize(version, monitoringv1.Prometheus{}, func(schema *types.Schema) {
				if withoutController {
					schema.ControllerExcluded = true
				}
				schema.MustCustomizeField("name", func(field types.Field) types.Field {
					field.Type = "dnsLabelRestricted"
					field.Nullable = false
					field.Required = true
					return field
				})
			}, projectOverride{}, struct {
				Description string `json:"description"`
			}{}).
			MustImportAndCustomize(version, monitoringv1.PrometheusRule{}, func(schema *types.Schema) {
				if withoutController {
					schema.ControllerExcluded = true
				}
			}, projectOverride{}).
			MustImportAndCustomize(version, monitoringv1.Alertmanager{}, func(schema *types.Schema) {
				if withoutController {
					schema.ControllerExcluded = true
				}
			}, projectOverride{}).
			MustImportAndCustomize(version, monitoringv1.ServiceMonitor{}, func(schema *types.Schema) {
				if withoutController {
					schema.ControllerExcluded = true
				}
			}, projectOverride{})
	}
}
