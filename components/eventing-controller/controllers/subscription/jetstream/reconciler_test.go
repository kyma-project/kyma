package jetstream

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/metrics"

	"github.com/nats-io/nats.go"

	utils "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/eventtype"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/sink"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	natstesting "github.com/kyma-project/kyma/components/eventing-controller/testing/nats"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
	emptyEventSource         = ""
)

type jetStreamTestEnsemble struct {
	reconciler       *Reconciler
	jetStreamBackend *handlers.JetStream
	*utils.TestEnsemble
}

// TestUnavailableNATSServer tests if a subscription is reconciled properly when the NATS backend is unavailable.
func TestUnavailableNATSServer(t *testing.T) {
	ctx := context.Background()
	g := gomega.NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(gomega.HaveOccurred())
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefix, g, natsPort)

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

	ens.NatsServer = startJetStream(natsPort)
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
				NatsSubscription: []gomegatypes.GomegaMatcher{
					natstesting.BeExistingSubscription(),
					natstesting.BeValidSubscription(),
					natstesting.BeJetStreamSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
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
				NatsSubscription: []gomegatypes.GomegaMatcher{
					natstesting.BeExistingSubscription(),
					natstesting.BeValidSubscription(),
					natstesting.BeJetStreamSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subscription := utils.CreateSubscription(ens.TestEnsemble, tc.givenSubscriptionOpts...)
			utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription, tc.want.K8sSubscription...)
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.want.K8sEvents...)
			testSubscriptionOnNATS(ens, subscription, reconcilertesting.OrderCreatedEventType, tc.want.NatsSubscription...)
			testDeletion(ens, subscription, reconcilertesting.OrderCreatedEventType)
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
				NatsSubscription: []gomegatypes.GomegaMatcher{
					natstesting.BeExistingSubscription(),
					natstesting.BeValidSubscription(),
					natstesting.BeSubscriptionWithSubject(utils.NewCleanEventType("")),
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
			// given
			subscription := utils.CreateSubscription(ens.TestEnsemble, tc.givenSubscriptionOpts...)

			utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription, tc.wantBefore.K8sSubscription...)
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.wantBefore.K8sEvents...)
			testSubscriptionOnNATS(ens, subscription,
				reconcilertesting.OrderCreatedEventType, tc.wantBefore.NatsSubscription...)

			// when
			tc.changeSubscription(subscription)
			utils.UpdateSubscriptionOnK8s(ens.TestEnsemble, subscription)

			// then
			utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription, tc.wantAfter.K8sSubscription...)
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.wantAfter.K8sEvents...)
			testSubscriptionOnNATS(ens, subscription,
				reconcilertesting.OrderCreatedEventType, tc.wantBefore.NatsSubscription...)
			testDeletion(ens, subscription, reconcilertesting.OrderCreatedEventType)
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

	// when
	subscription := utils.CreateSubscription(ens.TestEnsemble,
		reconcilertesting.WithFilter(emptyEventSource, reconcilertesting.OrderCreatedEventTypeNotCleanPrefixEmpty),
		reconcilertesting.WithSinkURLFromSvc(ens.SubscriberSvc),
	)

	// then
	utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription,
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

	testSubscriptionOnNATS(ens, subscription, reconcilertesting.OrderCreatedEventTypePrefixEmpty, expectedNatsSubscription...)

	testDeletion(ens, subscription, reconcilertesting.OrderCreatedEventTypePrefixEmpty)

	t.Cleanup(ens.Cancel)
}

func testSubscriptionOnNATS(ens *jetStreamTestEnsemble, subscription *eventingv1alpha1.Subscription, subject string, expectations ...gomegatypes.GomegaMatcher) {
	getSubscriptionFromJetStream(ens, subscription, ens.jetStreamBackend.GetJetstreamSubject(subject)).Should(gomega.And(expectations...))
}

func testDeletion(ens *jetStreamTestEnsemble, subscription *eventingv1alpha1.Subscription, subject string) {
	g := ens.G
	g.Expect(ens.K8sClient.Delete(ens.Ctx, subscription)).Should(gomega.BeNil())
	utils.IsSubscriptionDeletedOnK8s(ens.TestEnsemble, subscription).Should(reconcilertesting.HaveNotFoundSubscription())
	getSubscriptionFromJetStream(ens, subscription, subject).ShouldNot(natstesting.BeExistingSubscription())
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

	jetStreamHandler := handlers.NewJetStream(envConf, metricsCollector, defaultLogger)
	cleaner := eventtype.NewCleaner(envConf.EventTypePrefix, applicationLister, defaultLogger)

	k8sClient := k8sManager.GetClient()
	recorder := k8sManager.GetEventRecorderFor("eventing-controller-nats")

	ens.reconciler = NewReconciler(
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

	ens.jetStreamBackend = ens.reconciler.Backend.(*handlers.JetStream)

	go func() {
		err = k8sManager.Start(ctx)
		g.Expect(err).ToNot(gomega.HaveOccurred())
	}()

	ens.K8sClient = k8sManager.GetClient()
	g.Expect(ens.K8sClient).ToNot(gomega.BeNil())

	return ens
}

func getSubscriptionFromJetStream(ens *jetStreamTestEnsemble, subscription *eventingv1alpha1.Subscription, subject string) gomega.Assertion {
	g := ens.G

	return g.Expect(func() *nats.Subscription {
		subscriptions := ens.jetStreamBackend.GetAllSubscriptions()
		subscriptionSubject := handlers.NewSubscriptionSubjectIdentifier(subscription, subject)
		for key, sub := range subscriptions {
			if key.ConsumerName() == subscriptionSubject.ConsumerName() {
				return sub
			}
		}
		return nil
	}())
}
