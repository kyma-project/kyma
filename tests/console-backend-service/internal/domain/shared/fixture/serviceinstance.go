package fixture

import (
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
)

func ServiceInstanceFromClusterServiceClass(name string, namespace string) shared.ServiceInstance {
	return shared.ServiceInstance{
		Name:      name,
		Namespace: namespace,
		Labels:    []string{"test", "test2"},
		PlanSpec:  TestingBundleFullPlanSpec,
		ClusterServicePlan: shared.ClusterServicePlan{
			Name:         TestingBundleFullPlanName,
			ExternalName: TestingBundleFullPlanExternalName,
		},
		ClusterServiceClass: shared.ClusterServiceClass{
			Name:         TestingBundleClassName,
			ExternalName: TestingBundleClassExternalName,
		},
		Status: shared.ServiceInstanceStatus{
			Type: shared.ServiceInstanceStatusTypeRunning,
		},
		Bindable: true,
	}
}

func ServiceInstanceFromServiceClass(name string, namespace string) shared.ServiceInstance {
	return shared.ServiceInstance{
		Name:      name,
		Namespace: namespace,
		Labels:    []string{"test", "test2"},
		PlanSpec:  TestingBundleFullPlanSpec,
		ServicePlan: shared.ServicePlan{
			Name:         TestingBundleFullPlanName,
			ExternalName: TestingBundleFullPlanExternalName,
		},
		ServiceClass: shared.ServiceClass{
			Name:         TestingBundleClassName,
			ExternalName: TestingBundleClassExternalName,
			Namespace:    namespace,
		},
		Status: shared.ServiceInstanceStatus{
			Type: shared.ServiceInstanceStatusTypeRunning,
		},
		Bindable: true,
	}
}
