//go:build integration
// +build integration

package nats

import (
	"context"
	"fmt"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/core"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

type natsSubMgrMock struct {
	Client  dynamic.Interface
	Backend core.Backend
}

func (c *natsSubMgrMock) Init(_ manager.Manager) error {
	return nil
}

func (c *natsSubMgrMock) Start(_ env.DefaultSubscriptionConfig, _ subscriptionmanager.Params) error {
	return nil
}

func (c *natsSubMgrMock) Stop(_ bool) error {
	return nil
}

func TestCleanup(t *testing.T) {
	natsSubMgr := natsSubMgrMock{}
	g := gomega.NewWithT(t)
	data := "sampledata"
	expectedDataInStore := fmt.Sprintf("\"%s\"", data)

	// When
	// Create a test subscriber
	ctx := context.Background()
	subscriber := controllertesting.NewSubscriber()
	// Shutting down subscriber
	defer subscriber.Shutdown()

	// Create NATS Server
	natsPort := 4222
	natsServer := controllertesting.RunNatsServerOnPort(controllertesting.WithPort(natsPort))
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
	natsBackend := core.NewNats(envConf, subsConfig, metrics.NewCollector(), defaultLogger)
	natsSubMgr.Backend = natsBackend
	err = natsSubMgr.Backend.Initialize(nil)
	g.Expect(err).To(gomega.BeNil())

	// Create test subscription
	testSub := controllertesting.NewSubscription("test", "test",
		controllertesting.WithFakeSubscriptionStatus(),
		controllertesting.WithOrderCreatedFilter(),
		controllertesting.WithSinkURL(subscriber.SinkURL),
		controllertesting.WithStatusConfig(subsConfig),
	)

	// Create fake Dynamic clients
	fakeClient, err := controllertesting.NewFakeSubscriptionClient(testSub)
	g.Expect(err).To(gomega.BeNil())
	natsSubMgr.Client = fakeClient

	cleaner := func(et string) (string, error) {
		return et, nil
	}
	testSub.Status.CleanEventTypes, _ = backendnats.GetCleanSubjects(testSub, eventtype.CleanerFunc(cleaner))

	// Create NATS subscription
	err = natsSubMgr.Backend.SyncSubscription(testSub)
	g.Expect(err).To(gomega.BeNil())

	// Make sure subscriber works
	err = subscriber.CheckEvent("")
	if err != nil {
		t.Fatalf("subscriber did not receive the event: %v", err)
	}
	// Send an event
	err = core.SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}

	// Check for the event
	err = subscriber.CheckEvent(expectedDataInStore)
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
	expectedSubStatus := eventingv1alpha1.SubscriptionStatus{CleanEventTypes: []string{}}
	g.Expect(expectedSubStatus).To(gomega.Equal(gotSub.Status))

	// Test NATS subscriptions are gone
	// Send again an event and check subscriber, check subscriber should fail after 5 retries
	err = core.SendEventToNATS(natsBackend, data)
	if err != nil {
		t.Fatalf("publish event failed: %v", err)
	}
	err = subscriber.CheckEvent(expectedDataInStore)
	g.Expect(err).NotTo(gomega.BeNil())
}
