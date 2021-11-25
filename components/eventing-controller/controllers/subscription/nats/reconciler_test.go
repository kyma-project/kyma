package nats

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	natsPort = 4221

	smallTimeout         = 5 * time.Second
	smallPollingInterval = 1 * time.Second

	timeout         = 60 * time.Second
	pollingInterval = 5 * time.Second

	namespaceName          = "test"
	subscriptionNameFormat = "sub-%d"
)

type testCase func(id int, natsSubjectToPublish, eventTypeToSubscribe string) bool

var (
	testCases = []testCase{
		testCreateDeleteSubscription,
		testCreateSubscriptionWithInvalidSink,
		testCreateSubscriptionWithEmptyProtocolProtocolSettingsDialect,
		testChangeSubscriptionConfiguration,
		testCreateSubscriptionWithEmptyEventType,
	}
)

func testCreateDeleteSubscription(id int, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Create/Delete Subscription", func() {
		It("Should create/delete NATS Subscription", func() {
			ctx := context.Background()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)

			// create subscriber
			result := make(chan []byte)
			url, shutdown := newSubscriber(result)
			defer shutdown()

			// create subscription
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe), reconcilertesting.WithWebhookForNats)
			subscription.Spec.Sink = url
			ensureSubscriptionCreated(ctx, subscription)

			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
				reconcilertesting.HaveSubsConfiguration(&eventingv1alpha1.SubscriptionConfig{
					MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages,
				}),
			))

			// publish a message
			connection, err := connectToNats(natsURL)
			Expect(err).ShouldNot(HaveOccurred())
			err = connection.Publish(natsSubjectToPublish, []byte(reconcilertesting.StructuredCloudEvent))
			Expect(err).ShouldNot(HaveOccurred())

			// make sure that the subscriber received the message
			sent := fmt.Sprintf(`"%s"`, reconcilertesting.EventData)
			Eventually(func() ([]byte, error) {
				return getFromChanOrTimeout(result, smallPollingInterval)
			}, timeout, pollingInterval).Should(WithTransform(bytesStringer, Equal(sent)))

			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
			isSubscriptionDeleted(ctx, subscription).Should(reconcilertesting.HaveNotFoundSubscription(true))
		})
	})
}

func testCreateSubscriptionWithInvalidSink(id int, _, eventTypeToSubscribe string) bool {
	return When("Create Subscription with invalid sink", func() {
		It("Should mark the Subscription as not ready", func() {
			ctx := context.Background()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe), reconcilertesting.WithWebhookForNats)
			givenSubscription.Spec.Sink = "invalid"
			ensureSubscriptionCreated(ctx, givenSubscription)

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, "parse \"invalid\": invalid URI for request")),
			))
		})
	})
}

func testCreateSubscriptionWithEmptyProtocolProtocolSettingsDialect(id int, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Create Subscription with empty protocol, protocolsettings and dialect", func() {
		It("Should mark the Subscription as ready", func() {
			ctx := context.Background()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)

			// create subscriber
			result := make(chan []byte)
			url, shutdown := newSubscriber(result)
			defer shutdown()

			// create subscription
			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter("", eventTypeToSubscribe))
			subscription.Spec.Sink = url
			ensureSubscriptionCreated(ctx, subscription)

			getSubscription(ctx, subscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
			))

			// publish a message
			connection, err := connectToNats(natsURL)
			Expect(err).ShouldNot(HaveOccurred())
			err = connection.Publish(natsSubjectToPublish, []byte(reconcilertesting.StructuredCloudEvent))
			Expect(err).ShouldNot(HaveOccurred())

			// make sure that the subscriber received the message
			sent := fmt.Sprintf(`"%s"`, reconcilertesting.EventData)
			Eventually(func() ([]byte, error) {
				return getFromChanOrTimeout(result, smallPollingInterval)
			}, timeout, pollingInterval).Should(WithTransform(bytesStringer, Equal(sent)))
		})
	})
}

