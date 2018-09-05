package populator

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
)

// Converter is responsible for converting Service Catalog's Service Instance to internal representation
type Converter struct{}

// MapServiceInstance converts SC Service Instance to its internal representation
func (c *Converter) MapServiceInstance(in *v1beta1.ServiceInstance) *internal.Instance {
	var state internal.InstanceState

	if c.isServiceInstanceReady(in) {
		state = internal.InstanceStateSucceeded
	} else {
		state = internal.InstanceStateFailed
	}

	var planID internal.ServicePlanID
	var serviceID internal.ServiceID

	if in.Spec.ClusterServicePlanRef != nil {
		planID = internal.ServicePlanID(in.Spec.ClusterServicePlanRef.Name)
	} else {
		planID = internal.ServicePlanID(in.Spec.ServicePlanRef.Name)
	}

	if in.Spec.ClusterServiceClassRef != nil {
		serviceID = internal.ServiceID(in.Spec.ClusterServiceClassRef.Name)
	} else {
		serviceID = internal.ServiceID(in.Spec.ServiceClassRef.Name)
	}

	return &internal.Instance{
		ID:            internal.InstanceID(in.Spec.ExternalID),
		Namespace:     internal.Namespace(in.Namespace),
		ParamsHash:    "TODO",
		ServicePlanID: planID,
		ServiceID:     serviceID,
		State:         state,
	}
}

func (c *Converter) isServiceInstanceReady(instance *v1beta1.ServiceInstance) bool {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == v1beta1.ServiceInstanceConditionReady {
			return cond.Status == v1beta1.ConditionTrue
		}
	}
	return false
}
