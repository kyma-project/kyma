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
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	natsreconciler "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/nats"
	utils "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/metrics"
	nats2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/nats/core"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/sink"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/testing/nats"
)

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
	emptyEventSource         = ""
)

type natsTestEnsemble struct {
	reconciler  *natsreconciler.Reconciler
	natsBackend *nats2.Nats
	*utils.TestEnsemble
}

// TestUnavailableNATSServer tests if a subscription is reconciled properly when the NATS backend is unavailable.
func TestUnavailableNATSServer(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g, natsPort)
	defer utils.StopTestEnv(ens.TestEnsemble)

	subscription := utils.CreateSubscription(ens.TestEnsemble,
		reconcilertesting.WithFilter(emptyEventSource, utils.NewUncleanEventType("")),
		reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
	)
	utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription,
		reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
		reconcilertesting.HaveSubscriptionReady(),
		reconcilertesting.HaveCleanEventTypes([]string{utils.NewCleanEventType("")}),
	)

	ens.NatsServer.Shutdown()
	utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription,
		reconcilertesting.HaveSubscriptionNotReady(),
	)

	ens.NatsServer = startNATS(natsPort)
	utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription, reconcilertesting.HaveSubscriptionReady())

	t.Cleanup(ens.Cancel)
}

// TestCreateSubscription tests if subscriptions get created properly by the reconciler.
func TestCreateSubscription(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(gomega.HaveOccurred())

	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g, natsPort)
	defer utils.StopTestEnv(ens.TestEnsemble)

	var testCases = []struct {
		name                  string
		givenSubscriptionOpts []reconcilertesting.SubscriptionOpt
		want                  utils.Want
	}{
		{
			name: "create and delete",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					reconcilertesting.OrderCreatedEventType: {
						natstesting.BeExistingSubscription(),
						natstesting.BeValidSubscription(),
						natstesting.BeSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
					},
				},
			},
		},
		{
			name: "filter with empty event type",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, ""),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
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
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(utils.ConditionInvalidSink(sink.MissingSchemeErrMsg)),
				},
				K8sEvents: []v1.Event{utils.EventInvalidSink("Sink URL scheme should be HTTP or HTTPS: invalid")},
			},
		},
		{
			name: "invalid sink; invalid character",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL("http://127.0.0. 1"),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						utils.ConditionInvalidSink("not able to parse sink url with error: parse \"http://127.0.0. 1\": invalid character \" \" in host name")),
				},
				K8sEvents: []v1.Event{
					utils.EventInvalidSink("Not able to parse Sink URL with error: parse \"http://127.0.0. 1\": invalid character \" \" in host name")},
			},
		},
		{
			name: "invalid sink; missing suffix 'svc.cluster.local'",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL("http://127.0.0.1"),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						utils.ConditionInvalidSink("sink does not contain suffix: svc.cluster.local in the URL")),
				},
				K8sEvents: []v1.Event{
					utils.EventInvalidSink("Sink does not contain suffix: svc.cluster.local")},
			},
		},
		{
			name: "invalid sink; too many sub domains",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL(fmt.Sprintf("https://%s.%s.%s.svc.cluster.local", "testapp", "testsub", "test")),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						utils.ConditionInvalidSink("sink should contain 5 sub-domains: testapp.testsub.test.svc.cluster.local")),
				},
				K8sEvents: []v1.Event{
					utils.EventInvalidSink("Sink should contain 5 sub-domains: testapp.testsub.test.svc.cluster.local")},
			},
		},
		{
			name: "invalid sink; wrong namespace",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL(fmt.Sprintf("https://%s.%s.svc.cluster.local", "testapp", "wrong-ns")),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						utils.ConditionInvalidSink("namespace of subscription: test and the namespace of subscriber: wrong-ns are different")),
				},
				K8sEvents: []v1.Event{
					utils.EventInvalidSink("natsNamespace of subscription: test and the subscriber: wrong-ns are different")},
			},
		},
		{
			name: "invalid sink; not a valid cluster local service",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURL(
					reconcilertesting.ValidSinkURL(ens.SubscriberSvc.Namespace, "testapp")),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(
						utils.ConditionInvalidSink("sink is not a valid cluster local svc, failed with error: Service \"testapp\" not found")),
				},
				K8sEvents: []v1.Event{
					utils.EventInvalidSink("Sink does not correspond to a valid cluster local svc")},
			},
		},
		{
			name: "valid sink",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURL(utils.ValidSinkURL(ens.TestEnsemble)),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubscriptionReady(),
				},
			},
		},
		{
			name: "valid sink; with port",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURL(utils.ValidSinkURL(ens.TestEnsemble, ":8080")),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
				},
			},
		},
		{
			name: "valid sink; with port and endpoint",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURL(utils.ValidSinkURL(ens.TestEnsemble, ":8080", "/myEndpoint")),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubscriptionReady(),
				},
			},
		},
		{
			name: "empty protocol, protocol setting and dialect",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotClean),
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					reconcilertesting.OrderCreatedEventType: {
						natstesting.BeExistingSubscription(),
						natstesting.BeValidSubscription(),
						natstesting.BeSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			subscription := utils.CreateSubscription(ens.TestEnsemble, tc.givenSubscriptionOpts...)
			utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription, tc.want.K8sSubscription...)
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.want.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.want.NatsSubscriptions {
				testSubscriptionOnNATS(ens, eventType, matchers...)
			}
			testDeletion(ens, subscription)
		})
	}
	t.Cleanup(ens.Cancel)
}

