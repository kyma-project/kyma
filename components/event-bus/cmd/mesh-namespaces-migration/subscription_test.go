package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	kneventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	kneventingfakeclientset "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kymaeventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	kymaeventingfakeclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/fake"
)

func TestNewSubscriptionMigrator(t *testing.T) {
	testUserNamespaces := []string{
		"ns1",
		"ns2",
		// ns3 excluded
		"ns4",
	}

	testSubscriptions := []kymaeventingv1alpha1.Subscription{
		// ns1
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-sub-ns1",
				Namespace: "ns1",
			},
		},
		// ns2
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-sub-ns2-1",
				Namespace: "ns2",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-sub-ns2-2",
				Namespace: "ns2",
			},
		},
		// ns3
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-sub-ns3",
				Namespace: "ns3",
			},
		},
	}

	testTriggers := []kneventingv1alpha1.Trigger{
		// ns1
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-trigger-ns1",
				Namespace: "ns1",
			},
		},
		// ns3
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-trigger-ns3",
				Namespace: "ns3",
			},
		},
		// ns4
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-trigger-ns4-1",
				Namespace: "ns4",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-trigger-ns4-2",
				Namespace: "ns4",
			},
		},
	}

	kymaCli := kymaeventingfakeclientset.NewSimpleClientset(
		subscriptionsToObjectSlice(testSubscriptions)...,
	)
	knCli := kneventingfakeclientset.NewSimpleClientset(
		triggersToObjectSlice(testTriggers)...,
	)

	m, err := newSubscriptionMigrator(kymaCli, knCli, testUserNamespaces)
	if err != nil {
		t.Fatalf("Failed to initialize subscriptionMigrator: %s", err)
	}

	// expect
	//  1 sub  from ns1 (user namespace)
	//  2 subs from ns2 (user namespace)
	//  0 sub  from ns3 (non-user namespace)
	//  0 sub  from ns4 (does not contain any object)

	expect := sets.NewString(
		"ns1/my-sub-ns1",
		"ns2/my-sub-ns2-1",
		"ns2/my-sub-ns2-2",
	)
	got := sets.NewString(
		subscriptionsToKeys(m.subscriptions)...,
	)

	if !got.Equal(expect) {
		t.Errorf("Unexpected Subscriptions: (-:expect, +:got) %s", cmp.Diff(expect, got))
	}

	// expect
	//  1 trigger  from ns1 (user namespace)
	//  0 trigger  from ns2 (does not contain any object)
	//  0 trigger  from ns3 (non-user namespace)
	//  2 triggers from ns4 (user namespace)

	expect = sets.NewString(
		"ns1/my-trigger-ns1",
		"ns4/my-trigger-ns4-1",
		"ns4/my-trigger-ns4-2",
	)
	got = sets.NewString(
		triggersToKeys(m.triggersByNamespace)...,
	)

	if !got.Equal(expect) {
		t.Errorf("Unexpected Triggers: (-:expect, +:got) %s", cmp.Diff(expect, got))
	}
}

