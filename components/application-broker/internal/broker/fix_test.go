package broker

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func fixOperationID() internal.OperationID {
	return internal.OperationID("op-id-123")
}

func fixRe() *internal.RemoteEnvironment {
	return &internal.RemoteEnvironment{
		Name: fixReName(),
		Services: []internal.Service{
			{
				ID:          internal.RemoteServiceID(fixServiceID()),
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
		ObjectMeta: v1.ObjectMeta{
			Name:      fixServiceID(),
			Namespace: fixNs(),
			OwnerReferences: []metav1.OwnerReference{
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
			SourceID:    string(fixReName()),
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
		ParamsHash:  "TODO",
	}
}
func fixNewRemoveInstanceOperation() *internal.InstanceOperation {
	return &internal.InstanceOperation{
		InstanceID:  fixInstanceID(),
		OperationID: fixOperationID(),
		Type:        internal.OperationTypeRemove,
		State:       internal.OperationStateInProgress,
		ParamsHash:  "TODO",
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
		ParamsHash:    "TODO",
	}
}
func fixProvisionRequest() *osb.ProvisionRequest {
	return &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(fixInstanceID()),
		Context:           map[string]interface{}{"namespace": fixNs()},
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

func fixReName() internal.RemoteEnvironmentName {
	return "ec-prod"
}
func FixServiceInstance() *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
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
