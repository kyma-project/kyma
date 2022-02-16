package nats_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/events"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats.go"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	//logf "sigs.k8s.io/controller-runtime/pkg/log"
	//"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"k8s.io/client-go/kubernetes/scheme"

	natsreconciler "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/nats"
	natsserver "github.com/nats-io/nats-server/v2/server"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	natsPort = 4221

	smallTimeout         = 10 * time.Second
	smallPollingInterval = 1 * time.Second

	//todo can this be deleted?
	timeout         = 60 * time.Second
	pollingInterval = 5 * time.Second

	namespaceName          = "test"
	subscriptionNameFormat = "nats-sub-%d"
	subscriberNameFormat   = "subscriber-%d"
)

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

type testEnsemble struct {
	//todo order,
	//todo explain components
	testID                    int
	cfg                       *rest.Config
	k8sClient                 client.Client
	testEnv                   *envtest.Environment
	natsServer                *natsserver.Server
	reconciler                *natsreconciler.Reconciler
	natsBackend               *handlers.Nats
	defaultSubscriptionConfig env.DefaultSubscriptionConfig
	subscriberSvc             *v1.Service
	cancel                    context.CancelFunc
}

type expect struct {
	k8sSubscription  []gomegatypes.GomegaMatcher
	k8sEvents        []v1.Event
	natsSubscription []gomegatypes.GomegaMatcher
}

type changeFunc func(subscription eventingv1alpha1.Subscription)

//todo description
func TestCreateSubscription(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g)
	defer ens.cancel()

	var testCases = []struct {
		name               string
		subscriptionOpts   []reconcilertesting.SubscriptionOpt
		expect             expect
		shouldTestDeletion bool // a bool in golang is false by default
	}{
		{
			name: "create and delete",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens)),
				},
				natsSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.BeNotNil(),
					reconcilertesting.BeValid(),
					reconcilertesting.HaveSubject(reconcilertesting.OrderCreatedEventType),
				},
			},
			shouldTestDeletion: true,
		},

		{
			name: "empty event type",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, ""),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						eventingv1alpha1.MakeCondition(
							eventingv1alpha1.ConditionSubscriptionActive,
							eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
							v1.ConditionFalse, nats.ErrBadSubject.Error(),
						),
					),
				},
			},
		},

		{
			name: "invalid sink; misses 'http' and 'https'",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL("invalid"),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionInvalidSink("sink URL scheme should be 'http' or 'https'")),
				},
				k8sEvents: []v1.Event{eventInvalidSink("Sink URL scheme should be HTTP or HTTPS: invalid")},
			},
		},

		{
			name: "invalid sink; invalid character",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL("http://127.0.0. 1"),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						conditionInvalidSink("not able to parse sink url with error: parse \"http://127.0.0. 1\": invalid character \" \" in host name")),
				},
				k8sEvents: []v1.Event{
					eventInvalidSink("Not able to parse Sink URL with error: parse \"http://127.0.0. 1\": invalid character \" \" in host name")},
			},
		},

		{
			name: "invalid sink; missing suffix 'svc.cluster.local'",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL("http://127.0.0.1"),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						conditionInvalidSink("sink does not contain suffix: svc.cluster.local in the URL")),
				},
				k8sEvents: []v1.Event{
					eventInvalidSink("Sink does not contain suffix: svc.cluster.local")},
			},
		},

		{
			name: "invalid sink; too many sub domains",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL(fmt.Sprintf("https://%s.%s.%s.svc.cluster.local", "testapp", "testsub", "test")),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						conditionInvalidSink("sink should contain 5 sub-domains: testapp.testsub.test.svc.cluster.local")),
				},
				k8sEvents: []v1.Event{
					eventInvalidSink("Sink should contain 5 sub-domains: testapp.testsub.test.svc.cluster.local")},
			},
		},

		{
			name: "invalid sink; wrong namespace",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL(fmt.Sprintf("https://%s.%s.svc.cluster.local", "testapp", "wrong-ns")),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						conditionInvalidSink("namespace of subscription: test and the namespace of subscriber: wrong-ns are different")),
				},
				k8sEvents: []v1.Event{
					eventInvalidSink("natsNamespace of subscription: test and the subscriber: wrong-ns are different")},
			},
		},

		{
			name: "invalid sink; not a valid cluster local service",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL(
					reconcilertesting.ValidSinkURL(ens.subscriberSvc.Namespace, "testapp")),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						conditionInvalidSink("sink is not valid cluster local svc, failed with error: Service \"testapp\" not found")),
				},
				k8sEvents: []v1.Event{
					eventInvalidSink("Sink does not correspond to a valid cluster local svc")},
			},
		},

		{
			name: "valid sink",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter("", reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURL(validSinkURL(ens)),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubscriptionReady(),
				},
			},
			shouldTestDeletion: true,
		},

		{
			name: "valid sink; with port",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter("", reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURL(validSinkURL(ens, ":8080")),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubscriptionReady(),
				},
			},
			shouldTestDeletion: true,
		},

		{
			name: "valid sink; with port and endpoint",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter("", reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURL(validSinkURL(ens, ":8080", "/myEndpoint")),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubscriptionReady(),
				},
			},
			shouldTestDeletion: true,
		},

		{
			name: "empty protocol, protocol setting and dialect",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter("", reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			expect: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
				},
				natsSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.BeNotNil(),
					reconcilertesting.BeValid(),
					reconcilertesting.HaveSubject(reconcilertesting.OrderCreatedEventType),
				},
			},
			shouldTestDeletion: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subscription, subscriptionName := createSubscription(ctx, g, ens, tc.subscriptionOpts...)

			testSubscriptionOnK8s(ctx, g, ens, subscription, subscriptionName, tc.expect.k8sSubscription...)
			testEventsOnK8s(ctx, g, ens, tc.expect.k8sEvents...)
			testSubscriptionOnNATS(g, ens, subscriptionName, tc.expect.natsSubscription...)
			testDeletionOnK8s(ctx, g, ens, subscription, tc.shouldTestDeletion)
		})
	}
}

