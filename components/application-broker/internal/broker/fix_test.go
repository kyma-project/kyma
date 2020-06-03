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
		Name:            fixAppName(),
		CompassMetadata: fixAppID(),
		Services: []internal.Service{
			{
				ID:          fixAppServiceID(),
				DisplayName: fixDisplayName(),
				Entries: []internal.Entry{
					{
						Type: "API",
						APIEntry: &internal.APIEntry{
							GatewayURL:  "http://gateway.io",
							TargetURL:   "http://target.io",
							Name:        "api-mock",
							AccessLabel: "access-label",
						},
					},
					{
						Type: "Events",
					},
				},
				EventProvider: true,
			},
		},
	}
}

func fixAppID() internal.CompassMetadata {
	return internal.CompassMetadata{ApplicationID: "123"}
}

func fixEventActivation() *v1alpha1.EventActivation {
	return &v1alpha1.EventActivation{
		TypeMeta: v1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      string(fixServiceID()),
			Namespace: string(fixNs()),
		},
		Spec: v1alpha1.EventActivationSpec{
			DisplayName: fixDisplayName(),
			SourceID:    string(fixAppName()),
		},
	}
}

func fixInstanceID() internal.InstanceID {
	return internal.InstanceID("inst-123")
}

func fixNs() internal.Namespace {
	return "example-namespace"
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

func fixServiceID() internal.ServiceID {
	return "service-id"
}

func fixAppServiceID() internal.ApplicationServiceID {
	return internal.ApplicationServiceID(fixServiceID())
}

func fixPlanID() string {
	return "plan-id"
}

func fixNewInstance() *internal.Instance {
	return &internal.Instance{
		ID:            fixInstanceID(),
		Namespace:     fixNs(),
		ServiceID:     fixServiceID(),
		ServicePlanID: internal.ServicePlanID(fixPlanID()),
		State:         internal.InstanceStatePending,
	}
}
func fixProvisionRequest() *osb.ProvisionRequest {
	return &osb.ProvisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(fixInstanceID()),
		Context:           map[string]interface{}{"namespace": string(fixNs())},
		ServiceID:         string(fixServiceID()),
		PlanID:            fixPlanID(),
	}
}

func fixDeprovisionRequest() *osb.DeprovisionRequest {
	return &osb.DeprovisionRequest{
		AcceptsIncomplete: true,
		InstanceID:        string(fixInstanceID()),
		ServiceID:         string(fixServiceID()),
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
			Namespace: string(fixNs()),
		},
		Spec: v1beta1.ServiceInstanceSpec{
			ExternalID: string(fixInstanceID()),
		},
	}
}

func fixError() error {
	return errors.New("some error")
}

func fixProvisionSucceeded() *string {
	s := internal.OperationDescriptionProvisioningSucceeded
	return &s
}

func fixDeprovisionSucceeded() *string {
	s := internal.OperationDescriptionDeprovisioningSucceeded
	return &s
}

func fixDisplayName() string {
	return "Orders"
}

func fixEventProvider() bool {
	return true
}

func stringPtr(in string) *string {
	return &in
}
