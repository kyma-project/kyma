package test

import (
	"fmt"
	"log"
	"os"
	"testing"

	testingv1 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	testingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
)

// TestMain pre-hook and post-hook to run before and after all tests.
func TestMain(m *testing.M) {
	// Note: The setup will provision a single K8s env and
	// all the tests need to create and use a separate namespace

	// setup env test
	if err := setupSuite(); err != nil {
		panic(err)
	}

	// run tests
	code := m.Run()

	// tear down test env
	if err := tearDownSuite(); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func Test_ValidationWebhook(t *testing.T) {
	t.Parallel()
	var testCases = []struct {
		name                  string
		givenSubscriptionOpts []testingv2.SubscriptionOpt
		wantError             func(subName string) error
	}{
		{
			name: "should fail to create subscription with invalid event source",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithStandardTypeMatching(),
				testingv2.WithSource(""),
				testingv2.WithOrderCreatedV1Event(),
				testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
			},
			wantError: func(subName string) error {
				return GenerateInvalidSubscriptionError(subName,
					eventingv1alpha2.EmptyErrDetail, eventingv1alpha2.SourcePath)
			},
		},
		{
			name: "should fail to create subscription with invalid event types",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithStandardTypeMatching(),
				testingv2.WithSource("source"),
				testingv2.WithTypes([]string{}),
				testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
			},
			wantError: func(subName string) error {
				return GenerateInvalidSubscriptionError(subName,
					eventingv1alpha2.EmptyErrDetail, eventingv1alpha2.TypesPath)
			},
		},
		{
			name: "should fail to create subscription with invalid config",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithStandardTypeMatching(),
				testingv2.WithSource("source"),
				testingv2.WithOrderCreatedV1Event(),
				testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
				testingv2.WithMaxInFlightMessages("invalid"),
			},
			wantError: func(subName string) error {
				return GenerateInvalidSubscriptionError(subName,
					eventingv1alpha2.StringIntErrDetail, eventingv1alpha2.ConfigPath)
			},
		},
		{
			name: "should fail to create subscription with invalid sink",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithStandardTypeMatching(),
				testingv2.WithSource("source"),
				testingv2.WithOrderCreatedV1Event(),
				testingv2.WithSink("https://svc2.test.local"),
			},
			wantError: func(subName string) error {
				return GenerateInvalidSubscriptionError(subName,
					eventingv1alpha2.SuffixMissingErrDetail, eventingv1alpha2.SinkPath)
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			t.Log("creating the k8s subscription")
			sub := NewSubscription(jsTestEnsemble.Ensemble, tc.givenSubscriptionOpts...)

			EnsureNamespaceCreatedForSub(t, jsTestEnsemble.Ensemble, sub)

			// attempt to create subscription
			EnsureK8sResourceNotCreated(t, jsTestEnsemble.Ensemble, sub, tc.wantError(sub.Name))
		})
	}
}

// TestUnavailableNATSServer tests if a subscription is reconciled properly when the NATS backend is unavailable.
func Test_UnavailableNATSServer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// prepare the subscription
	sub := CreateSubscription(t, jsTestEnsemble.Ensemble,
		testingv2.WithSourceAndType(testingv2.EventSourceClean, testingv2.OrderCreatedEventType),
		testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
	)

	// test the subscription was reconciled properly and has the expected status
	CheckSubscriptionOnK8s(g, jsTestEnsemble.Ensemble, sub,
		testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
		testingv2.HaveSubscriptionReady(),
		testingv2.HaveStatusTypes([]eventingv1alpha2.EventType{
			{
				OriginalType: testingv2.OrderCreatedEventType,
				CleanType:    testingv2.OrderCreatedEventType,
			},
		}),
	)

	// stopping NATS server should trigger the subscription become un-ready
	jsTestEnsemble.NatsServer.Shutdown()
	CheckSubscriptionOnK8s(g, jsTestEnsemble.Ensemble, sub,
		testingv2.HaveSubscriptionNotReady(),
	)

	// should trigger the subscription become ready again
	jsTestEnsemble.NatsServer = testingv1.StartDefaultJetStreamServer(jsTestEnsemble.NatsPort)
	CheckSubscriptionOnK8s(g, jsTestEnsemble.Ensemble, sub, testingv2.HaveSubscriptionReady())
}

