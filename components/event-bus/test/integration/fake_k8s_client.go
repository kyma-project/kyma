package integration

import (
	"context"
	"log"
	"time"

	"github.com/gofrs/uuid"
	subApi "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/event-bus/generated/push/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
)

var (
	client *fake.Clientset
)

func newFakeInformer(ctx context.Context) cache.SharedIndexInformer {

	client = fake.NewSimpleClientset()

	informers := externalversions.NewSharedInformerFactory(client, 0)

	informer := informers.Eventing().V1alpha1().Subscriptions().Informer()

	informers.Start(ctx.Done())

	if !informer.HasSynced() {
		time.Sleep(10 * time.Millisecond)
	}

	return informer
}

func createNewSubscription(name string, namespace string, subscriberEventEndpointURL string, eventType string, eventTypeVersion string,
	sourceID string) (*subApi.Subscription, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Error while generating UID: %v", err)
	}
	return client.EventingV1alpha1().Subscriptions(namespace).Create(getSubscriptionResource(name, namespace, uid.String(), subscriberEventEndpointURL, sourceID, eventType, eventTypeVersion))
}

func updateSubscription(name string, namespace string, subscriberEventEndpointURL string, eventType string, eventTypeVersion string,
	sourceID string) (*subApi.Subscription, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Error while generating UID: %v", err)
	}
	return client.EventingV1alpha1().Subscriptions(namespace).Update(getSubscriptionResource(name, namespace, uid.String(), subscriberEventEndpointURL, sourceID, eventType, eventTypeVersion))
}

func getSubscriptionResource(name string, namespace string, uid string, subscriberEventEndpointURL string, sourceID string, eventType string, eventTypeVersion string) *subApi.Subscription {
	return &subApi.Subscription{
		TypeMeta: metav1.TypeMeta{APIVersion: subApi.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
		},

		SubscriptionSpec: subApi.SubscriptionSpec{
			Endpoint:                      subscriberEventEndpointURL,
			IncludeSubscriptionNameHeader: false,
			MaxInflight:                   100,
			PushRequestTimeoutMS:          10,
			SourceID:                      sourceID,
			EventType:                     eventType,
			EventTypeVersion:              eventTypeVersion,
		},

		Status: subApi.SubscriptionStatus{
			Status: subApi.Status{
				Conditions: []subApi.SubscriptionCondition{
					{
						Type:   subApi.EventsActivated,
						Status: subApi.ConditionTrue,
					},
				},
			},
		},
	}
}

func deleteSubscription(name string, namespace string) error {
	return client.EventingV1alpha1().Subscriptions(namespace).Delete(name,
		&metav1.DeleteOptions{TypeMeta: metav1.TypeMeta{APIVersion: subApi.SchemeGroupVersion.String()}},
	)
}
