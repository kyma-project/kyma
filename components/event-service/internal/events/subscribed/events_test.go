package subscribed

import (
	"testing"

	"github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/stretchr/testify/assert"
	coretypes "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEvents_getEventsFromSubscriptions(t *testing.T) {

	testNamespace := "namespace"
	appName := "test_app"
	eventType1 := "someType"
	eventType2 := "testType"

	t.Run("Should return event types from subscriptions", func(t *testing.T) {
		//given
		testSub1 := createSubscription(testNamespace, appName, eventType1, "some_sub")
		testSub2 := createSubscription(testNamespace, appName, eventType2, "test_sub")

		subscriptions := []v1alpha1.Subscription{*testSub1, *testSub2}

		subscriptionList := &v1alpha1.SubscriptionList{Items: subscriptions}

		//when
		events := getEventsFromSubscriptions(subscriptionList, appName)

		//then
		assert.Equal(t, 2, len(events))
		assert.True(t, containsEventName(events, eventType1))
		assert.True(t, containsEventName(events, eventType2))
	})

	t.Run("Should add event to list only if subscription matches application name", func(t *testing.T) {
		//given

		testSub1 := createSubscription(testNamespace, appName, eventType1, "some_sub")
		testSub2 := createSubscription(testNamespace, "someName", eventType2, "test_sub")

		subscriptions := []v1alpha1.Subscription{*testSub1, *testSub2}

		subscriptionList := &v1alpha1.SubscriptionList{Items: subscriptions}

		//when
		events := getEventsFromSubscriptions(subscriptionList, appName)

		//then
		assert.Equal(t, 1, len(events))
		assert.True(t, containsEventName(events, eventType1))
		assert.False(t, containsEventName(events, eventType2))
	})
}

func TestEvents_extractNamespacesNames(t *testing.T) {
	t.Run("Should extract names from namespaces", func(t *testing.T) {
		//given
		ns1 := *createNamespace("some_ns")
		ns2 := *createNamespace("test_ns")
		namespaceList := &coretypes.NamespaceList{Items: []coretypes.Namespace{ns1, ns2}}

		//when
		namespacesNames := extractNamespacesNames(namespaceList)

		//then
		assert.Equal(t, 2, len(namespacesNames))
	})
}

func TestEvents_removeDuplicates(t *testing.T) {
	t.Run("Should remove events with duplicated name and version", func(t *testing.T) {
		//given
		event := Event{Name: "someEvent", Version: "v1"}
		events := []Event{event, event}

		//when
		events = removeDuplicates(events)

		//then
		assert.Equal(t, 1, len(events))
		assert.True(t, containsEvent(events, event))
	})

	t.Run("Should not remove event when duplicated name but different version", func(t *testing.T) {
		//given
		eventV1 := Event{Name: "someEvent", Version: "v1"}
		eventV2 := Event{Name: "someEvent", Version: "v2"}

		events := []Event{eventV1, eventV2}

		//when
		events = removeDuplicates(events)

		//then
		assert.Equal(t, 2, len(events))
		assert.True(t, containsEvent(events, eventV1))
		assert.True(t, containsEvent(events, eventV2))
	})
}

func createSubscription(namespace, application, eventType, testSubscriptionName string) *v1alpha1.Subscription {
	return &v1alpha1.Subscription{
		TypeMeta: v1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      testSubscriptionName,
			Namespace: namespace,
		},
		SubscriptionSpec: v1alpha1.SubscriptionSpec{
			Endpoint:                      "https://some.test.endpoint",
			IncludeSubscriptionNameHeader: true,
			MaxInflight:                   400,
			PushRequestTimeoutMS:          2000,
			EventType:                     eventType,
			EventTypeVersion:              "v1",
			SourceID:                      application,
		},
	}
}

func createNamespace(name string) *coretypes.Namespace {
	return &coretypes.Namespace{
		TypeMeta: v1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}
}

func containsEventName(events []Event, eventType string) bool {
	for _, e := range events {
		if e.Name == eventType {
			return true
		}
	}
	return false
}

func containsEvent(events []Event, event Event) bool {
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}
