package schema

import (
	"github.com/rancher/norman/types"
)

type projectOverride struct {
	types.Namespaced
	ProjectID string `norman:"type=reference[/v3/schemas/project],noupdate"`
}
