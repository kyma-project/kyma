package subscribed

import (
	"k8s.io/apimachinery/pkg/runtime"
	kneventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/stretchr/testify/assert"
	coretypes "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing/pkg/client/clientset/versioned/fake"
)

func newStubSubscription(subscriptionList *v1alpha1.SubscriptionList) *stubSubscriptions {
	return &stubSubscriptions{
		subscriptionList: subscriptionList,
	}
}

type stubSubscriptions struct {
	subscriptionList *v1alpha1.SubscriptionList
}

func (sSub *stubSubscriptions) Create(*v1alpha1.Subscription) (*v1alpha1.Subscription, error) {
	return nil, nil
}
func (sSub *stubSubscriptions) Update(*v1alpha1.Subscription) (*v1alpha1.Subscription, error) {
	return nil, nil
}
func (sSub *stubSubscriptions) UpdateStatus(*v1alpha1.Subscription) (*v1alpha1.Subscription, error) {
	return nil, nil
}
func (sSub *stubSubscriptions) Delete(name string, options *v1.DeleteOptions) error { return nil }
func (sSub *stubSubscriptions) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return nil
}
func (sSub *stubSubscriptions) Get(name string, options v1.GetOptions) (*v1alpha1.Subscription, error) {
	return nil, nil
}
func (sSub *stubSubscriptions) List(opts v1.ListOptions) (*v1alpha1.SubscriptionList, error) {
	return sSub.subscriptionList, nil
}
func (*stubSubscriptions) Watch(opts v1.ListOptions) (watch.Interface, error) { return nil, nil }
func (*stubSubscriptions) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Subscription, err error) {
	return nil, nil
}

//TODO(marcobebway) Enable this test
func _TestEvents_GetSubscribedEvents(t *testing.T) {

	testNamespace1 := "namespace"
	testNamespace2 := "test-namespace"

	appName := "test_app"
	eventType1 := "someType"
	eventType2 := "testType"

	t.Run("Should return subscribed events from multiple namespaces", func(t *testing.T) {
		//given
		testSub1 := createSubscription(testNamespace1, appName, eventType1, "some_sub", "v1")
		testSub2 := createSubscription(testNamespace2, appName, eventType2, "test_sub", "v1")

		subscriptions := &v1alpha1.SubscriptionList{Items: []v1alpha1.Subscription{*testSub1, *testSub2}}

		stubSubscriptions := newStubSubscription(subscriptions)

		stubSubscriptionsGetter := &mocks.SubscriptionsGetter{}
		stubSubscriptionsGetter.On("Subscriptions", mock.Anything).Return(stubSubscriptions)

		ns1 := *createNamespace(testNamespace1)
		ns2 := *createNamespace(testNamespace2)
		namespaceList := &coretypes.NamespaceList{Items: []coretypes.Namespace{ns1, ns2}}

		nsClient := &mocks.NamespacesClient{}
		nsClient.On("List", v1.ListOptions{}).Return(namespaceList, nil)

		eventsClient := NewEventsClient(nil)

		//when
		events, e := eventsClient.GetSubscribedEvents(appName)

		//then
		require.NoError(t, e)
		eventsInfo := events.EventsInfo

		assert.Equal(t, 2, len(eventsInfo))
		assert.True(t, containsEventName(eventsInfo, eventType1))
		assert.True(t, containsEventName(eventsInfo, eventType2))
	})

	t.Run("Should return subscribed events without duplicates", func(t *testing.T) {
		//given
		testSub1 := createSubscription(testNamespace1, appName, eventType1, "some_sub", "v1")
		testSub2 := createSubscription(testNamespace2, appName, eventType1, "test_sub", "v1")

		subscriptions := &v1alpha1.SubscriptionList{Items: []v1alpha1.Subscription{*testSub1, *testSub2}}

		stubSubscriptions := newStubSubscription(subscriptions)

		stubSubscriptionsGetter := &mocks.SubscriptionsGetter{}
		stubSubscriptionsGetter.On("Subscriptions", mock.Anything).Return(stubSubscriptions)

		ns1 := *createNamespace(testNamespace1)
		ns2 := *createNamespace(testNamespace2)
		namespaceList := &coretypes.NamespaceList{Items: []coretypes.Namespace{ns1, ns2}}

		nsClient := &mocks.NamespacesClient{}
		nsClient.On("List", v1.ListOptions{}).Return(namespaceList, nil)

		eventsClient := NewEventsClient(nil)

		//when
		events, e := eventsClient.GetSubscribedEvents(appName)

		//then
		require.NoError(t, e)
		eventsInfo := events.EventsInfo

		assert.Equal(t, 1, len(eventsInfo))
		assert.True(t, containsEventName(eventsInfo, eventType1))
	})

	t.Run("Should return error when fetching namespaces fails", func(t *testing.T) {
		//given
		testSub1 := createSubscription(testNamespace1, appName, eventType1, "some_sub", "v1")
		testSub2 := createSubscription(testNamespace2, appName, eventType1, "test_sub", "v1")

		subscriptions := &v1alpha1.SubscriptionList{Items: []v1alpha1.Subscription{*testSub1, *testSub2}}

		stubSubscriptions := newStubSubscription(subscriptions)

		stubSubscriptionsGetter := &mocks.SubscriptionsGetter{}
		stubSubscriptionsGetter.On("Subscriptions", mock.Anything).Return(stubSubscriptions)

		nsClient := &mocks.NamespacesClient{}
		nsClient.On("List", v1.ListOptions{}).Return(&coretypes.NamespaceList{}, errors.New("Some error"))

		eventsClient := NewEventsClient(nil)

		//when
		_, e := eventsClient.GetSubscribedEvents(appName)

		//then
		require.Error(t, e)
	})
}

func createSubscription(namespace, application, eventType, testSubscriptionName, version string) *v1alpha1.Subscription {
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
			EventType:                     eventType,
			EventTypeVersion:              version,
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

func Test_stuff(t *testing.T) {
	// create trigger objects, with the correct source
	// create trigger objects, with the incorrect source
	// pass these objects to the NewSimpleClientset
	// then assert the correct result is returned

	tr := kneventingv1alpha1.Trigger{
		Spec: kneventingv1alpha1.TriggerSpec{
			Filter: &kneventingv1alpha1.TriggerFilter{
				Attributes: &kneventingv1alpha1.TriggerFilterAttributes{
					"source":           "mock",
					"type":             "test-type-1",
					"eventtypeversion": "test-eventtypeversion-1",
				},
			},
		},
	}

	objects := make([]runtime.Object, 0)
	objects = append(objects, &tr)

	clientSet := fake.NewSimpleClientset(objects...)
	eventClient := NewEventsClient(clientSet)
	events, err := eventClient.GetSubscribedEvents("mock")
	if err != nil {
		t.Fatalf("error: %+v", err)
	}
	t.Logf("Events: %+v", events)
}
