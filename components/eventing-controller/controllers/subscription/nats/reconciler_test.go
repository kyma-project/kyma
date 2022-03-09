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
	natsreconciler "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/testing/nats"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	smallTimeout           = 10 * time.Second
	smallPollingInterval   = 1 * time.Second
	subscriptionNameFormat = "nats-sub-%d"

	useExistingCluster       = false
	attachControlPlaneOutput = false
	emptyEventSource         = ""
)

type testEnsemble struct {
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
	ctx                       context.Context
	g                         *gomega.GomegaWithT
}

type want struct {
	k8sSubscription  []gomegatypes.GomegaMatcher
	k8sEvents        []v1.Event
	natsSubscription []gomegatypes.GomegaMatcher
}

// TestUnavailableNATSServer tests if a subscription is reconciled properly when the NATS backend is unavailable.
func TestUnavailableNATSServer(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g, natsPort)

	subscription := createSubscription(ens,
		reconcilertesting.WithFilter(emptyEventSource, newUncleanEventType("")),
		reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
	)
	testSubscriptionOnK8s(ens, subscription,
		reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
		reconcilertesting.HaveSubscriptionReady(),
		reconcilertesting.HaveCleanEventTypes([]string{newCleanEventType("")}),
	)

	ens.natsServer.Shutdown()
	testSubscriptionOnK8s(ens, subscription,
		reconcilertesting.HaveSubscriptionNotReady(),
	)

	ens.natsServer = startNATS(natsPort)
	testSubscriptionOnK8s(ens, subscription, reconcilertesting.HaveSubscriptionReady())

	t.Cleanup(ens.cancel)
}

// TestCreateSubscription tests if subscriptions get created properly by the reconciler.
func TestCreateSubscription(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(gomega.HaveOccurred())

	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g, natsPort)

	var testCases = []struct {
		name                  string
		givenSubscriptionOpts []reconcilertesting.SubscriptionOpt
		want                  want
	}{
		{
			name: "create and delete",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			want: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
				},
				natsSubscription: []gomegatypes.GomegaMatcher{
					natstesting.BeExistingSubscription(),
					natstesting.BeValidSubscription(),
					natstesting.BeSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
				},
			},
		},
		{
			name: "filter with empty event type",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, ""),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			want: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveConditionBadSubject(),
				},
			},
		},
		{
			name: "invalid sink; misses 'http' and 'https'",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL("invalid"),
			},
			want: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(conditionInvalidSink("sink URL scheme should be 'http' or 'https'")),
				},
				k8sEvents: []v1.Event{eventInvalidSink("Sink URL scheme should be HTTP or HTTPS: invalid")},
			},
		},
		{
			name: "invalid sink; invalid character",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL("http://127.0.0. 1"),
			},
			want: want{
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
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL("http://127.0.0.1"),
			},
			want: want{
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
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL(fmt.Sprintf("https://%s.%s.%s.svc.cluster.local", "testapp", "testsub", "test")),
			},
			want: want{
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
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL(fmt.Sprintf("https://%s.%s.svc.cluster.local", "testapp", "wrong-ns")),
			},
			want: want{
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
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL(
					reconcilertesting.ValidSinkURL(ens.subscriberSvc.Namespace, "testapp")),
			},
			want: want{
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
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURL(validSinkURL(ens)),
			},
			want: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubscriptionReady(),
				},
			},
		},
		{
			name: "valid sink; with port",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURL(validSinkURL(ens, ":8080")),
			},
			want: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
				},
			},
		},
		{
			name: "valid sink; with port and endpoint",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURL(validSinkURL(ens, ":8080", "/myEndpoint")),
			},
			want: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubscriptionReady(),
				},
			},
		},
		{
			name: "empty protocol, protocol setting and dialect",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			want: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
				},
				natsSubscription: []gomegatypes.GomegaMatcher{
					natstesting.BeExistingSubscription(),
					natstesting.BeValidSubscription(),
					natstesting.BeSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
				},
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			subscription := createSubscription(ens, tc.givenSubscriptionOpts...)
			testSubscriptionOnK8s(ens, subscription, tc.want.k8sSubscription...)
			testEventsOnK8s(ens, tc.want.k8sEvents...)
			testSubscriptionOnNATS(ens, subscription.Name, tc.want.natsSubscription...)
			testDeletion(ens, subscription)
		})
	}
	t.Cleanup(ens.cancel)
}

