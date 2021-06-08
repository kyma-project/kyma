package nats

import (
	"context"
	"fmt"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

type FakeCommander struct {
	client  dynamic.Interface
	backend *handlers.Nats
}

func (c *FakeCommander) Init(mgr manager.Manager) error {
	return nil
}

type FakeCleaner struct {
}

func (c *FakeCleaner) Clean(eventType string) (string, error) {
	// Cleaning is not needed in this test
	return eventType, nil
}

func (c *FakeCommander) Start() error {
	return nil
}

func (c *FakeCommander) Stop() error {
	return nil
}

func TestCleanup(t *testing.T) {
	natsCommander := FakeCommander{}
	g := gomega.NewWithT(t)
	data := "sampledata"
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	subscriberStoreURL, subscriberCheckURL := "", ""

	// When
	// Create a test subscriber
	ctx := context.Background()
	log := ctrl.Log.WithName("test-cleaner-nats")
	subscriberPort := 8081
	subscriber := controllertesting.NewSubscriber(fmt.Sprintf(":%d", subscriberPort))
	subscriber.Start()
	// Shutting down subscriber
	defer subscriber.Shutdown()

	subscriberStoreURL = fmt.Sprintf("http://localhost:%d%s", subscriberPort, subscriber.StoreEndpoint)
	subscriberCheckURL = fmt.Sprintf("http://localhost:%d%s", subscriberPort, subscriber.CheckEndpoint)

	// Create test subscription
	testSub := controllertesting.NewSubscription("test", "test", controllertesting.WithFakeSubscriptionStatus, controllertesting.WithEventTypeFilter)
	testSub.Spec.Sink = subscriberStoreURL

	// Create NATS Server
	natsPort := 4222
	natsServer := controllertesting.RunNatsServerOnPort(natsPort)
	natsURL := natsServer.ClientURL()
	envConf := env.NatsConfig{
		Url:             natsURL,
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: controllertesting.EventTypePrefix,
	}
	natsCommander.backend = handlers.NewNats(envConf, log)
	err := natsCommander.backend.Initialize(env.Config{})
	g.Expect(err).To(gomega.BeNil())

	// Create fake Dynamic clients
	fakeClient, err := NewFakeClient(testSub)
	g.Expect(err).To(gomega.BeNil())
	natsCommander.client = fakeClient

	fakeCleaner := FakeCleaner{}

	// Create NATS subscription
	_, err = natsCommander.backend.SyncSubscription(testSub, &fakeCleaner)
	g.Expect(err).To(gomega.BeNil())

	// Make sure subscriber works
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}
	// Send an event
	err = handlers.SendEventToNATS(natsCommander.backend, data)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Check for the event
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Then
	err = cleanup(natsCommander.backend, natsCommander.client)
	g.Expect(err).To(gomega.BeNil())

	// Expect
	unstructuredSub, err := natsCommander.client.Resource(SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, testSub.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	gotSub, err := toSubscription(unstructuredSub)
	g.Expect(err).To(gomega.BeNil())
	expectedSubStatus := eventingv1alpha1.SubscriptionStatus{}
	g.Expect(expectedSubStatus).To(gomega.Equal(gotSub.Status))

	// Test NATS subscriptions are gone
	// Send again an event and check subscriber, check subscriober should fail after 5 retries
	err = handlers.SendEventToNATS(natsCommander.backend, data)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	g.Expect(err).NotTo(gomega.BeNil())
}

func toSubscription(unstructuredSub *unstructured.Unstructured) (*eventingv1alpha1.Subscription, error) {
	subscription := new(eventingv1alpha1.Subscription)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredSub.Object, subscription)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

type Client struct {
	Resource dynamic.NamespaceableResourceInterface
}

func NewFakeClient(sub *eventingv1alpha1.Subscription) (dynamic.Interface, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return nil, err
	}

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, sub)
	return dynamicClient, nil
}

func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func SubscriptionGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  eventingv1alpha1.GroupVersion.Version,
		Group:    eventingv1alpha1.GroupVersion.Group,
		Resource: "subscriptions",
	}
}