//todo description
func TestChangeSubscription(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g)
	defer ens.cancel()

	var testCases = []struct {
		name               string
		subscriptionOpts   []reconcilertesting.SubscriptionOpt
		expectBefore       expect
		changeSubscription func(subscription *eventingv1alpha1.Subscription)
		expectAfter        expect
		shouldTestDeletion bool
	}{
		{
			name: "clean event types; add filters to subscription without filters",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithEmptyFilter(),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			expectBefore: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens)),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				eventTypes := []string{
					uncleanEventType("0"),
					uncleanEventType("1"),
				}
				for _, eventType := range eventTypes {
					reconcilertesting.AddFilter(reconcilertesting.EventSource, eventType, subscription)
				}
			},
			expectAfter: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens)),
					reconcilertesting.HaveCleanEventTypes([]string{
						cleanEventType("0"),
						cleanEventType("1"),
					}),
				},
			},
		},

		{
			name: "clean event types; change filters",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, uncleanEventType("0")),
				reconcilertesting.WithFilter(reconcilertesting.EventSource, uncleanEventType("1")),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			expectBefore: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens)),
					reconcilertesting.HaveCleanEventTypes([]string{
						cleanEventType("0"),
						cleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				//change all the filters by adding "alpha" to the event type
				for _, f := range subscription.Spec.Filter.Filters {
					f.EventType.Value = fmt.Sprintf("%salpha", f.EventType.Value)
				}
			},
			expectAfter: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens)),
					reconcilertesting.HaveCleanEventTypes([]string{
						cleanEventType("0alpha"),
						cleanEventType("1alpha"),
					}),
				},
			},
		},

		{
			name: "clean event types; delete a filter",
			subscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, uncleanEventType("0")),
				reconcilertesting.WithFilter(reconcilertesting.EventSource, uncleanEventType("1")),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			expectBefore: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens)),
					reconcilertesting.HaveCleanEventTypes([]string{
						cleanEventType("0"),
						cleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				subscription.Spec.Filter.Filters = subscription.Spec.Filter.Filters[:1]
			},
			expectAfter: expect{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionValidSubscription("")),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens)),
					reconcilertesting.HaveCleanEventTypes([]string{
						cleanEventType("0"),
					}),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subscription, subscriptionName := createSubscription(ctx, g, ens, tc.subscriptionOpts...)

			testSubscriptionOnK8s(ctx, g, ens, subscription, subscriptionName, tc.expectBefore.k8sSubscription...)
			testEventsOnK8s(ctx, g, ens, tc.expectBefore.k8sEvents...)
			testSubscriptionOnNATS(g, ens, subscriptionName, tc.expectBefore.natsSubscription...)

			tc.changeSubscription(subscription)
			updateSubscriptionOnK8s(ctx, g, ens, subscription)

			testSubscriptionOnK8s(ctx, g, ens, subscription, subscriptionName, tc.expectAfter.k8sSubscription...)
			testEventsOnK8s(ctx, g, ens, tc.expectAfter.k8sEvents...)
			testSubscriptionOnNATS(g, ens, subscriptionName, tc.expectAfter.natsSubscription...)
			testDeletionOnK8s(ctx, g, ens, subscription, tc.shouldTestDeletion)
		})
	}
}

func updateSubscriptionOnK8s(ctx context.Context, g *gomega.GomegaWithT, ens *testEnsemble, subscription *eventingv1alpha1.Subscription) {
	err := ens.k8sClient.Update(ctx, subscription)
	g.Expect(err).Should(gomega.BeNil())
}