// TestChangeSubscription tests if existing subscriptions are reconciled properly after getting changed.
func TestChangeSubscription(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g, natsPort)

	var testCases = []struct {
		name                  string
		givenSubscriptionOpts []reconcilertesting.SubscriptionOpt
		wantBefore            want
		changeSubscription    func(subscription *eventingv1alpha1.Subscription)
		wantAfter             want
	}{
		{
			name: "CleanEventTypes; add filters to subscription without filters",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithEmptyFilter(),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			wantBefore: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				eventTypes := []string{
					newUncleanEventType("0"),
					newUncleanEventType("1"),
				}
				for _, eventType := range eventTypes {
					reconcilertesting.AddFilter(reconcilertesting.EventSource, eventType, subscription)
				}
			},
			wantAfter: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						newCleanEventType("0"),
						newCleanEventType("1"),
					}),
				},
			},
		},
		{
			name: "CleanEventTypes; change filters",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, newUncleanEventType("0")),
				reconcilertesting.WithFilter(reconcilertesting.EventSource, newUncleanEventType("1")),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			wantBefore: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						newCleanEventType("0"),
						newCleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				// change all the filters by adding "alpha" to the event type
				for _, f := range subscription.Spec.Filter.Filters {
					f.EventType.Value = fmt.Sprintf("%salpha", f.EventType.Value)
				}
			},
			wantAfter: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						newCleanEventType("0alpha"),
						newCleanEventType("1alpha"),
					}),
				},
			},
		},
		{
			name: "CleanEventTypes; delete a filter",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, newUncleanEventType("0")),
				reconcilertesting.WithFilter(reconcilertesting.EventSource, newUncleanEventType("1")),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			wantBefore: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						newCleanEventType("0"),
						newCleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				subscription.Spec.Filter.Filters = subscription.Spec.Filter.Filters[:1]
			},
			wantAfter: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						newCleanEventType("0"),
					}),
				},
			},
		},
		{
			name: "change configuration",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, newUncleanEventType("")),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			wantBefore: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				subscription.Spec.Config = &eventingv1alpha1.SubscriptionConfig{
					MaxInFlightMessages: 101,
				}
			},
			wantAfter: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(
						&eventingv1alpha1.SubscriptionConfig{
							MaxInFlightMessages: 101,
						},
					),
				},
				natsSubscription: []gomegatypes.GomegaMatcher{
					natstesting.BeExistingSubscription(),
					natstesting.BeValidSubscription(),
					natstesting.BeSubscriptionWithSubject(newCleanEventType("")),
				},
			},
		},
		{
			name: "resolve multiple conditions",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithEmptyFilter(),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithMultipleConditions(),
				reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
			},
			wantBefore: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCleanEventTypes(nil),
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveSubscriptionReady(),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				reconcilertesting.AddFilter(reconcilertesting.EventSource,
					reconcilertesting.OrderCreatedEventTypeNotClean,
					subscription,
				)
			},
			wantAfter: want{
				k8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveSubscriptionReady(),
					reconcilertesting.HaveCleanEventTypes([]string{reconcilertesting.OrderCreatedEventType}),
					gomega.Not(reconcilertesting.HaveCondition(reconcilertesting.MultipleDefaultConditions()[0])),
					gomega.Not(reconcilertesting.HaveCondition(reconcilertesting.MultipleDefaultConditions()[1])),
				},
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			subscription := createSubscription(ens, tc.givenSubscriptionOpts...)

			testSubscriptionOnK8s(ens, subscription, tc.wantBefore.k8sSubscription...)
			testEventsOnK8s(ens, tc.wantBefore.k8sEvents...)
			testSubscriptionOnNATS(ens, subscription.Name, tc.wantBefore.natsSubscription...)

			// when
			tc.changeSubscription(subscription)
			updateSubscriptionOnK8s(ens, subscription)

			// then
			testSubscriptionOnK8s(ens, subscription, tc.wantAfter.k8sSubscription...)
			testEventsOnK8s(ens, tc.wantAfter.k8sEvents...)
			testSubscriptionOnNATS(ens, subscription.Name, tc.wantAfter.natsSubscription...)
			testDeletion(ens, subscription)
		})
	}
	t.Cleanup(ens.cancel)
}

// TestEmptyEventTypePrefix tests if a subscription is reconciled properly if the NATS backend is unavailable.
func TestEmptyEventTypePrefix(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).NotTo(gomega.HaveOccurred())
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefixEmpty, g, natsPort)

	subscription := createSubscription(ens,
		reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotCleanPrefixEmpty),
		reconcilertesting.WithSinkURLFromSvc(ens.subscriberSvc),
	)

	testSubscriptionOnK8s(ens, subscription,
		reconcilertesting.HaveCleanEventTypes([]string{reconcilertesting.OrderCreatedEventTypePrefixEmpty}),
		reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
		reconcilertesting.HaveSubsConfiguration(configDefault(ens.defaultSubscriptionConfig.MaxInFlightMessages)),
		reconcilertesting.HaveSubscriptionReady(),
	)

	testSubscriptionOnNATS(ens, subscription.Name,
		natstesting.BeExistingSubscription(),
		natstesting.BeValidSubscription(),
		natstesting.BeSubscriptionWithSubject(reconcilertesting.OrderCreatedEventTypePrefixEmpty),
	)

	testDeletion(ens, subscription)

	t.Cleanup(ens.cancel)
}

