// todo integration

package jetstream_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	subscriptionjetstream "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/jetstream"
	utils "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	backendjetstream "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/jetstream"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/testing/nats"
)

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
	emptyEventSource         = ""
)

type jetStreamTestEnsemble struct {
	reconciler       *subscriptionjetstream.Reconciler
	jetStreamBackend *backendjetstream.JetStream
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

	sub := utils.CreateSubscription(ens.TestEnsemble,
		reconcilertesting.WithFilter(emptyEventSource, utils.NewUncleanEventType("")),
		reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
	)

	utils.TestSubscriptionOnK8s(ens.TestEnsemble, sub,
		reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
		reconcilertesting.HaveSubscriptionReady(),
		reconcilertesting.HaveCleanEventTypes([]string{utils.NewCleanEventType("")}),
	)

	ens.NatsServer.Shutdown()
	utils.TestSubscriptionOnK8s(ens.TestEnsemble, sub,
		reconcilertesting.HaveSubscriptionNotReady(),
	)

	ens.NatsServer = startJetStream(natsPort)
	utils.TestSubscriptionOnK8s(ens.TestEnsemble, sub, reconcilertesting.HaveSubscriptionReady())

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
				reconcilertesting.WithFinalizers([]string{}),
			},
			want: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
					reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
					reconcilertesting.HaveSubscriptionFinalizer(eventingv1alpha1.Finalizer),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					reconcilertesting.OrderCreatedEventType: {
						natstesting.BeExistingSubscription(),
						natstesting.BeValidSubscription(),
						natstesting.BeJetStreamSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
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
						utils.ConditionInvalidSink("failed to parse subscription sink URL: parse \"http://127.0.0. 1\": invalid character \" \" in host name")),
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
						utils.ConditionInvalidSink("failed to validate subscription sink URL. It does not contain suffix: svc.cluster.local")),
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
						utils.ConditionInvalidSink("failed to validate subscription sink URL. It should contain 5 sub-domains: testapp.testsub.test.svc.cluster.local")),
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
						utils.ConditionInvalidSink("failed to validate subscription sink URL. It is not a valid cluster local svc: Service \"testapp\" not found")),
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
						natstesting.BeJetStreamSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// The new testing.T instance needs to be used for making assertions. If ignored, errors like this will be printed:
			// testing.go:1336: test executed panic(nil) or runtime.Goexit: subtest may have called FailNow on a parent test
			ens := ens
			ens.G = gomega.NewGomegaWithT(t)

			t.Log("creating the k8s subscription")
			sub := utils.CreateSubscription(ens.TestEnsemble, tc.givenSubscriptionOpts...)
			t.Log("testing the k8s subscription")
			utils.TestSubscriptionOnK8s(ens.TestEnsemble, sub, tc.want.K8sSubscription...)

			t.Log("testing the k8s events")
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.want.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.want.NatsSubscriptions {
				testSubscriptionOnNATS(ens, sub, eventType, matchers...)
			}

			t.Log("testing the deletion of the subscription")
			testSubscriptionDeletion(ens, sub)

			t.Log("testing the deletion of the NATS subscription(s)")
			for _, filter := range sub.Spec.Filter.Filters {
				ensureNATSSubscriptionIsDeleted(ens, sub, filter.EventType.Value)
			}
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
			// Ensure subscriptions on NATS are not added when adding a Subscription filter and providing an invalid sink to prevent event-loss.
			// The reason for this is that the dispatcher will retry only for a finite number and then give up.
			// Since the sink is invalid, the dispatcher cannot dispatch the event and will stop if the maximum number of retries is reached.
			name: "Disallow the creation of a NATS subscription with an invalid sink",
			givenSubscriptionOpts: []reconcilertesting.SubscriptionOpt{
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewCleanEventType("0")),
				// valid sink
				reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveSubscriptionReady(),
					// for each filter we want to have a clean event type
					reconcilertesting.HaveCleanEventTypes([]string{
						utils.NewCleanEventType("0"),
					}),
				},
				K8sEvents: nil,
				// ensure that each filter results in a NATS consumer
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					utils.NewCleanEventType("0"): {
						natstesting.BeExistingSubscription(),
						natstesting.BeValidSubscription(),
						natstesting.BeJetStreamSubscriptionWithSubject(utils.NewCleanEventType("0")),
					},
				},
			},
			changeSubscription: func(subscription *eventingv1alpha1.Subscription) {
				reconcilertesting.AddFilter(reconcilertesting.EventSource, utils.NewCleanEventType("1"), subscription)

				// induce an error by making the sink invalid
				subscription.Spec.Sink = "invalid"
			},
			wantAfter: utils.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					reconcilertesting.HaveSubscriptionNotReady(),
					// for each filter we want to have a clean event type
					reconcilertesting.HaveCleanEventTypes([]string{
						utils.NewCleanEventType("0"),
						utils.NewCleanEventType("1"),
					}),
				},
				K8sEvents: nil,
				// ensure that each filter results in a NATS consumer
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					utils.NewCleanEventType("0"): {
						natstesting.BeExistingSubscription(),
					},
					// the newly added filter is not synced to NATS as the sink is invalid
					utils.NewCleanEventType("1"): {
						gomega.Not(natstesting.BeExistingSubscription()),
					},
				},
			},
		},
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
					utils.NewCleanEventType("0"),
					utils.NewCleanEventType("1"),
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
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewCleanEventType("0")),
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewCleanEventType("1")),
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
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewCleanEventType("0")),
				reconcilertesting.WithFilter(reconcilertesting.EventSource, utils.NewCleanEventType("1")),
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
						natstesting.BeJetStreamSubscriptionWithSubject(utils.NewCleanEventType("")),
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// The new t instance needs to be used for making assertions. If ignored, errors like this will be printed:
			// testing.go:1336: test executed panic(nil) or runtime.Goexit: subtest may have called FailNow on a parent test
			ens := ens
			g = gomega.NewGomegaWithT(t)
			ens.G = g

			// given
			t.Log("creating the k8s subscription")
			sub := utils.CreateSubscription(ens.TestEnsemble, tc.givenSubscriptionOpts...)

			t.Log("testing the k8s subscription")
			utils.TestSubscriptionOnK8s(ens.TestEnsemble, sub, tc.wantBefore.K8sSubscription...)

			t.Log("testing the k8s events")
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.wantBefore.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.wantBefore.NatsSubscriptions {
				testSubscriptionOnNATS(ens, sub, eventType, matchers...)
			}

			// when
			t.Log("change and update the subscription")
			utils.EventuallyUpdateSubscriptionOnK8s(ctx, ens.TestEnsemble, sub, func(sub *eventingv1alpha1.Subscription) error {
				tc.changeSubscription(sub)
				return ens.K8sClient.Update(ens.Ctx, sub)
			})

			// then
			t.Log("testing the k8s subscription")
			utils.TestSubscriptionOnK8s(ens.TestEnsemble, sub, tc.wantAfter.K8sSubscription...)

			t.Log("testing the k8s events")
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.wantAfter.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.wantAfter.NatsSubscriptions {
				testSubscriptionOnNATS(ens, sub, eventType, matchers...)
			}

			t.Log("testing the deletion of the subscription")
			testSubscriptionDeletion(ens, sub)

			t.Log("testing the deletion of the NATS subscription(s)")
			for _, filter := range sub.Spec.Filter.Filters {
				ensureNATSSubscriptionIsDeleted(ens, sub, filter.EventType.Value)
			}
		})
	}
	t.Cleanup(ens.Cancel)
}

