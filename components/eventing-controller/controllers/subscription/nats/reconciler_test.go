package nats_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	utils "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscription/testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers/sink"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
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
	. "github.com/onsi/gomega"
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

type natsTestEnsemble struct {
	reconciler  *natsreconciler.Reconciler
	natsBackend *handlers.Nats
	*utils.TestEnsemble
}

// TestUnavailableNATSServer tests if a subscription is reconciled properly when the NATS backend is unavailable.
func TestUnavailableNATSServer(t *testing.T) {
	ctx := context.Background()
	g := NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(HaveOccurred())
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

	ens.NatsServer = startNATS(natsPort)
	utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription, reconcilertesting.HaveSubscriptionReady())

	t.Cleanup(ens.Cancel)
}

// TestCreateSubscription tests if subscriptions get created properly by the reconciler.
func TestCreateSubscription(t *testing.T) {
	ctx := context.Background()
	g := NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(HaveOccurred())

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
					natstesting.BeSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
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
						utils.ConditionInvalidSink("sink is not valid cluster local svc, failed with error: Service \"testapp\" not found")),
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
					natstesting.BeSubscriptionWithSubject(reconcilertesting.OrderCreatedEventType),
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
			testSubscriptionOnNATS(ens, subscription.Name, tc.want.NatsSubscription...)
			testDeletion(ens, subscription)
		})
	}
	t.Cleanup(ens.Cancel)
}

// TestChangeSubscription tests if existing subscriptions are reconciled properly after getting changed.
func TestChangeSubscription(t *testing.T) {
	ctx := context.Background()
	g := NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).ToNot(HaveOccurred())
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
					reconcilertesting.HaveCleanEventTypes(nil),
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
					Not(reconcilertesting.HaveCondition(reconcilertesting.MultipleDefaultConditions()[0])),
					Not(reconcilertesting.HaveCondition(reconcilertesting.MultipleDefaultConditions()[1])),
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
			testSubscriptionOnNATS(ens, subscription.Name, tc.wantBefore.NatsSubscription...)

			// when
			tc.changeSubscription(subscription)
			utils.UpdateSubscriptionOnK8s(ens.TestEnsemble, subscription)

			// then
			utils.TestSubscriptionOnK8s(ens.TestEnsemble, subscription, tc.wantAfter.K8sSubscription...)
			utils.TestEventsOnK8s(ens.TestEnsemble, tc.wantAfter.K8sEvents...)
			testSubscriptionOnNATS(ens, subscription.Name, tc.wantAfter.NatsSubscription...)
			testDeletion(ens, subscription)
		})
	}
	t.Cleanup(ens.Cancel)
}

// TestEmptyEventTypePrefix tests if a subscription is reconciled properly if the NATS backend is unavailable.
func TestEmptyEventTypePrefix(t *testing.T) {
	ctx := context.Background()
	g := NewGomegaWithT(t)

	natsPort, err := reconcilertesting.GetFreePort()
	g.Expect(err).NotTo(HaveOccurred())
	ens := setupTestEnsemble(ctx, reconcilertesting.EventTypePrefixEmpty, g, natsPort)

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
	getSubscriptionFromNATS(ens, subscriptionName).Should(And(expectations...))
}

func testDeletion(ens *natsTestEnsemble, subscription *eventingv1alpha1.Subscription) {
	g := ens.G

	g.Expect(ens.K8sClient.Delete(ens.Ctx, subscription)).Should(BeNil())
	utils.IsSubscriptionDeletedOnK8s(ens.TestEnsemble, subscription).Should(reconcilertesting.HaveNotFoundSubscription())
	getSubscriptionFromNATS(ens, subscription.Name).ShouldNot(natstesting.BeExistingSubscription())
}