func updateSubscriptionOnK8s(ens *testEnsemble, subscription *eventingv1alpha1.Subscription) {
	g := ens.g

	err := ens.k8sClient.Update(ens.ctx, subscription)
	g.Expect(err).Should(gomega.BeNil())
}

func createSubscription(ens *testEnsemble, subscriptionOpts ...reconcilertesting.SubscriptionOpt) *eventingv1alpha1.Subscription {
	subscriptionName := fmt.Sprintf(subscriptionNameFormat, ens.testID)
	ens.testID++
	subscription := reconcilertesting.NewSubscription(subscriptionName, ens.subscriberSvc.Namespace, subscriptionOpts...)
	subscription = createSubscriptionInK8s(ens, subscription)
	return subscription
}

func testSubscriptionOnK8s(ens *testEnsemble, subscription *eventingv1alpha1.Subscription, expectations ...gomegatypes.GomegaMatcher) {
	subExpectations := append(expectations, reconcilertesting.HaveSubscriptionName(subscription.Name))
	getSubscriptionOnK8S(ens, subscription).Should(gomega.And(subExpectations...))
}

func testEventsOnK8s(ens *testEnsemble, expectations ...v1.Event) {
	for _, event := range expectations {
		getK8sEvents(ens).Should(reconcilertesting.HaveEvent(event))
	}
}

func testSubscriptionOnNATS(ens *testEnsemble, subscriptionName string, expectations ...gomegatypes.GomegaMatcher) {
	getSubscriptionFromNATS(ens, subscriptionName).Should(gomega.And(expectations...))
}

func testDeletion(ens *testEnsemble, subscription *eventingv1alpha1.Subscription) {
	g := ens.g

	g.Expect(ens.k8sClient.Delete(ens.ctx, subscription)).Should(gomega.BeNil())
	isSubscriptionDeletedOnK8s(ens, subscription).Should(reconcilertesting.HaveNotFoundSubscription())
	getSubscriptionFromNATS(ens, subscription.Name).ShouldNot(natstesting.BeExistingSubscription())
}

func validSinkURL(ens *testEnsemble, additions ...string) string {
	url := reconcilertesting.ValidSinkURL(ens.subscriberSvc.Namespace, ens.subscriberSvc.Name)
	for _, a := range additions {
		url = fmt.Sprintf("%s%s", url, a)
	}
	return url
}

func newUncleanEventType(ending string) string {
	return fmt.Sprintf("%s%s", reconcilertesting.OrderCreatedEventTypeNotClean, ending)
}

func newCleanEventType(ending string) string {
	return fmt.Sprintf("%s%s", reconcilertesting.OrderCreatedEventType, ending)
}

// isSubscriptionDeletedOnK8s checks a subscription is deleted and allows making assertions on it
func isSubscriptionDeletedOnK8s(ens *testEnsemble, subscription *eventingv1alpha1.Subscription) gomega.AsyncAssertion {
	g := ens.g

	return g.Eventually(func() bool {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.k8sClient.Get(ens.ctx, lookupKey, subscription); err != nil {
			return k8serrors.IsNotFound(err)
		}
		return false
	}, smallTimeout, smallPollingInterval)
}

func configDefault(maxInFlightMsg int) *eventingv1alpha1.SubscriptionConfig {
	return &eventingv1alpha1.SubscriptionConfig{MaxInFlightMessages: maxInFlightMsg}
}

func conditionInvalidSink(msg string) eventingv1alpha1.Condition {
	return eventingv1alpha1.MakeCondition(
		eventingv1alpha1.ConditionSubscriptionActive,
		eventingv1alpha1.ConditionReasonNATSSubscriptionNotActive,
		v1.ConditionFalse, msg)
}

func eventInvalidSink(msg string) v1.Event {
	return v1.Event{
		Reason:  string(events.ReasonValidationFailed),
		Message: msg,
		Type:    v1.EventTypeWarning,
	}
}