// TestEmptyEventTypePrefix tests if a subscription is reconciled properly if the EventTypePrefix is empty.
func TestEmptyEventTypePrefix(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).NotTo(gomega.HaveOccurred())
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefixEmpty, g, natsPort)
	defer utils.StopTestEnv(ens.TestEnsemble)

	// when
	sub := utils.CreateSubscription(ens.TestEnsemble,
		reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotCleanPrefixEmpty),
		reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
	)

	// then
	utils.TestSubscriptionOnK8s(ens.TestEnsemble, sub,
		reconcilertesting.HaveCleanEventTypes([]string{reconcilertesting.OrderCreatedEventTypePrefixEmpty}),
		reconcilertesting.HaveCondition(reconcilertesting.DefaultReadyCondition()),
		reconcilertesting.HaveSubsConfiguration(utils.ConfigDefault(ens.DefaultSubscriptionConfig.MaxInFlightMessages)),
		reconcilertesting.HaveSubscriptionReady(),
	)

	expectedNatsSubscription := []gomegatypes.GomegaMatcher{
		natstesting.BeExistingSubscription(),
		natstesting.BeValidSubscription(),
		natstesting.BeJetStreamSubscriptionWithSubject(reconcilertesting.OrderCreatedEventTypePrefixEmpty),
	}

	testSubscriptionOnNATS(ens, sub, reconcilertesting.OrderCreatedEventTypePrefixEmpty, expectedNatsSubscription...)

	testSubscriptionDeletion(ens, sub)
	ensureNATSSubscriptionIsDeleted(ens, sub, reconcilertesting.OrderCreatedEventTypePrefixEmpty)

	t.Cleanup(ens.Cancel)
}

func testSubscriptionOnNATS(ens *jetStreamTestEnsemble, subscription *eventingv1alpha1.Subscription, subject string, expectations ...gomegatypes.GomegaMatcher) {
	description := "Failed to match nats subscriptions"
	getSubscriptionFromJetStream(ens, subscription, ens.jetStreamBackend.GetJetStreamSubject(subject)).Should(gomega.And(expectations...), description)
}