func createSubscription(ctx context.Context, g *gomega.GomegaWithT, ens *testEnsemble, subscriptionOpts ...reconcilertesting.SubscriptionOpt) (*eventingv1alpha1.Subscription, string) {
	subscriptionName := fmt.Sprintf(subscriptionNameFormat, ens.testID)
	ens.testID++
	subscription := reconcilertesting.NewSubscription(subscriptionName, ens.subscriberSvc.Namespace, subscriptionOpts...)
	subscription = createSubscriptionInK8s(ctx, ens, subscription, g)
	return subscription, subscriptionName
}

func testSubscriptionOnK8s(ctx context.Context, g *gomega.GomegaWithT, ens *testEnsemble, subscription *eventingv1alpha1.Subscription, subscriptionName string, expectations ...gomegatypes.GomegaMatcher) {
	subExpectations := append(expectations, reconcilertesting.HaveSubscriptionName(subscriptionName))
	getSubscriptionOnK8S(ctx, ens, subscription, g).Should(gomega.And(subExpectations...))
}

func testEventsOnK8s(ctx context.Context, g *gomega.GomegaWithT, ens *testEnsemble, expectations ...v1.Event) {
	for _, event := range expectations { //todo replace with gomega.And(events)
		getK8sEvents(ctx, ens, g).Should(reconcilertesting.HaveEvent(event))
	}
}

func testSubscriptionOnNATS(g *gomega.GomegaWithT, ens *testEnsemble, subscriptionName string, expectations ...gomegatypes.GomegaMatcher) {
	getSubscriptionFromNATS(ens.natsBackend, subscriptionName, g).Should(gomega.And(expectations...))
}

func testDeletionOnK8s(ctx context.Context, g *gomega.GomegaWithT, ens *testEnsemble, subscription *eventingv1alpha1.Subscription, shouldTest bool) {
	if shouldTest {
		g.Expect(ens.k8sClient.Delete(ctx, subscription)).Should(gomega.BeNil())
		isSubscriptionDeleted(ctx, ens, subscription, g).Should(reconcilertesting.HaveNotFoundSubscription(true))
	}
}

func validSinkURL(ens *testEnsemble, additions ...string) string {
	url := reconcilertesting.ValidSinkURL(ens.subscriberSvc.Namespace, ens.subscriberSvc.Name)
	for _, a := range additions {
		url = fmt.Sprintf("%s%s", url, a)
	}
	return url
}

func uncleanEventType(suffix string) string {
	return fmt.Sprintf("%s%s", reconcilertesting.OrderCreatedEventTypeNotClean, suffix)
}

func cleanEventType(suffix string) string {
	return fmt.Sprintf("%s%s", reconcilertesting.OrderCreatedEventType, suffix)
}

// isSubscriptionDeleted checks a subscription is deleted and allows making assertions on it
func isSubscriptionDeleted(ctx context.Context, ens *testEnsemble, subscription *eventingv1alpha1.Subscription,
	g *gomega.GomegaWithT) gomega.AsyncAssertion {
	return g.Eventually(func() bool {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return k8serrors.IsNotFound(err)
		}
		return false
	}, smallTimeout, smallPollingInterval)
}

func configDefault(ens *testEnsemble) *eventingv1alpha1.SubscriptionConfig {
	return &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: ens.defaultSubscriptionConfig.MaxInFlightMessages}
}

func conditionValidSubscription(msg string) eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(
		eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		v1.ConditionTrue, msg)
}

func conditionInvalidSink(msg string) eventingv1alpha1.Condition { //todo should this be reconcilertesting
	return eventingv1alpha1.MakeCondition(
		eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionActive,
		v1.ConditionFalse, msg)
}

func eventInvalidSink(msg string) v1.Event { // todo should this be in reoncilertesting
	return v1.Event{
		Reason:  string(events.ReasonValidationFailed),
		Message: msg,
		Type:    v1.EventTypeWarning,
	}
}

func setupTestEnsemble(ctx context.Context, eventTypePrefix string, g *gomega.GomegaWithT) *testEnsemble {
	useExistingCluster := useExistingCluster
	ens := &testEnsemble{
		defaultSubscriptionConfig: env.DefaultSubscriptionConfig{
			MaxInFlightMessages:   1,
			DispatcherRetryPeriod: time.Second,
			DispatcherMaxRetries:  1,
		},
		natsServer: startNATS(natsPort),
		testEnv: &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join("../../../", "config", "crd", "bases"),
				filepath.Join("../../../", "config", "crd", "external"),
			},
			AttachControlPlaneOutput: attachControlPlaneOutput,
			UseExistingCluster:       &useExistingCluster,
		},
	}

	startTestEnv(ens, g)
	startReconciler(eventTypePrefix, ens, g)
	startSubscriberSvc(ctx, ens, g)
	return ens
}