func testChangeSubscriptionConfiguration(id int, natsSubjectToPublish, eventTypeToSubscribe string) bool {
	return When("Change Subscription configuration", func() {
		It("Should reflect the new config in the subscription status", func() {
			By("Creating the subscription using the default config")
			ctx := context.Background()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)

			// create subscriber
			result := make(chan []byte)
			url, shutdown := newSubscriber(result)
			defer shutdown()

			// create subscription
			sub := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, eventTypeToSubscribe), reconcilertesting.WithWebhookForNats)
			sub.Spec.Sink = url
			ensureSubscriptionCreated(ctx, sub)

			getSubscription(ctx, sub).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionTrue, "")),
				reconcilertesting.HaveSubsConfiguration(&eventingv1alpha1.SubscriptionConfig{
					MaxInFlightMessages: defaultSubsConfig.MaxInFlightMessages,
				}),
			))

			By("Updating the subscription configuration in the spec")

			newMaxInFlight := defaultSubsConfig.MaxInFlightMessages + 1
			changedSub := sub.DeepCopy()
			changedSub.Spec.Config = &eventingv1alpha1.SubscriptionConfig{
				MaxInFlightMessages: newMaxInFlight,
			}
			Expect(k8sClient.Update(ctx, changedSub)).Should(BeNil())

			Eventually(subscriptionGetter(ctx, sub.Name, sub.Namespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName),
					reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
						eventingv1alpha1.ConditionSubscriptionActive,
						eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
						v1.ConditionTrue, ""),
					),
					reconcilertesting.HaveSubsConfiguration(&eventingv1alpha1.SubscriptionConfig{
						MaxInFlightMessages: newMaxInFlight,
					}),
				))

			connection, err := connectToNats(natsURL)
			Expect(err).ShouldNot(HaveOccurred())
			toSend := fmt.Sprintf(`"%s"`, reconcilertesting.EventData)
			msgData := []byte(reconcilertesting.StructuredCloudEvent)
			// Publish the message and make sure that the subscriber received the message.
			// Since the subscriptions on NATS are removed and recreated due to the MaxInFlight change, it could be
			// that an immediate publication after the subscription change is not delivered, therefore, we try
			// sending/receiving in an Eventually block.
			Eventually(func() ([]byte, error) {
				if err = connection.Publish(natsSubjectToPublish, msgData); err != nil {
					return nil, err
				}
				return getFromChanOrTimeout(result, smallPollingInterval)
			}, timeout, pollingInterval).Should(WithTransform(bytesStringer, Equal(toSend)))

			Expect(k8sClient.Delete(ctx, sub)).Should(BeNil())
			isSubscriptionDeleted(ctx, sub).Should(reconcilertesting.HaveNotFoundSubscription(true))
		})
	})
}

func testCreateSubscriptionWithEmptyEventType(id int, _, _ string) bool {
	return When("Create Subscription with empty event type", func() {
		It("Should mark the subscription as not ready", func() {
			ctx := context.Background()
			subscriptionName := fmt.Sprintf(subscriptionNameFormat, id)

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithFilter(reconcilertesting.EventSource, ""), reconcilertesting.WithWebhookForNats)
			reconcilertesting.WithValidSink("foo", "bar", givenSubscription)
			ensureSubscriptionCreated(ctx, givenSubscription)

			getSubscription(ctx, givenSubscription).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(eventingv1alpha1.MakeCondition(
					eventingv1alpha1.ConditionSubscriptionActive,
					eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
					v1.ConditionFalse, nats.ErrBadSubject.Error())),
			))
		})
	})
}

var (
	_ = Describe("NATS Subscription reconciler tests with non-empty eventTypePrefix", reconcilerTestsExecutor(reconcilertesting.EventTypePrefix, reconcilertesting.OrderCreatedEventType, reconcilertesting.OrderCreatedEventTypeNotClean))
	_ = Describe("NATS Subscription reconciler tests with empty eventTypePrefix", reconcilerTestsExecutor(reconcilertesting.EventTypePrefixEmpty, reconcilertesting.OrderCreatedEventTypePrefixEmpty, reconcilertesting.OrderCreatedEventTypeNotCleanPrefixEmpty))
)