// Check the reconciler idempotency by adding a label to the Kyma subscription.
func Test_Idempotency(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Log("create the subscription")
	sub := CreateSubscription(t, jsTestEnsemble.Ensemble,
		testingv2.WithTypeMatchingExact(),
		testingv2.WithSourceAndType(testingv2.EventSourceClean, testingv2.OrderCreatedEventType),
		testingv2.WithMaxInFlight(jsTestEnsemble.DefaultSubscriptionConfig.MaxInFlightMessages),
		testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
	)

	t.Log("test the subscription was properly reconciled")
	CheckSubscriptionOnK8s(g, jsTestEnsemble.Ensemble, sub,
		testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
		testingv2.HaveSubscriptionReady(),
		testingv2.HaveStatusTypes([]eventingv1alpha2.EventType{
			{
				OriginalType: testingv2.OrderCreatedEventType,
				CleanType:    testingv2.OrderCreatedEventType,
			},
		}),
	)

	t.Log("test the subscription on NATS was created as expected")
	testNatsSub := func() {
		testSubscriptionOnNATS(g, sub, testingv2.OrderCreatedEventType,
			BeExistingSubscription(),
			BeValidSubscription(),
			BeNatsSubWithMaxPending(jsTestEnsemble.DefaultSubscriptionConfig.MaxInFlightMessages),
			BeJetStreamSubscriptionWithSubject(testingv2.EventSource,
				testingv2.OrderCreatedEventType, eventingv1alpha2.TypeMatchingExact, jsTestEnsemble.jetStreamBackend.Config),
		)
	}
	testNatsSub()

	t.Log("add a label to subscription to trigger the reconciliation")
	k8sSubBefore := sub.DeepCopy()
	newLabels := map[string]string{
		"newLabel": "label",
	}
	sub.ObjectMeta.Labels = newLabels
	require.NoError(t, jsTestEnsemble.K8sClient.Update(jsTestEnsemble.Ctx, sub))

	// check the labels got updated
	assert.Equal(t, sub.Labels, newLabels)

	// set the fields which should be change anyway during the resource update
	t.Log("check that reconciliation did no change to the both Kyma and NATS subscriptions")
	require.Equal(t, k8sSubBefore.Spec, sub.Spec)
	require.Equal(t, k8sSubBefore.Status, sub.Status)
	testNatsSub()
}

// TestCreateSubscription tests if subscriptions get created properly by the reconciler.
func Test_CreateSubscription(t *testing.T) {
	t.Parallel()
	var testCases = []struct {
		name                  string
		givenSubscriptionOpts []testingv2.SubscriptionOpt
		want                  Want
	}{
		{
			name: "create and delete",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, testingv2.OrderCreatedEventType),
				testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
				testingv2.WithMaxInFlight(jsTestEnsemble.DefaultSubscriptionConfig.MaxInFlightMessages),
				testingv2.WithWebhookForNATS(),
				testingv2.WithFinalizers([]string{}),
			},
			want: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveMaxInFlight(jsTestEnsemble.DefaultSubscriptionConfig.MaxInFlightMessages),
					testingv2.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					testingv2.OrderCreatedEventType: {
						BeExistingSubscription(),
						BeValidSubscription(),
						BeJetStreamSubscriptionWithSubject(testingv2.EventSource,
							testingv2.OrderCreatedEventType, eventingv1alpha2.TypeMatchingExact,
							jsTestEnsemble.jetStreamBackend.Config),
					},
				},
			},
		},
		{
			name: "valid sink; with port and endpoint",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(emptyEventSource, NewUncleanEventType("0")),
				testingv2.WithSinkURL(ValidSinkURL(jsTestEnsemble.Ensemble, ":8080", "/myEndpoint")),
			},
			want: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveSubscriptionReady(),
				},
			},
		},
		{
			name: "invalid sink; not a valid cluster local service",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, NewUncleanEventType("0")),
				testingv2.WithWebhookForNATS(),
				testingv2.WithSinkURL(
					testingv2.ValidSinkURL(jsTestEnsemble.SubscriberSvc.Namespace, "testapp"),
				),
			},
			want: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(
						ConditionInvalidSink(
							"failed to validate subscription sink URL. It is not a valid cluster local svc: Service \"testapp\" not found",
						)),
				},
				K8sEvents: []corev1.Event{
					EventInvalidSink("Sink does not correspond to a valid cluster local svc")},
			},
		},
		{
			name: "empty protocol, protocol setting and dialect",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, testingv2.OrderCreatedEventType),
				testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
				testingv2.WithFinalizers([]string{}),
			},
			want: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					testingv2.OrderCreatedEventType: {
						BeExistingSubscription(),
						BeValidSubscription(),
						BeJetStreamSubscriptionWithSubject(
							testingv2.EventTypePrefixEmpty,
							testingv2.OrderCreatedEventType,
							eventingv1alpha2.TypeMatchingExact,
							jsTestEnsemble.jetStreamBackend.Config,
						),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewGomegaWithT(t)

			t.Log("creating the k8s subscription")
			sub := CreateSubscription(t, jsTestEnsemble.Ensemble, tc.givenSubscriptionOpts...)

			t.Log("testing the k8s subscription")
			CheckSubscriptionOnK8s(g, jsTestEnsemble.Ensemble, sub, tc.want.K8sSubscription...)

			t.Log("testing the k8s events")
			CheckEventsOnK8s(g, jsTestEnsemble.Ensemble, tc.want.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.want.NatsSubscriptions {
				log.Printf("eventType: %v", eventType)
				testSubscriptionOnNATS(g, sub, eventType, matchers...)
			}

			t.Log("testing the deletion of the subscription")
			testSubscriptionDeletion(g, sub)

			t.Log("testing the deletion of the NATS subscription(s)")
			for _, eventType := range sub.Spec.Types {
				ensureNATSSubscriptionIsDeleted(g, sub, eventType)
			}
		})
	}
}

