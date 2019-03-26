package istioconfig

import (
	"net/http"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/types"
	v3client "github.com/rancher/types/client/management/v3"

	"github.com/pkg/errors"
)

func Validator(resquest *types.APIContext, schema *types.Schema, data map[string]interface{}) error {
	if resquest.Method != http.MethodPost {
		return nil
	}

	var currentConfigs []v3client.IstioConfig
	if err := access.List(resquest, resquest.Version, v3client.IstioConfigType, &types.QueryOptions{}, &currentConfigs); err != nil {
		return errors.Wrap(err, "list istioConfigs failed")
	}

	if len(currentConfigs) == 0 {
		return nil
	}

	return errors.New("istioConfig already exist in current cluster, please update it")
}