// TestChangeSubscription tests if existing subscriptions are reconciled properly after getting changed.
func TestChangeSubscription(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g, natsPort)
	defer utils.StopTestEnv(ens.TestEnsemble)

	var testCases = []struct {
		name                  string
		givenSubscriptionOpts []reconcilertesting.SubscriptionOpt
		wantBefore            utils.Want
		changeSubscription    func(subscription *eventingv1alpha1.Subscription)
		wantAfter             utils.Want
	}{
		{
			name: "CleanEventTypes; add filters to subscription without filters",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithEmptyFilter(),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				eventTypes := []string{
					utils.NewUncleanEventType("0"),
					utils.NewUncleanEventType("1"),
				}
				for _, eventType := range eventTypes {
					reconcilertesting.AddFilter(reconcilertesting.EventSource, eventType, subscription)
				}
			},
			wantAfter: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						utils.NewCleanEventType("0"),
						utils.NewCleanEventType("1"),
					}),
				},
			},
		},
		{
			name: "CleanEventTypes; change filters",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewUncleanEventType("0")),
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewUncleanEventType("1")),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						utils.NewCleanEventType("0"),
						utils.NewCleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				// change all the filters by adding "alpha" to the event type
				for _, f := range subscription.Spec.Filter.Filters {
					f.EventType.Value = fmt.Sprintf("%salpha", f.EventType.Value)
				}
			},
			wantAfter: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						utils.NewCleanEventType("0alpha"),
						utils.NewCleanEventType("1alpha"),
					}),
				},
			},
		},
		{
			name: "CleanEventTypes; delete a filter",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewUncleanEventType("0")),
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewUncleanEventType("1")),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						utils.NewCleanEventType("0"),
						utils.NewCleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				subscription.Spec.Filter.Filters = subscription.Spec.Filter.Filters[:1]
			},
			wantAfter: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveCleanEventTypes([]string{
						utils.NewCleanEventType("0"),
					}),
				},
			},
		},
		{
			name: "change configuration",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewUncleanEventType("")),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				subscription.Spec.Config = &eventingv1alpha1.SubscriptionConfig{
					MaxInFlightMessages: 101,
				}
			},
			wantAfter: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(
						&eventingv1alpha1.SubscriptionConfig{
							MaxInFlightMessages: 101,
						},
					),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					utils.NewCleanEventType(""): {
						natstesting.BeExistingSubscription(),
						natstesting.BeValidSubscription(),
						natstesting.BeSubscriptionWithSubject(utils.NewCleanEventType("")),
					},
				},
			},
		},
		{
			name: "resolve multiple conditions",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithEmptyFilter(),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithMultipleConditions(),
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCleanEventTypesEmpty(),
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveSubscriptionReady(),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				reconcilertesting.AddFilter(reconcilertesting.EventSource,
					reconcilertesting.OrderCreatedEventTypeNotClean,
					subscription,
				)
			},
			wantAfter: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveSubscriptionReady(),
					reconcilertesting.HaveCleanEventTypes([]string{reconcilertesting.OrderCreatedEventType}),
					gomega.Not(reconcilertesting.HaveCondition(reconcilertesting.MultipleDefaultConditions()[0])),
					gomega.Not(reconcilertesting.HaveCondition(reconcilertesting.MultipleDefaultConditions()[1])),
				},
			},
		},
		{
			name: "CleanEventTypes; update valid filter to a filter with an invalid prefix",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewCleanEventType("0")),
				reconcilertesting.WithWebhookForNATS(),
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveCleanEventTypes([]string{
						utils.NewCleanEventType("0"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				// change the filter  to the event type
				for _, f := range subscription.Spec.Filter.Filters {
					f.EventType.Value = fmt.Sprintf("invalid%s", f.EventType.Value)
				}
			},
			wantAfter: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveConditionInvalidPrefix(),
					reconcilertesting.HaveCleanEventTypesEmpty(),
				},
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			// given
			subscription := utils.CreateSubscription(ens.TestEnsemble, tc.givenSubscriptionOpts...)

			utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription, tc.wantBefore.K8sSubscription...)
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.wantBefore.K8sEvents...)
			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.wantBefore.NatsSubscriptions {
				testSubscriptionOnNATS(ens, eventType, matchers...)
			}

			// when
			tc.changeSubscription(subscription)
			g.Expect(ens.K8sClient.Update(ens.Ctx, subscription)).ShouldNot(gomega.HaveOccurred())

			// then
			utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription, tc.wantAfter.K8sSubscription...)
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.wantAfter.K8sEvents...)
			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.wantAfter.NatsSubscriptions {
				testSubscriptionOnNATS(ens, eventType, matchers...)
			}
			testDeletion(ens, subscription)
		})
	}
	t.Cleanup(ens.Cancel)
}

