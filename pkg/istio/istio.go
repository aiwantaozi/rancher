package istio

import (
	"github.com/rancher/norman/types"
)

var (
	APIVersion = types.APIVersion{
		Version: "v1alpha3",
		Group:   "networking.istio.io",
		Path:    "/v3/project",
	}
)