func startTestEnv(ens *testEnsemble, g *gomega.GomegaWithT) {
	k8sCfg, err := ens.testEnv.Start()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(k8sCfg).ToNot(gomega.BeNil())
	ens.cfg = k8sCfg
}

func startNATS(port int) *natsserver.Server {
	natsServer := reconcilertesting.RunNatsServerOnPort(port)
	log.Printf("NATS server started %v", natsServer.ClientURL())
	return natsServer
}

func startReconciler(eventTypePrefix string, ens *testEnsemble, g *gomega.GomegaWithT) *testEnsemble {
	ctx, cancel := context.WithCancel(context.Background())
	ens.cancel = cancel
	//logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter))) //todo

	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(ens.cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: "localhost:7070",
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	envConf := env.NatsConfig{
		URL:             ens.natsServer.ClientURL(),
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: eventTypePrefix,
	}

	// prepare application-lister
	app := applicationtest.NewApplication(reconcilertesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), app)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(gomega.BeNil())

	ens.reconciler = natsreconciler.NewReconciler(
		ctx,
		k8sManager.GetClient(),
		applicationLister,
		defaultLogger,
		k8sManager.GetEventRecorderFor("eventing-controller-nats"),
		envConf,
		ens.defaultSubscriptionConfig,
	)

	err = ens.reconciler.SetupUnmanaged(k8sManager)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	ens.natsBackend = ens.reconciler.Backend.(*handlers.Nats)

	go func() { //todo is this func needed any longer?
		err = k8sManager.Start(ctx)
		g.Expect(err).ToNot(gomega.HaveOccurred())
	}()

	ens.k8sClient = k8sManager.GetClient()
	g.Expect(ens.k8sClient).ToNot(gomega.BeNil())

	return ens
}

func startSubscriberSvc(ctx context.Context, ens *testEnsemble, g *gomega.GomegaWithT) {
	ens.subscriberSvc = reconcilertesting.NewSubscriberSvc("test-subscriber", "test")
	createSubscriberSvcInK8s(ctx, ens, g)
}

func createSubscriberSvcInK8s(ctx context.Context, ens *testEnsemble, g *gomega.GomegaWithT) {
	//if the namespace is not "default" create it on the cluster
	if ens.subscriberSvc.Namespace != "default " {
		namespace := fixtureNamespace(ens.subscriberSvc.Namespace)
		if namespace.Name != "default" {
			err := ens.k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	// create subscriber svc on cluster
	err := ens.k8sClient.Create(ctx, ens.subscriberSvc)
	g.Expect(err).Should(gomega.BeNil())
}

func createSubscriptionInK8s(ctx context.Context, ens *testEnsemble, subscription *eventingv1alpha1.Subscription,
	g *gomega.GomegaWithT) *eventingv1alpha1.Subscription {
	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := ens.k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	// create subscription on cluster
	err := ens.k8sClient.Create(ctx, subscription)
	g.Expect(err).Should(gomega.BeNil())
	return subscription
}

func fixtureNamespace(name string) *v1.Namespace {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "natsNamespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

// getSubscription fetches a subscription using the lookupKey and allows making assertions on it
func getSubscriptionOnK8S(ctx context.Context, ens *testEnsemble, subscription *eventingv1alpha1.Subscription,
	g *gomega.GomegaWithT, intervals ...interface{}) gomega.AsyncAssertion {
	if len(intervals) == 0 {
		intervals = []interface{}{smallTimeout, smallPollingInterval}
	}
	return g.Eventually(func() *eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}
		}
		return subscription
	}, intervals...)
}

// getK8sEvents returns all kubernetes events for the given namespace.
// The result can be used in a gomega assertion.
func getK8sEvents(ctx context.Context, ens *testEnsemble, g *gomega.GomegaWithT) gomega.AsyncAssertion {
	eventList := v1.EventList{}
	return g.Eventually(func() v1.EventList {
		err := ens.k8sClient.List(ctx, &eventList, client.InNamespace(ens.subscriberSvc.Namespace))
		if err != nil {
			return v1.EventList{}
		}
		return eventList
	}, smallTimeout, smallPollingInterval)
}

func getSubscriptionFromNATS(natsHandler *handlers.Nats, subscriptionName string, g *gomega.GomegaWithT) gomega.Assertion {
	return g.Expect(func() *nats.Subscription {
		subscriptions := natsHandler.GetAllSubscriptions()
		for key, subscription := range subscriptions {
			//the key does NOT ONLY contain the subscription name
			if strings.Contains(key, subscriptionName) {
				return subscription
			}
		}
		return nil
	}()) // todo do we need func()*nats.Subscription{}()? can we remove func?
}