func reconcilerTestsExecutor(eventTypePrefix, natsSubjectToPublish, eventTypeToSubscribe string) func() {
	return func() {
		var cancel context.CancelFunc

		BeforeEach(func() {
			cancel = startReconciler(eventTypePrefix)
		})

		AfterEach(func() {
			cancel()
		})

		for _, tc := range testCases {
			tc(testID, natsSubjectToPublish, eventTypeToSubscribe)
			testID++
		}
	}
}

func ensureSubscriptionCreated(ctx context.Context, subscription *eventingv1alpha1.Subscription) {
	By(fmt.Sprintf("Ensuring the test namespace %q is created", subscription.Namespace))
	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				Expect(err).ShouldNot(HaveOccurred())
			}
		}
	}

	By(fmt.Sprintf("Ensuring the subscription %q is created", subscription.Name))
	// create subscription
	err := k8sClient.Create(ctx, subscription)
	Expect(err).Should(BeNil())
}

func fixtureNamespace(name string) *v1.Namespace {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

func subscriptionGetter(ctx context.Context, name, namespace string) func() (*eventingv1alpha1.Subscription, error) {
	return func() (*eventingv1alpha1.Subscription, error) {
		lookupKey := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}
		subscription := &eventingv1alpha1.Subscription{}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}, err
		}
		return subscription, nil
	}
}

// getSubscription fetches a subscription using the lookupKey and allows making assertions on it
func getSubscription(ctx context.Context, subscription *eventingv1alpha1.Subscription) AsyncAssertion {
	return Eventually(func() *eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}
		}
		return subscription
	}, smallTimeout, smallPollingInterval)
}

// isSubscriptionDeleted checks a subscription is deleted and allows making assertions on it
func isSubscriptionDeleted(ctx context.Context, subscription *eventingv1alpha1.Subscription) AsyncAssertion {
	return Eventually(func() bool {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return k8serrors.IsNotFound(err)
		}
		return false
	}, smallTimeout, smallPollingInterval)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test Suite setup ////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// These tests use Ginkgo (BDD-style Go controllertesting framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// TODO: make configurable
const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

var testID int
var natsURL string
var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var natsServer *natsserver.Server
var defaultSubsConfig = env.DefaultSubscriptionConfig{MaxInFlightMessages: 1}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t, "NATS Controller Suite", []Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	By("bootstrapping test environment")
	natsServer, natsURL = startNATS(natsPort)
	useExistingCluster := useExistingCluster
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../../", "config", "crd", "bases"),
			filepath.Join("../../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	reconcilertesting.ShutDownNATSServer(natsServer)
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
}, 60)

func startNATS(port int) (*natsserver.Server, string) {
	natsServer := reconcilertesting.RunNatsServerOnPort(port)
	clientURL := natsServer.ClientURL()
	log.Printf("NATS server started %v", clientURL)
	return natsServer, clientURL
}

func startReconciler(eventTypePrefix string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: ":7070",
	})
	Expect(err).ToNot(HaveOccurred())

	envConf := env.NatsConfig{
		URL:             natsURL,
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: eventTypePrefix,
	}

	// prepare application-lister
	app := applicationtest.NewApplication(reconcilertesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), app)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	Expect(err).To(BeNil())

	err = NewReconciler(
		ctx,
		k8sManager.GetClient(),
		applicationLister,
		k8sManager.GetCache(),
		defaultLogger,
		k8sManager.GetEventRecorderFor("eventing-controller-nats"),
		envConf,
		defaultSubsConfig,
	).SetupUnmanaged(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	return cancel
}

func connectToNats(natsURL string) (*nats.Conn, error) {
	connection, err := nats.Connect(natsURL, nats.RetryOnFailedConnect(true), nats.MaxReconnects(3), nats.ReconnectWait(time.Second))
	if err != nil {
		return nil, err
	}
	if connection.Status() != nats.CONNECTED {
		return nil, err
	}
	return connection, nil
}

func newSubscriber(result chan []byte) (string, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		result <- body
	}))
	return server.URL, server.Close
}

func getFromChanOrTimeout(ch <-chan []byte, t time.Duration) ([]byte, error) {
	select {
	case received := <-ch:
		return received, nil
	case <-time.After(t):
		return nil, fmt.Errorf("timed out waiting for a message")
	}
}

func bytesStringer(bs []byte) string {
	return string(bs)
}
