package logging

import (
	"github.com/rancher/types/apis/management.cattle.io/v3"

	"github.com/rancher/types/config/dialer"
)

// type Validator struct {
// 	ClusterDialer dialer.Factory
// }

// func (v *Validator) ClusterLoggingValidator(resquest *types.APIContext, schema *types.Schema, data map[string]interface{}) error {
// 	var clusterLogging v3.ClusterLoggingSpec
// 	if err := convert.ToObj(data, &clusterLogging); err != nil {
// 		return httperror.NewAPIError(httperror.InvalidBodyContent, fmt.Sprintf("%v", err))
// 	}

// 	clusterLogging.ClusterName = convert.ToString(data["clusterId"])

// 	return testClusterLoggingReachable(clusterLogging, v.ClusterDialer)
// }

// func (v *Validator) ProjectLoggingValidator(resquest *types.APIContext, schema *types.Schema, data map[string]interface{}) error {
// 	var projectLogging v3.ProjectLoggingSpec
// 	if err := convert.ToObj(data, &projectLogging); err != nil {
// 		return httperror.NewAPIError(httperror.InvalidBodyContent, fmt.Sprintf("%v", err))
// 	}

// 	clusterName, _ := ref.Parse(convert.ToString(data["projectId"]))
// 	return testProjectLoggingReachable(clusterName, projectLogging, v.ClusterDialer)
// }

func testClusterLoggingReachable(input v3.LoggingInput, ClusterDialer dialer.Factory) error {
	// wp := utils.WrapClusterLogging{
	// 	ClusterLoggingSpec: spec,
	// }

	// if err := wp.Validate(ClusterDialer); err != nil {
	// 	return httperror.NewAPIError(httperror.InvalidFormat, fmt.Sprintf("%v", err))
	// }
	return nil
}

func testProjectLoggingReachable(clusterName string, input v3.LoggingInput, ClusterDialer dialer.Factory) error {
	// wp := utils.WrapProjectLogging{
	// 	ClusterName:        clusterName,
	// 	ProjectLoggingSpec: spec,
	// }

	// if err := wp.Validate(ClusterDialer); err != nil {
	// 	return httperror.NewAPIError(httperror.InvalidFormat, fmt.Sprintf("%v", err))
	// }
	return nil
}