// TestChangeSubscription tests if existing subscriptions are reconciled properly after getting changed.
func Test_ChangeSubscription(t *testing.T) {
	t.Parallel()
	var testCases = []struct {
		name                  string
		givenSubscriptionOpts []testingv2.SubscriptionOpt
		wantBefore            Want
		changeSubscription    func(subscription *eventingv1alpha2.Subscription)
		wantAfter             Want
	}{
		{
			// Ensure subscriptions on NATS are not added when adding
			// a Subscription filter and providing an invalid sink to prevent event-loss.
			// The reason for this is that the dispatcher will retry only for a finite number and then give up.
			// Since the sink is invalid, the dispatcher cannot dispatch the event
			// and will stop if the maximum number of retries is reached.
			name: "Disallow the creation of a NATS subscription with an invalid sink",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, NewCleanEventType("0")),
				// valid sink
				testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
			},
			wantBefore: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveSubscriptionReady(),
					// for each filter we want to have a clean event type
					testingv2.HaveTypes([]string{
						NewCleanEventType("0"),
					}),
				},
				K8sEvents: nil,
				// ensure that each filter results in a NATS consumer
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					NewCleanEventType("0"): {
						BeExistingSubscription(),
						BeValidSubscription(),
						BeJetStreamSubscriptionWithSubject(
							testingv2.EventTypePrefix, NewCleanEventType("0"),
							eventingv1alpha2.TypeMatchingExact, jsTestEnsemble.jetStreamBackend.Config,
						),
					},
				},
			},
			changeSubscription: func(subscription *eventingv1alpha2.Subscription) {
				testingv2.AddEventType(NewCleanEventType("1"), subscription)

				// induce an error by making the sink invalid
				subscription.Spec.Sink = testingv2.ValidSinkURL(subscription.Namespace, "invalid")
			},
			wantAfter: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveSubscriptionNotReady(),
					// for each filter we want to have a clean event type
					testingv2.HaveTypes([]string{
						NewCleanEventType("0"),
						NewCleanEventType("1"),
					}),
				},
				K8sEvents: nil,
				// ensure that each filter results in a NATS consumer
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					NewCleanEventType("0"): {
						BeExistingSubscription(),
					},
					// the newly added filter is not synced to NATS as the sink is invalid
					NewCleanEventType("1"): {
						gomega.Not(BeExistingSubscription()),
					},
				},
			},
		},
		{
			name: "CleanEventTypes; change filters",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, NewCleanEventType("0")),
				testingv2.WithSourceAndType(testingv2.EventSource, NewCleanEventType("1")),
				testingv2.WithWebhookForNATS(),
				testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
			},
			wantBefore: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveTypes([]string{
						NewCleanEventType("0"),
						NewCleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha2.Subscription) {
				// change all the filters by adding "alpha" to the event type
				for i, eventType := range subscription.Spec.Types {
					subscription.Spec.Types[i] = fmt.Sprintf("%salpha", eventType)
				}
			},
			wantAfter: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveTypes([]string{
						NewCleanEventType("0alpha"),
						NewCleanEventType("1alpha"),
					}),
				},
			},
		},
		{
			name: "CleanEventTypes; delete a filter",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, NewCleanEventType("0")),
				testingv2.WithSourceAndType(testingv2.EventSource, NewCleanEventType("1")),
				testingv2.WithWebhookForNATS(),
				testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
			},
			wantBefore: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveTypes([]string{
						NewCleanEventType("0"),
						NewCleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha2.Subscription) {
				subscription.Spec.Types = subscription.Spec.Types[:1]
			},
			wantAfter: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveTypes([]string{
						NewCleanEventType("0"),
					}),
				},
			},
		},
		{
			name: "change configuration",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, testingv2.OrderCreatedEventType),
				testingv2.WithMaxInFlight(jsTestEnsemble.DefaultSubscriptionConfig.MaxInFlightMessages),
				testingv2.WithWebhookForNATS(),
				testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
			},
			wantBefore: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveMaxInFlight(jsTestEnsemble.DefaultSubscriptionConfig.MaxInFlightMessages),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					testingv2.OrderCreatedEventType: {
						BeExistingSubscription(),
						BeValidSubscription(),
						BeNatsSubWithMaxPending(jsTestEnsemble.DefaultSubscriptionConfig.MaxInFlightMessages),
					},
				},
			},
			changeSubscription: func(subscription *eventingv1alpha2.Subscription) {
				subscription.Spec.Config = map[string]string{
					eventingv1alpha2.MaxInFlightMessages: "101",
				}
			},
			wantAfter: Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveMaxInFlight(101),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					testingv2.OrderCreatedEventType: {
						BeExistingSubscription(),
						BeValidSubscription(),
						BeJetStreamSubscriptionWithSubject(
							testingv2.EventTypePrefix, testingv2.OrderCreatedEventType, eventingv1alpha2.TypeMatchingExact,
							jsTestEnsemble.jetStreamBackend.Config,
						),
						BeNatsSubWithMaxPending(101),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewGomegaWithT(t)

			// given
			t.Log("creating the k8s subscription")
			sub := CreateSubscription(t, jsTestEnsemble.Ensemble, tc.givenSubscriptionOpts...)

			t.Log("testing the k8s subscription")
			CheckSubscriptionOnK8s(g, jsTestEnsemble.Ensemble, sub, tc.wantBefore.K8sSubscription...)

			t.Log("testing the k8s events")
			CheckEventsOnK8s(g, jsTestEnsemble.Ensemble, tc.wantBefore.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.wantBefore.NatsSubscriptions {
				testSubscriptionOnNATS(g, sub, eventType, matchers...)
			}

			// when
			t.Log("change and update the subscription")
			require.NoError(t, EventuallyUpdateSubscriptionOnK8s(jsTestEnsemble.Ctx, jsTestEnsemble.Ensemble,
				sub, func(sub *eventingv1alpha2.Subscription) error {
					tc.changeSubscription(sub)
					return jsTestEnsemble.K8sClient.Update(jsTestEnsemble.Ctx, sub)
				}))

			// then
			t.Log("testing the k8s subscription")
			CheckSubscriptionOnK8s(g, jsTestEnsemble.Ensemble, sub, tc.wantAfter.K8sSubscription...)

			t.Log("testing the k8s events")
			CheckEventsOnK8s(g, jsTestEnsemble.Ensemble, tc.wantAfter.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.wantAfter.NatsSubscriptions {
				testSubscriptionOnNATS(g, sub, eventType, matchers...)
			}

			t.Log("testing the deletion of the subscription")
			testSubscriptionDeletion(g, sub)

			t.Log("testing the deletion of the NATS subscription(s)")
			for _, filter := range sub.Spec.Types {
				ensureNATSSubscriptionIsDeleted(g, sub, filter)
			}
		})
	}
}

