package nats

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander/fake"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/onsi/gomega"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCleanup(t *testing.T) {
	natsCommander := fake.Commander{}
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
	defer controllertesting.ShutDownNATSServer(natsServer)

	envConf := env.NatsConfig{
		Url:             natsURL,
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: controllertesting.EventTypePrefix,
	}
	natsBackend := handlers.NewNats(envConf, log)
	natsCommander.Backend = natsBackend
	err := natsCommander.Backend.Initialize(env.Config{})
	g.Expect(err).To(gomega.BeNil())

	// Create fake Dynamic clients
	fakeClient, err := controllertesting.NewFakeSubscriptionClient(testSub)
	g.Expect(err).To(gomega.BeNil())
	natsCommander.Client = fakeClient

	fakeCleaner := fake.Cleaner{}

	// Create NATS subscription
	_, err = natsCommander.Backend.SyncSubscription(testSub, &fakeCleaner)
	g.Expect(err).To(gomega.BeNil())

	// Make sure subscriber works
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}
	// Send an event
	err = handlers.SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	// Check for the event
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Then
	err = cleanup(natsCommander.Backend, natsCommander.Client)
	g.Expect(err).To(gomega.BeNil())

	// Expect
	unstructuredSub, err := natsCommander.Client.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, testSub.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	gotSub, err := controllertesting.ToSubscription(unstructuredSub)
	g.Expect(err).To(gomega.BeNil())
	expectedSubStatus := eventingv1alpha1.SubscriptionStatus{}
	g.Expect(expectedSubStatus).To(gomega.Equal(gotSub.Status))

	// Test NATS subscriptions are gone
	// Send again an event and check subscriber, check subscriober should fail after 5 retries
	err = handlers.SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	g.Expect(err).NotTo(gomega.BeNil())
}