func setupTestEnsemble(ctx context.Context, eventTypePrefix string, g *GomegaWithT, natsPort int) *natsTestEnsemble {
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
	g.Expect(err).NotTo(HaveOccurred())

	var metricsPort int
	metricsPort, err = reconcilertesting.GetFreePort()
	g.Expect(err).NotTo(HaveOccurred())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(ens.Cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: fmt.Sprintf("localhost:%v", metricsPort),
	})
	g.Expect(err).ToNot(HaveOccurred())

	envConf := env.NatsConfig{
		URL:             ens.NatsServer.ClientURL(),
		MaxReconnects:   10,
		ReconnectWait:   time.Second,
		EventTypePrefix: eventTypePrefix,
	}

	// prepare application-lister
	app := applicationtest.NewApplication(reconcilertesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), app)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(BeNil())

	natsHandler := handlers.NewNats(envConf, ens.DefaultSubscriptionConfig, defaultLogger)
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
	g.Expect(err).ToNot(HaveOccurred())

	ens.natsBackend = ens.reconciler.Backend.(*handlers.Nats)

	go func() {
		err = k8sManager.Start(ctx)
		g.Expect(err).ToNot(HaveOccurred())
	}()

	ens.K8sClient = k8sManager.GetClient()
	g.Expect(ens.K8sClient).ToNot(BeNil())

	return ens
}

