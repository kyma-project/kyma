package broker

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func fixOperationID() internal.OperationID {
	return internal.OperationID("op-id-123")
}

func fixApp() *internal.Application {
	return &internal.Application{
		Name: fixAppName(),
		Services: []internal.Service{
			{
				ID:          internal.ApplicationServiceID(fixServiceID()),
				DisplayName: "Orders",
				APIEntry: &internal.APIEntry{
					GatewayURL:  "www.gate.com",
					AccessLabel: "free",
				},
				EventProvider: true,
			},
		},
	}
}

func fixEventActivation() *v1alpha1.EventActivation {
	return &v1alpha1.EventActivation{
		TypeMeta: v1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      fixServiceID(),
			Namespace: fixNs(),
			OwnerReferences: []v1.OwnerReference{
				{
					UID:        fixServiceInstanceUID(),
					Name:       fixServiceInstanceName(),
					APIVersion: "servicecatalog.k8s.io/v1beta1",
					Kind:       "ServiceInstance",
				},
			},
		},
		Spec: v1alpha1.EventActivationSpec{
			DisplayName: "Orders",
			SourceID:    string(fixAppName()),
		},
	}
}

func fixInstanceID() internal.InstanceID {
	return internal.InstanceID("inst-123")
}

func fixNs() string {
	return "example-namesapce"
}

func fixNewCreateInstanceOperation() *internal.InstanceOperation {
	return &internal.InstanceOperation{
		InstanceID:  fixInstanceID(),
		OperationID: fixOperationID(),
		Type:        internal.OperationTypeCreate,
		State:       internal.OperationStateInProgress,
	}
}
func fixNewRemoveInstanceOperation() *internal.InstanceOperation {
	return &internal.InstanceOperation{
		InstanceID:  fixInstanceID(),
		OperationID: fixOperationID(),
		Type:        internal.OperationTypeRemove,
		State:       internal.OperationStateInProgress,
	}
}

func fixServiceID() string {
	return "service-id"
}

func fixPlanID() string {
	return "plan-id"
}

func fixNewInstance() *internal.Instance {
	return &internal.Instance{
		ID:            fixInstanceID(),
		Namespace:     internal.Namespace(fixNs()),
		ServiceID:     internal.ServiceID(fixServiceID()),
		ServicePlanID: internal.ServicePlanID(fixPlanID()),
		State:         internal.InstanceStatePending,
	}
}
func fixProvisionRequest() *osb.ProvisionRequest {
	return &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(fixInstanceID()),
		Context:           map[string]interface{}{},
		ServiceID:         fixServiceID(),
		PlanID:            fixPlanID(),
	}
}

func fixDeprovisionRequest() *osb.DeprovisionRequest {
	return &osb.DeprovisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(fixInstanceID()),
		ServiceID:         fixServiceID(),
		PlanID:            fixPlanID(),
	}
}

func fixServiceInstanceName() string {
	return "service-instance-name"
}

func fixServiceInstanceUID() types.UID {
	return types.UID("service-instance-uid-abcd-000")
}

func fixAppName() internal.ApplicationName {
	return "ec-prod"
}
func FixServiceInstance() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: v1.ObjectMeta{
			Name:      fixServiceInstanceName(),
			UID:       fixServiceInstanceUID(),
			Namespace: fixNs(),
		},
		Spec: v1beta1.ServiceInstanceSpec{
			ExternalID: string(fixInstanceID()),
		},
	}
}

func fixError() error {
	return errors.New("some erorr")
}