func TestMigrateSubscription(t *testing.T) {
	const (
		testEndpoint         = "http://example.com:8080/event"
		testSource           = "varkes"
		testEventType        = "order.created"
		testEventTypeVersion = "v1"
	)

	testSubscription := &kymaeventingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-subscription",
			Namespace: "ns",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		SubscriptionSpec: kymaeventingv1alpha1.SubscriptionSpec{
			Endpoint:         testEndpoint,
			SourceID:         testSource,
			EventType:        testEventType,
			EventTypeVersion: testEventTypeVersion,
		},
	}

	testCases := map[string]struct {
		triggers     []kneventingv1alpha1.Trigger
		expectCreate bool
	}{
		"Namespace does not contain any Trigger": {
			triggers:     nil,
			expectCreate: true,
		},
		"Namespace contains a matching Trigger": {
			triggers: []kneventingv1alpha1.Trigger{
				newTrigger(testSource, testEventType, testEventTypeVersion, testSubscription.Endpoint),
			},
			expectCreate: false,
		},
		"Namespace contains a non-matching Trigger": {
			triggers: []kneventingv1alpha1.Trigger{
				newTrigger("commerce", "order.cancelled", "v2", "http://example.com:8080/invalid"),
			},
			expectCreate: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m := subscriptionMigrator{
				kymaClient: kymaeventingfakeclientset.NewSimpleClientset(testSubscription),
				knativeClient: kneventingfakeclientset.NewSimpleClientset(
					triggersToObjectSlice(tc.triggers)...,
				),
				subscriptions:       subscriptionsList{*testSubscription},
				triggersByNamespace: triggersByNamespaceMap{"ns": tc.triggers},
			}

			err := m.migrateAll()
			if err != nil {
				t.Fatalf("Failed to migrate Subscription: %s", err)
			}

			// 1. Assert that the namespace contains the expected number of Triggers

			triggers, err := m.knativeClient.EventingV1alpha1().Triggers("ns").List(metav1.ListOptions{})
			if err != nil {
				t.Fatalf("Failed to list Triggers: %s", err)
			}

			expectTriggerCount := len(tc.triggers)
			if tc.expectCreate {
				expectTriggerCount++
			}
			if count := len(triggers.Items); count != expectTriggerCount {
				found := triggersToKeys(triggersByNamespaceMap{
					"ns": triggers.Items,
				})
				t.Fatalf("Expected %d Triggers, got %d: %q", expectTriggerCount, count, found)
			}

			// 2. Assert that the matching Trigger contains the correct spec/attributes

			// fake ClientSets return results by order of creation,
			// so a newly created Trigger will always be the last result
			trigger := triggers.Items[len(triggers.Items)-1]

			gotEndpoint := trigger.Spec.Subscriber.URI.String()
			expectEndpoint := testSubscription.Endpoint
			if gotEndpoint != expectEndpoint {
				t.Errorf("Expected Endpoint to be %q, got %q", expectEndpoint, gotEndpoint)
			}

			for attr, expectVal := range map[string]string{
				sourceAttr:           testSubscription.SourceID,
				eventTypeAttr:        testSubscription.EventType,
				eventTypeVersionAttr: testSubscription.EventTypeVersion,
			} {
				gotVal := (*trigger.Spec.Filter.Attributes)[attr]
				if gotVal != expectVal {
					t.Errorf("Expected %q attribute to be %q, got %q", attr, expectVal, gotVal)
				}
			}

			// 3. Ensure the Subscription was deleted

			_, err = m.kymaClient.EventingV1alpha1().Subscriptions("ns").Get("my-subscription", metav1.GetOptions{})
			if !apierrors.IsNotFound(err) {
				t.Errorf("Expected Subscription to be deleted, got error %v", err)
			}
		})
	}
}

func subscriptionsToObjectSlice(subs []kymaeventingv1alpha1.Subscription) []runtime.Object {
	objects := make([]runtime.Object, len(subs))
	for i := range subs {
		objects[i] = &subs[i]
	}
	return objects
}

func triggersToObjectSlice(triggers []kneventingv1alpha1.Trigger) []runtime.Object {
	objects := make([]runtime.Object, len(triggers))
	for i := range triggers {
		objects[i] = &triggers[i]
	}
	return objects
}

func subscriptionsToKeys(subs subscriptionsList) []string {
	keys := make([]string, len(subs))
	for i, sub := range subs {
		keys[i] = fmt.Sprintf("%s/%s", sub.Namespace, sub.Name)
	}
	return keys
}

func triggersToKeys(triggersByNs triggersByNamespaceMap) []string {
	var keys []string
	for _, triggers := range triggersByNs {
		for _, trigger := range triggers {
			keys = append(keys, fmt.Sprintf("%s/%s", trigger.Namespace, trigger.Name))
		}
	}
	return keys
}

func newTrigger(source, eventType, eventTypeVersion, subscriberURI string) kneventingv1alpha1.Trigger {
	endpointURL, _ := apis.ParseURL(subscriberURI)

	return kneventingv1alpha1.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-trigger-ns",
			Namespace: "ns",
		},
		Spec: kneventingv1alpha1.TriggerSpec{
			Broker: defaultBrokerName,
			Filter: &kneventingv1alpha1.TriggerFilter{
				Attributes: &kneventingv1alpha1.TriggerFilterAttributes{
					sourceAttr:           source,
					eventTypeAttr:        eventType,
					eventTypeVersionAttr: eventTypeVersion,
				},
			},
			Subscriber: duckv1.Destination{
				URI: endpointURL,
			},
		},
	}
}
