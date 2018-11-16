package util

import (
	apiv1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	eaApis "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
	"github.com/satori/go.uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NewSubscription creates a new subscription
func NewSubscription(name string, namespace string, subscriberEventEndpointURL string, eventType string, eventTypeVersion string,
	sourceID string) *apiv1.Subscription {
	uid := uuid.NewV4().String()
	return &apiv1.Subscription{
		TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
		},

		SubscriptionSpec: apiv1.SubscriptionSpec{
			Endpoint:                      subscriberEventEndpointURL,
			IncludeSubscriptionNameHeader: false,
			MaxInflight:                   100,
			PushRequestTimeoutMS:          10,
			SourceID:                      sourceID,
			EventType:                     eventType,
			EventTypeVersion:              eventTypeVersion,
		},
	}
}

// NewEventActivation creates a new event activation
func NewEventActivation(name string, namespace string, sourceID string) *eaApis.EventActivation {
	return &eaApis.EventActivation{
		TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		EventActivationSpec: eaApis.EventActivationSpec{
			DisplayName: name,
			SourceID:    sourceID,
		},
	}
}
