package alertpolicy

import (
	"fmt"

	"github.com/rancher/norman/api/access"
	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rancher/pkg/ref"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	v3client "github.com/rancher/types/client/management/v3"
)

const monitoringEnabled = "MonitoringEnabled"

func ClusterAlertPolicyValidator(resquest *types.APIContext, schema *types.Schema, data map[string]interface{}) error {
	clusterID := data["clusterId"].(string)
	var spec v3.ClusterPolicySpec
	if err := convert.ToObj(data, &spec); err != nil {
		return httperror.NewAPIError(httperror.InvalidBodyContent, fmt.Sprintf("%v", err))
	}

	if spec.Metrics != nil {
		tmp := map[string]interface{}{}
		fmt.Println(resquest.Version, v3client.ClusterType, clusterID)
		if err := access.ByID(resquest, resquest.Version, v3client.ClusterType, clusterID, &tmp); err != nil {
			return err
		}

		var cluster v3client.Cluster
		fmt.Println(resquest.Version, v3client.ClusterType, clusterID)
		if err := access.ByID(resquest, resquest.Version, v3client.ClusterType, clusterID, &cluster); err != nil {
			return err
		}

		if cluster.Conditions != nil {
			for _, v := range cluster.Conditions {
				if v.Type == monitoringEnabled && v.Status == "True" {
					return nil
				}
			}
		}
		return fmt.Errorf("if you want to use metric alert, need to enable monitoring for cluster %s", clusterID)
	}

	return nil
}

func ProjectAlertPolicyValidator(resquest *types.APIContext, schema *types.Schema, data map[string]interface{}) error {
	projectID := data["projectId"].(string)
	clusterID, _ := ref.Parse(projectID)
	var spec v3.ProjectPolicySpec
	if err := convert.ToObj(data, &spec); err != nil {
		return httperror.NewAPIError(httperror.InvalidBodyContent, fmt.Sprintf("%v", err))
	}

	if spec.Metrics != nil {
		cluster := &v3.Cluster{}
		fmt.Println(resquest.Version, v3client.ClusterType, clusterID)
		if err := access.ByID(resquest, resquest.Version, v3client.ClusterType, clusterID, cluster); err != nil {
			return err
		}
		if v3.ClusterConditionMonitoringEnabled.IsTrue(cluster) {
			return nil
		}
		return fmt.Errorf("if you want to use metric alert, need to enable monitoring for cluster %s", clusterID)
	}

	return nil
}
