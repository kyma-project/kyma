package util

import (
	"log"

	"github.com/gofrs/uuid"
	eaApis "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	apiv1 "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NewSubscription creates a new subscription
func NewSubscription(name string, namespace string, subscriberEventEndpointURL string, eventType string, eventTypeVersion string,
	sourceID string) *apiv1.Subscription {
	uid, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Error while generating UID: %v", err)
	}
	return &apiv1.Subscription{
		TypeMeta: metav1.TypeMeta{APIVersion: apiv1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid.String()),
		},

		SubscriptionSpec: apiv1.SubscriptionSpec{
			Endpoint:                      subscriberEventEndpointURL,
			IncludeSubscriptionNameHeader: false,
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
