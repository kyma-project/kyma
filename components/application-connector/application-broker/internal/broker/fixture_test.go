package broker_test

import (
	"github.com/kyma-project/kyma/components/application-broker/internal"
)

type expAll struct {
	InstanceID  internal.InstanceID
	OperationID internal.OperationID

	Service struct {
		ID internal.ServiceID
	}
	ServicePlan struct {
		ID internal.ServicePlanID
	}
	Namespace internal.Namespace
}

func (exp *expAll) Populate() {
	exp.InstanceID = internal.InstanceID("fix-I-ID")
	exp.OperationID = internal.OperationID("fix-OP-ID")

	exp.Namespace = internal.Namespace("fix-namespace")
}

func (exp *expAll) NewInstance() *internal.Instance {
	return &internal.Instance{
		ID:            exp.InstanceID,
		ServiceID:     exp.Service.ID,
		ServicePlanID: exp.ServicePlan.ID,
		Namespace:     exp.Namespace,
	}
}

func (exp *expAll) NewInstanceOperation(tpe internal.OperationType, state internal.OperationState) *internal.InstanceOperation {
	return &internal.InstanceOperation{
		InstanceID:  exp.InstanceID,
		OperationID: exp.OperationID,
		Type:        tpe,
		State:       state,
	}
}