// TestEmptyEventTypePrefix tests if a subscription is reconciled properly if the NATS backend is unavailable.
func TestEmptyEventTypePrefix(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).NotTo(gomega.HaveOccurred())
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefixEmpty, g, natsPort)
	defer utils.StopTestEnv(ens.TestEnsemble)

	subscription := utils.CreateSubscription(ens.TestEnsemble,
		reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotCleanPrefixEmpty),
		reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
	)

	utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription,
		reconcilertesting.HaveCleanEventTypes([]string{reconcilertesting.OrderCreatedEventTypePrefixEmpty}),
		reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
		reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
		reconcilertesting.HaveSubscriptionReady(),
	)

	testSubscriptionOnNATS(ens, subscription.Name,
		natstesting.BeExistingSubscription(),
		natstesting.BeValidSubscription(),
		natstesting.BeSubscriptionWithSubject(reconcilertesting.OrderCreatedEventTypePrefixEmpty),
	)

	testDeletion(ens, subscription)

	t.Cleanup(ens.Cancel)
}

func testSubscriptionOnNATS(ens *natsTestEnsemble, subscriptionName string, expectations ...gomegatypes.GomegaMatcher) {
	getSubscriptionFromNATS(ens, subscriptionName).Should(gomega.And(expectations...))
}

func testDeletion(ens *natsTestEnsemble, subscription *eventingv1alpha1.Subscription) {
	g := ens.G

	g.Expect(ens.K8sClient.Delete(ens.Ctx, subscription)).Should(gomega.BeNil())
	utils.IsSubscriptionDeletedOnK8s(ens.TestEnsemble, subscription).Should(reconcilertesting.HaveNotFoundSubscription())
	getSubscriptionFromNATS(ens, subscription.Name).ShouldNot(natstesting.BeExistingSubscription())
}

func setupTestEnsemble(ctx context.Context, eventTypePrefix string, g *gomega.GomegaWithT, natsPort int) *natsTestEnsemble {
	useExistingCluster := useExistingCluster
	ens := &utils.TestEnsemble{
		Ctx: ctx,
		G:   g,
		DefaultSubscriptionConfig: env.DefaultSubscriptionConfig{
			MaxInFlightMessages:   1,
			DispatcherRetryPeriod: time.Second,
			DispatcherMaxRetries:  1,
		},
		NatsServer: startNATS(natsPort),
		TestEnv: &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join("../../../", "config", "crd", "bases"),
				filepath.Join("../../../", "config", "crd", "external"),
			},
			AttachControlPlaneOutput: attachControlPlaneOutput,
			UseExistingCluster:       &useExistingCluster,
		},
	}

	natsTestEns := natsTestEnsemble{
		TestEnsemble: ens,
	}

	utils.StartTestEnv(ens)
	startReconciler(eventTypePrefix, &natsTestEns)
	utils.StartSubscriberSvc(ens)

	return &natsTestEns
}

func startNATS(port int) *natsserver.Server {
	natsServer := reconcilertesting.RunNatsServerOnPort(
		reconcilertesting.WithPort(port),
	)
	log.Printf("NATS server started %v", natsServer.ClientURL())
	return natsServer
}

func startReconciler(eventTypePrefix string, ens *natsTestEnsemble) *natsTestEnsemble {
	g := ens.G

	ctx, cancel := context.WithCancel(context.Background())
	ens.Cancel = cancel

	err := eventingv1alpha1.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	var metricsPort int
	metricsPort, err = reconcilertesting.GetFreePort()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(ens.Cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: fmt.Sprintf("localhost:%v", metricsPort),
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	envConf := env.NatsConfig{
		URL:             ens.NatsServer.ClientURL(),
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: eventTypePrefix,
	}

	// prepare application-lister
	app := applicationtest.NewApplication(reconcilertesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), app)

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(gomega.BeNil())

	natsHandler := nats2.NewNats(envConf, ens.DefaultSubscriptionConfig, metricsCollector, defaultLogger)
	cleaner := eventtype.NewCleaner(envConf.EventTypePrefix, applicationLister, defaultLogger)

	k8sClient := k8sManager.GetClient()
	recorder := k8sManager.GetEventRecorderFor("eventing-controller-nats")

	ens.reconciler = natsreconciler.NewReconciler(
		ctx,
		k8sClient,
		natsHandler,
		cleaner,
		defaultLogger,
		recorder,
		ens.DefaultSubscriptionConfig,
		sink.NewValidator(ctx, k8sClient, recorder, defaultLogger),
	)

	err = ens.reconciler.SetupUnmanaged(k8sManager)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	ens.natsBackend = ens.reconciler.Backend.(*nats2.Nats)

	go func() {
		err = k8sManager.Start(ctx)
		g.Expect(err).ToNot(gomega.HaveOccurred())
	}()

	ens.K8sClient = k8sManager.GetClient()
	g.Expect(ens.K8sClient).ToNot(gomega.BeNil())

	return ens
}

func getSubscriptionFromNATS(ens *natsTestEnsemble, subscriptionName string) gomega.Assertion {
	g := ens.G

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
