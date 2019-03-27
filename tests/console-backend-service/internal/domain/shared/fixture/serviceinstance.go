package fixture

import (
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
)

func ServiceInstance(name string, namespace string) shared.ServiceInstance {
	return shared.ServiceInstance{
		Name:      name,
		Namespace: namespace,
		Labels:    []string{"test", "test2"},
		PlanSpec: map[string]interface{}{
			"planName": "test",
			"additionalData": "foo",
		},
		ClusterServicePlan: shared.ClusterServicePlan{
			Name:         "a6078799-70a1-4674-af91-aba44dd6a56",
			ExternalName: "full",
		},
		ClusterServiceClass: shared.ClusterServiceClass{
			Name:         "faebbe18-0a84-11e9-ab14-d663bd873d94",
			ExternalName: "testing",
		},
		Status: shared.ServiceInstanceStatus{
			Type: shared.ServiceInstanceStatusTypeRunning,
		},
		Bindable: true,
	}
}