func getSubscriptionFromNATS(ens *natsTestEnsemble, subscriptionName string) Assertion {
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

// Test the return values of the Reconcile() method of the reconciler. This is important, as it dictates whether the
// reconciliation should be requeued by Controller Runtime, and if so with how much initial delay.
// Returning error or a `Result{Requeue: true}` would cause the reconciliation to be requeued.
// Everything else is mocked since we are only interested in the logic of the Reconcile method and not the reconciler dependencies.
func TestReconciler_Reconcile(t *testing.T) {
	ctx := context.Background()
	g := NewGomegaWithT(t)

	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	g.Expect(err).To(BeNil())
	recorder := &record.FakeRecorder{}
	defaultSubConfig := env.DefaultSubscriptionConfig{}
	// A subscription with the correct Finalizer, ready for reconciliation with the backend.
	testSub := reconcilertesting.NewSubscription("sub1", "test",
		reconcilertesting.WithFinalizers([]string{natsreconciler.Finalizer}),
		reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventType),
	)
	// A subscription marked for deletion.
	testSubUnderDeletion := reconcilertesting.NewSubscription("sub2", "test",
		reconcilertesting.WithNonZeroDeletionTimestamp(),
		reconcilertesting.WithFinalizers([]string{natsreconciler.Finalizer}),
		reconcilertesting.WithFilter(reconcilertesting.EventSource, reconcilertesting.OrderCreatedEventType),
	)

	backendSyncErr := errors.New("backend sync error")
	backendDeleteErr := errors.New("backend delete error")
	validatorErr := errors.New("invalid sink")
	cleanerErr := errors.New("invalid event type format")
	happyCleaner := eventtype.CleanerFunc(func(et string) (string, error) { return et, nil })
	unhappyCleaner := eventtype.CleanerFunc(func(et string) (string, error) { return et, cleanerErr })
	happyValidator := sink.ValidatorFunc(func(s *eventingv1alpha1.Subscription) error { return nil })
	unhappyValidator := sink.ValidatorFunc(func(s *eventingv1alpha1.Subscription) error { return validatorErr })

	var testCases = []struct {
		name                       string
		givenSubscription          *eventingv1alpha1.Subscription
		givenReconcilerSetup       func() *natsreconciler.Reconciler
		wantReconcileResultMatcher gomegatypes.GomegaMatcher
		wantReconcileErrorMatcher  gomegatypes.GomegaMatcher
	}{
		{
			name:              "Return nil and default Result{} when there is no error from the reconciler dependencies",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *natsreconciler.Reconciler {
				fakeClient := clientBuilderWithScheme(g).WithObjects(testSub).Build()
				natsBackend := &mocks.NatsBackend{}
				natsBackend.On("Initialize", mock.Anything).Return(nil)
				natsBackend.On("SyncSubscription", mock.Anything).Return(nil)
				return natsreconciler.NewReconciler(ctx, fakeClient, natsBackend, happyCleaner, defaultLogger, recorder, defaultSubConfig, happyValidator)
			},
			wantReconcileResultMatcher: Equal(ctrl.Result{}),
			wantReconcileErrorMatcher:  BeNil(),
		},
		{
			name:              "Return error and default Result{} when backend sync returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *natsreconciler.Reconciler {
				fakeClient := clientBuilderWithScheme(g).WithObjects(testSub).Build()
				natsBackend := &mocks.NatsBackend{}
				natsBackend.On("Initialize", mock.Anything).Return(nil)
				natsBackend.On("SyncSubscription", mock.Anything).Return(backendSyncErr)
				return natsreconciler.NewReconciler(ctx, fakeClient, natsBackend, happyCleaner, defaultLogger, recorder, defaultSubConfig, happyValidator)
			},
			wantReconcileResultMatcher: Equal(ctrl.Result{}),
			wantReconcileErrorMatcher:  Equal(backendSyncErr),
		},
		{
			name:              "Return error and default Result{} when backend delete returns error",
			givenSubscription: testSubUnderDeletion,
			givenReconcilerSetup: func() *natsreconciler.Reconciler {
				fakeClient := clientBuilderWithScheme(g).WithObjects(testSubUnderDeletion).Build()
				natsBackend := &mocks.NatsBackend{}
				natsBackend.On("Initialize", mock.Anything).Return(nil)
				natsBackend.On("DeleteSubscription", mock.Anything).Return(backendDeleteErr)
				return natsreconciler.NewReconciler(ctx, fakeClient, natsBackend, happyCleaner, defaultLogger, recorder, defaultSubConfig, happyValidator)
			},
			wantReconcileResultMatcher: Equal(ctrl.Result{}),
			wantReconcileErrorMatcher:  Equal(backendDeleteErr),
		},
		{
			name:              "Return error and default Result{} when validator returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *natsreconciler.Reconciler {
				fakeClient := clientBuilderWithScheme(g).WithObjects(testSub).Build()
				natsBackend := &mocks.NatsBackend{}
				natsBackend.On("Initialize", mock.Anything).Return(nil)
				natsBackend.On("SyncSubscription", mock.Anything).Return(nil)
				return natsreconciler.NewReconciler(ctx, fakeClient, natsBackend, happyCleaner, defaultLogger, recorder, defaultSubConfig, unhappyValidator)
			},
			wantReconcileResultMatcher: Equal(ctrl.Result{}),
			wantReconcileErrorMatcher:  Equal(validatorErr),
		},
		{
			name:              "Return error and default Result{} when event type cleaner returns error",
			givenSubscription: testSub,
			givenReconcilerSetup: func() *natsreconciler.Reconciler {
				fakeClient := clientBuilderWithScheme(g).WithObjects(testSub).Build()
				natsBackend := &mocks.NatsBackend{}
				natsBackend.On("Initialize", mock.Anything).Return(nil)
				natsBackend.On("SyncSubscription", mock.Anything).Return(nil)
				return natsreconciler.NewReconciler(ctx, fakeClient, natsBackend, unhappyCleaner, defaultLogger, recorder, defaultSubConfig, happyValidator)
			},
			wantReconcileResultMatcher: Equal(ctrl.Result{}),
			wantReconcileErrorMatcher:  Equal(cleanerErr),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			reconciler := testCase.givenReconcilerSetup()
			req := ctrl.Request{NamespacedName: types.NamespacedName{
				Namespace: tc.givenSubscription.Namespace,
				Name:      tc.givenSubscription.Name,
			}}
			res, err := reconciler.Reconcile(context.Background(), req)
			g.Expect(res).To(tc.wantReconcileResultMatcher)
			g.Expect(err).To(tc.wantReconcileErrorMatcher)
		})
	}
}

func clientBuilderWithScheme(g *WithT) *ctrlfake.ClientBuilder {
	eventingScheme := scheme.Scheme
	g.Expect(eventingv1alpha1.AddToScheme(eventingScheme)).To(Succeed())
	return ctrlfake.NewClientBuilder().WithScheme(eventingScheme)
}