// testSubscriptionDeletion deletes the subscription and ensures it is not found anymore on the apiserver.
func testSubscriptionDeletion(ens *jetStreamTestEnsemble, subscription *eventingv1alpha1.Subscription) {
	g := ens.G
	g.Eventually(func() error {
		return ens.K8sClient.Delete(ens.Ctx, subscription)
	}, utils.SmallTimeout, utils.SmallPollingInterval).ShouldNot(gomega.HaveOccurred())
	utils.IsSubscriptionDeletedOnK8s(ens.TestEnsemble, subscription).Should(reconcilertesting.HaveNotFoundSubscription(), "Failed to delete subscription")
}

// ensureNATSSubscriptionIsDeleted ensures that the NATS subscription is not found anymore.
// This ensures the controller did delete it correctly then the Subscription was deleted.
func ensureNATSSubscriptionIsDeleted(ens *jetStreamTestEnsemble, subscription *eventingv1alpha1.Subscription, subject string) {
	getSubscriptionFromJetStream(ens, subscription, subject).ShouldNot(natstesting.BeExistingSubscription(), "Failed to delete NATS subscription")
}

func setupTestEnsemble(ctx context.Context, eventTypePrefix string, g *gomega.GomegaWithT, natsPort int) *jetStreamTestEnsemble {
	useExistingCluster := useExistingCluster
	ens := &utils.TestEnsemble{
		Ctx: ctx,
		G:   g,
		DefaultSubscriptionConfig: env.DefaultSubscriptionConfig{
			MaxInFlightMessages: 1,
		},
		NatsServer: startJetStream(natsPort),
		TestEnv: &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join("../../../", "config", "crd", "bases"),
				filepath.Join("../../../", "config", "crd", "external"),
			},
			AttachControlPlaneOutput: attachControlPlaneOutput,
			UseExistingCluster:       &useExistingCluster,
		},
	}

	jsTestEnsemble := &jetStreamTestEnsemble{
		TestEnsemble: ens,
	}

	utils.StartTestEnv(ens)
	startReconciler(eventTypePrefix, jsTestEnsemble)
	utils.StartSubscriberSvc(ens)

	return jsTestEnsemble
}

func startJetStream(port int) *natsserver.Server {
	natsServer := reconcilertesting.RunNatsServerOnPort(
		reconcilertesting.WithPort(port),
		reconcilertesting.WithJetStreamEnabled(),
	)
	log.Printf("NATS server with JetStream started %v", natsServer.ClientURL())
	return natsServer
}

func startReconciler(eventTypePrefix string, ens *jetStreamTestEnsemble) *jetStreamTestEnsemble {
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
		URL:                     ens.NatsServer.ClientURL(),
		MaxReconnects:           10,
		ReconnectWait:           time.Second,
		EventTypePrefix:         eventTypePrefix,
		JSStreamName:            reconcilertesting.JSStreamName,
		JSStreamStorageType:     "memory",
		JSStreamMaxBytes:        -1,
		JSStreamMaxMessages:     -1,
		JSStreamRetentionPolicy: "interest",
	}

	// prepare application-lister
	app := applicationtest.NewApplication(reconcilertesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), app)

	// init the metrics collector
	metricsCollector := metrics.NewCollector()

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(gomega.BeNil())

	jetStreamHandler := backendjetstream.NewJetStream(envConf, metricsCollector, defaultLogger)
	cleaner := eventtype.NewCleaner(envConf.EventTypePrefix, applicationLister, defaultLogger)

	k8sClient := k8sManager.GetClient()
	recorder := k8sManager.GetEventRecorderFor("eventing-controller-nats")

	ens.reconciler = subscriptionjetstream.NewReconciler(
		ctx,
		k8sClient,
		jetStreamHandler,
		defaultLogger,
		recorder,
		cleaner,
		ens.DefaultSubscriptionConfig,
		sink.NewValidator(ctx, k8sClient, recorder, defaultLogger),
	)

	err = ens.reconciler.SetupUnmanaged(k8sManager)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	ens.jetStreamBackend = ens.reconciler.Backend.(*backendjetstream.JetStream)

	go func() {
		err = k8sManager.Start(ctx)
		g.Expect(err).ToNot(gomega.HaveOccurred())
	}()

	ens.K8sClient = k8sManager.GetClient()
	g.Expect(ens.K8sClient).ToNot(gomega.BeNil())

	return ens
}

// getSubscriptionFromJetStream returns a NATS subscription for a given subscription and subject.
// NOTE: We need to give the controller enough time to react on the changes. Otherwise, the returned NATS subscription could have the wrong state.
// For this reason Eventually is used here.
func getSubscriptionFromJetStream(ens *jetStreamTestEnsemble, subscription *eventingv1alpha1.Subscription, subject string) gomega.AsyncAssertion {
	g := ens.G

	return g.Eventually(func() nats.Subscriber {
		subscriptions := ens.jetStreamBackend.GetAllSubscriptions()
		subscriptionSubject := backendjetstream.NewSubscriptionSubjectIdentifier(subscription, subject)
		for key, sub := range subscriptions {
			if key.ConsumerName() == subscriptionSubject.ConsumerName() {
				return sub
			}
		}
		return nil
	}, utils.SmallTimeout, utils.SmallPollingInterval)
}
