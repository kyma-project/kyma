package nats

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/onsi/gomega"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager/mock"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func TestCleanup(t *testing.T) {
	natsSubMgr := mock.Manager{}
	g := gomega.NewWithT(t)
	data := "sampledata"
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)
	subscriberStoreURL, subscriberCheckURL := "", ""

	// When
	// Create a test subscriber
	ctx := context.Background()
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

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(gomega.BeNil())

	envConf := env.NatsConfig{
		URL:             natsURL,
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: controllertesting.EventTypePrefix,
	}
	subsConfig := env.DefaultSubscriptionConfig{MaxInFlightMessages: 9}
	natsBackend := handlers.NewNats(envConf, subsConfig, defaultLogger)
	natsSubMgr.Backend = natsBackend
	err = natsSubMgr.Backend.Initialize(env.Config{})
	g.Expect(err).To(gomega.BeNil())

	// Create fake Dynamic clients
	fakeClient, err := controllertesting.NewFakeSubscriptionClient(testSub)
	g.Expect(err).To(gomega.BeNil())
	natsSubMgr.Client = fakeClient

	fakeCleaner := mock.Cleaner{}

	// Create NATS subscription
	_, err = natsSubMgr.Backend.SyncSubscription(testSub, &fakeCleaner)
	g.Expect(err).To(gomega.BeNil())

	// Make sure subscriber works
	err = subscriber.CheckEvent("", subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}
	// Send an event
	err = handlers.SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}

	// Check for the event
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}

	// Then
	err = cleanup(natsSubMgr.Backend, natsSubMgr.Client, defaultLogger.WithContext())
	g.Expect(err).To(gomega.BeNil())

	// Expect
	unstructuredSub, err := natsSubMgr.Client.Resource(controllertesting.SubscriptionGroupVersionResource()).Namespace("test").Get(ctx, testSub.Name, metav1.GetOptions{})
	g.Expect(err).To(gomega.BeNil())
	gotSub, err := controllertesting.ToSubscription(unstructuredSub)
	g.Expect(err).To(gomega.BeNil())
	expectedSubStatus := eventingv1alpha1.SubscriptionStatus{}
	g.Expect(expectedSubStatus).To(gomega.Equal(gotSub.Status))

	// Test NATS subscriptions are gone
	// Send again an event and check subscriber, check subscriber should fail after 5 retries
	err = handlers.SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	err = subscriber.CheckEvent(expectedDataInStore, subscriberCheckURL)
	g.Expect(err).NotTo(gomega.BeNil())
}