func setupTestEnsemble(ctx context.Context, eventTypePrefix string, g *gomega.GomegaWithT, natsPort int) *testEnsemble {
	useExistingCluster := useExistingCluster
	ens := &testEnsemble{
		ctx: ctx,
		g:   g,
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

	startTestEnv(ens)
	startReconciler(eventTypePrefix, ens)
	startSubscriberSvc(ens)

	return ens
}

func startTestEnv(ens *testEnsemble) {
	g := ens.g

	k8sCfg, err := ens.testEnv.Start()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(k8sCfg).ToNot(gomega.BeNil())
	ens.cfg = k8sCfg
}

func startNATS(port int) *natsserver.Server {
	natsServer := reconcilertesting.RunNatsServerOnPort(
		reconcilertesting.WithPort(port),
	)
	log.Printf("NATS server started %v", natsServer.ClientURL())
	return natsServer
}

func startReconciler(eventTypePrefix string, ens *testEnsemble) *testEnsemble {
	g := ens.g

	ctx, cancel := context.WithCancel(context.Background())
	ens.cancel = cancel

	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	var metricsPort int
	metricsPort, err = reconcilertesting.GetFreePort()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(ens.cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: fmt.Sprintf("localhost:%v", metricsPort),
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

	natsHandler := handlers.NewNats(envConf, ens.defaultSubscriptionConfig, defaultLogger)
	cleaner := eventtype.NewCleaner(envConf.EventTypePrefix, applicationLister, defaultLogger)

	ens.reconciler = natsreconciler.NewReconciler(
		ctx,
		k8sManager.GetClient(),
		natsHandler,
		cleaner,
		defaultLogger,
		k8sManager.GetEventRecorderFor("eventing-controller-nats"),
		ens.defaultSubscriptionConfig,
	)

	err = ens.reconciler.SetupUnmanaged(k8sManager)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	ens.natsBackend = ens.reconciler.Backend.(*handlers.Nats)

	go func() {
		err = k8sManager.Start(ctx)
		g.Expect(err).ToNot(gomega.HaveOccurred())
	}()

	ens.k8sClient = k8sManager.GetClient()
	g.Expect(ens.k8sClient).ToNot(gomega.BeNil())

	return ens
}

func startSubscriberSvc(ens *testEnsemble) {
	ens.subscriberSvc = reconcilertesting.NewSubscriberSvc("test-subscriber", "test")
	createSubscriberSvcInK8s(ens)
}

func createSubscriberSvcInK8s(ens *testEnsemble) {
	g := ens.g

	// if the namespace is not "default" create it on the cluster
	if ens.subscriberSvc.Namespace != "default " {
		namespace := fixtureNamespace(ens.subscriberSvc.Namespace)
		if namespace.Name != "default" {
			err := ens.k8sClient.Create(ens.ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	// create subscriber svc on cluster
	err := ens.k8sClient.Create(ens.ctx, ens.subscriberSvc)
	g.Expect(err).Should(gomega.BeNil())
}

// createSubscriptionInK8s creates a Subscription on the K8s client of the testEnsemble. All the reconciliation
// happening will be reflected in the subscription.
func createSubscriptionInK8s(ens *testEnsemble, subscription *eventingv1alpha1.Subscription) *eventingv1alpha1.Subscription {
	g := ens.g

	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := ens.k8sClient.Create(ens.ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				g.Expect(err).ShouldNot(gomega.HaveOccurred())
			}
		}
	}

	// create subscription on cluster
	err := ens.k8sClient.Create(ens.ctx, subscription)
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
func getSubscriptionOnK8S(ens *testEnsemble, subscription *eventingv1alpha1.Subscription, intervals ...interface{}) gomega.AsyncAssertion {
	g := ens.g

	if len(intervals) == 0 {
		intervals = []interface{}{smallTimeout, smallPollingInterval}
	}
	return g.Eventually(func() *eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := ens.k8sClient.Get(ens.ctx, lookupKey, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}
		}
		return subscription
	}, intervals...)
}

// getK8sEvents returns all kubernetes events for the given namespace.
// The result can be used in a gomega assertion.
func getK8sEvents(ens *testEnsemble) gomega.AsyncAssertion {
	g := ens.g

	eventList := v1.EventList{}
	return g.Eventually(func() v1.EventList {
		err := ens.k8sClient.List(ens.ctx, &eventList, client.InNamespace(ens.subscriberSvc.Namespace))
		if err != nil {
			return v1.EventList{}
		}
		return eventList
	}, smallTimeout, smallPollingInterval)
}

func getSubscriptionFromNATS(ens *testEnsemble, subscriptionName string) gomega.Assertion {
	g := ens.g

	return g.Expect(func() *nats.Subscription {
		subscriptions := ens.natsBackend.GetAllSubscriptions()
		for key, subscription := range subscriptions {
			// the key does NOT ONLY contain the subscription name
			if strings.Contains(key, subscriptionName) {
				return subscription
			}
		}
		return nil
	}())
}
