package fixture

import (
	tester "github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/domain/shared"
)

func ServiceInstance(name string) shared.ServiceInstance {
	return shared.ServiceInstance{
		Name:        name,
		Environment: tester.DefaultNamespace,
		Labels:      []string{"test", "test2"},
		PlanSpec: map[string]interface{}{
			"first": "1",
			"second": map[string]interface{}{
				"value": "2",
			},
		},
		ClusterServicePlan: shared.ClusterServicePlan{
			Name:         "86064792-7ea2-467b-af93-ac9694d96d52",
			ExternalName: "default",
		},
		ClusterServiceClass: shared.ClusterServiceClass{
			Name:         "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
			ExternalName: "user-provided-service",
		},
		Status: shared.ServiceInstanceStatus{
			Type: shared.ServiceInstanceStatusTypeRunning,
		},
		Bindable: true,
	}
}