// TestEmptyEventTypePrefix tests if a subscription is reconciled properly if the EventTypePrefix is empty.
func Test_EmptyEventTypePrefix(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// when
	sub := CreateSubscription(t,
		jsTestEnsemble.Ensemble,
		testingv2.WithTypeMatchingExact(),
		testingv2.WithSourceAndType(emptyEventSource, testingv2.OrderCreatedEventTypePrefixEmpty),
		testingv2.WithSinkURLFromSvc(jsTestEnsemble.SubscriberSvc),
	)

	// then
	CheckSubscriptionOnK8s(g, jsTestEnsemble.Ensemble, sub,
		testingv2.HaveTypes([]string{testingv2.OrderCreatedEventTypePrefixEmpty}),
		testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
		testingv2.HaveSubscriptionReady(),
	)

	expectedNatsSubscription := []gomegatypes.GomegaMatcher{
		BeExistingSubscription(),
		BeValidSubscription(),
		BeJetStreamSubscriptionWithSubject(testingv2.EventSource,
			testingv2.OrderCreatedEventTypePrefixEmpty,
			eventingv1alpha2.TypeMatchingExact,
			jsTestEnsemble.jetStreamBackend.Config,
		),
	}

	testSubscriptionOnNATS(g, sub, testingv2.OrderCreatedEventTypePrefixEmpty, expectedNatsSubscription...)

	testSubscriptionDeletion(g, sub)
	ensureNATSSubscriptionIsDeleted(g, sub, testingv2.OrderCreatedEventTypePrefixEmpty)
}
