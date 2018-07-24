package util

import (
	apiv1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma.cx/v1alpha1"
	eaApis "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/remoteenvironment.kyma.cx/v1alpha1"
	"github.com/satori/go.uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NewSubscription creates a new subscription
func NewSubscription(namespace string,
	subscriberEventEndpointURL string,
	eventType string,
	eventTypeVersion string,
	sourceEnvironment string,
	sourceNamespace string,
	sourceType string) *apiv1.Subscription {
	uid := uuid.NewV4().String()
	return &apiv1.Subscription{
		TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sub",
			Namespace: namespace,
			UID:       types.UID(uid),
		},

		SubscriptionSpec: apiv1.SubscriptionSpec{
			Endpoint:                      subscriberEventEndpointURL,
			IncludeSubscriptionNameHeader: false,
			IncludeTopicHeader:            false,
			MaxInflight:                   100,
			PushRequestTimeoutMS:          10,
			Source:                        apiv1.Source{SourceEnvironment: sourceEnvironment, SourceNamespace: sourceNamespace, SourceType: sourceType},
			EventType:                     eventType,
			EventTypeVersion:              eventTypeVersion,
		},
	}
}

// NewEventActivation creates a new event activation
func NewEventActivation(namespace string) *eaApis.EventActivation {
	return &eaApis.EventActivation{
		TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ea",
			Namespace: namespace,
		},
		EventActivationSpec: eaApis.EventActivationSpec{
			DisplayName: "e2e-test-event-activation",
			Source:      eaApis.Source{Environment: "test", Namespace: "local.kyma.commerce", Type: "commerce"},
		},
	}
}
