package jetstream

import (
	"fmt"
	"log"
	"testing"

	reconcilertestingv2 "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/reconcilertesting"
	testingv1 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	testingv2 "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
)

func Test_ValidationWebhook(t *testing.T) {
	ens := setupTestEnsemble(t)
	defer cleanupResources(ens)

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
				testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantError: func(subName string) error {
				return reconcilertestingv2.GenerateInvalidSubscriptionError(subName,
					eventingv1alpha2.EmptyErrDetail, eventingv1alpha2.SourcePath)
			},
		},
		{
			name: "should fail to create subscription with invalid event types",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithStandardTypeMatching(),
				testingv2.WithSource("source"),
				testingv2.WithTypes([]string{}),
				testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantError: func(subName string) error {
				return reconcilertestingv2.GenerateInvalidSubscriptionError(subName,
					eventingv1alpha2.EmptyErrDetail, eventingv1alpha2.TypesPath)
			},
		},
		{
			name: "should fail to create subscription with invalid config",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithStandardTypeMatching(),
				testingv2.WithSource("source"),
				testingv2.WithOrderCreatedV1Event(),
				testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
				testingv2.WithMaxInFlightMessages("invalid"),
			},
			wantError: func(subName string) error {
				return reconcilertestingv2.GenerateInvalidSubscriptionError(subName,
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
				return reconcilertestingv2.GenerateInvalidSubscriptionError(subName,
					eventingv1alpha2.SuffixMissingErrDetail, eventingv1alpha2.SinkPath)
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			ens := ens

			t.Log("creating the k8s subscription")
			sub := reconcilertestingv2.NewSubscription(ens.TestEnsemble, tc.givenSubscriptionOpts...)

			reconcilertestingv2.EnsureNamespaceCreatedForSub(ens.TestEnsemble, sub)

			// attempt to create subscription
			reconcilertestingv2.EnsureK8sResourceNotCreated(ens.TestEnsemble, t, sub, tc.wantError(sub.Name))
		})
	}
	t.Cleanup(ens.Cancel)
}

// TestUnavailableNATSServer tests if a subscription is reconciled properly when the NATS backend is unavailable.
func TestUnavailableNATSServer(t *testing.T) {
	// prepare the test resources and run test reconciler
	ens := setupTestEnsemble(t)
	defer cleanupResources(ens)

	// prepare the subscription
	sub := reconcilertestingv2.CreateSubscription(ens.TestEnsemble,
		testingv2.WithSourceAndType(testingv2.EventSourceClean, testingv2.OrderCreatedEventType),
		testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
	)

	t.Log("test the subscription was reconciled properly and has the expected status")
	reconcilertestingv2.TestSubscriptionOnK8s(ens.TestEnsemble, sub,
		testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
		testingv2.HaveSubscriptionReady(),
		testingv2.HaveStatusTypes([]eventingv1alpha2.EventType{
			{
				OriginalType: testingv2.OrderCreatedEventType,
				CleanType:    testingv2.OrderCreatedEventType,
			},
		}),
	)

	t.Log("stopping NATS server should trigger the subscription become un-ready")
	ens.NatsServer.Shutdown()
	reconcilertestingv2.TestSubscriptionOnK8s(ens.TestEnsemble, sub,
		testingv2.HaveSubscriptionNotReady(),
	)

	// should trigger the subscription become ready again
	t.Log("startup the NATS Server and test that sub becomes ready")
	ens.NatsServer = testingv1.StartDefaultJetStreamServer(ens.NatsPort)
	reconcilertestingv2.TestSubscriptionOnK8s(ens.TestEnsemble, sub, testingv2.HaveSubscriptionReady())

	t.Cleanup(ens.Cancel)
}

// Check the reconciler idempotency by adding a label to the Kyma subscription.
func TestIdempotency(t *testing.T) {
	// prepare the test resources and run test reconciler
	ens := setupTestEnsemble(t)
	defer cleanupResources(ens)

	t.Log("create the subscription")
	sub := reconcilertestingv2.CreateSubscription(ens.TestEnsemble,
		testingv2.WithTypeMatchingExact(),
		testingv2.WithSourceAndType(testingv2.EventSourceClean, testingv2.OrderCreatedEventType),
		testingv2.WithMaxInFlight(ens.DefaultSubscriptionConfig.MaxInFlightMessages),
		testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
	)

	t.Log("test the subscription was properly reconciled")
	reconcilertestingv2.TestSubscriptionOnK8s(ens.TestEnsemble, sub,
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
		testSubscriptionOnNATS(ens, sub, testingv2.OrderCreatedEventType,
			reconcilertestingv2.BeExistingSubscription(),
			reconcilertestingv2.BeValidSubscription(),
			reconcilertestingv2.BeNatsSubWithMaxPending(ens.DefaultSubscriptionConfig.MaxInFlightMessages),
			reconcilertestingv2.BeJetStreamSubscriptionWithSubject(testingv2.EventSource,
				testingv2.OrderCreatedEventType, eventingv1alpha2.TypeMatchingExact, ens.jetStreamBackend.Config),
		)
	}
	testNatsSub()

	t.Log("add a label to subscription to trigger the reconciliation")
	k8sSubBefore := sub.DeepCopy()
	newLabels := map[string]string{
		"newLabel": "label",
	}
	sub.ObjectMeta.Labels = newLabels
	require.NoError(t, ens.K8sClient.Update(ens.Ctx, sub))

	// check the labels got updated
	assert.Equal(t, sub.Labels, newLabels)

	// set the fields which should be change anyway during the resource update
	t.Log("check that reconciliation did no change to the both Kyma and NATS subscriptions")
	require.Equal(t, k8sSubBefore.Spec, sub.Spec)
	require.Equal(t, k8sSubBefore.Status, sub.Status)
	testNatsSub()

	t.Cleanup(ens.Cancel)
}

// TestCreateSubscription tests if subscriptions get created properly by the reconciler.
func TestCreateSubscription(t *testing.T) {
	ens := setupTestEnsemble(t)
	defer cleanupResources(ens)

	var testCases = []struct {
		name                  string
		givenSubscriptionOpts []testingv2.SubscriptionOpt
		want                  reconcilertestingv2.Want
	}{
		{
			name: "create and delete",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, testingv2.OrderCreatedEventType),
				testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
				testingv2.WithMaxInFlight(ens.DefaultSubscriptionConfig.MaxInFlightMessages),
				testingv2.WithFinalizers([]string{}),
			},
			want: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveMaxInFlight(ens.DefaultSubscriptionConfig.MaxInFlightMessages),
					testingv2.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					testingv2.OrderCreatedEventType: {
						reconcilertestingv2.BeExistingSubscription(),
						reconcilertestingv2.BeValidSubscription(),
						reconcilertestingv2.BeJetStreamSubscriptionWithSubject(testingv2.EventSource,
							testingv2.OrderCreatedEventType, eventingv1alpha2.TypeMatchingExact,
							ens.jetStreamBackend.Config),
					},
				},
			},
		},
		{
			name: "valid sink; with port and endpoint",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(emptyEventSource, reconcilertestingv2.NewUncleanEventType("0")),
				testingv2.WithSinkURL(reconcilertestingv2.ValidSinkURL(ens.TestEnsemble, ":8080", "/myEndpoint")),
			},
			want: reconcilertestingv2.Want{
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
				testingv2.WithSourceAndType(testingv2.EventSource, reconcilertestingv2.NewUncleanEventType("0")),
				testingv2.WithSinkURL(
					testingv2.ValidSinkURL(ens.SubscriberSvc.Namespace, "testapp"),
				),
			},
			want: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(
						reconcilertestingv2.ConditionInvalidSink(
							"failed to validate subscription sink URL. It is not a valid cluster local svc: Service \"testapp\" not found",
						)),
				},
				K8sEvents: []corev1.Event{
					reconcilertestingv2.EventInvalidSink("Sink does not correspond to a valid cluster local svc")},
			},
		},
		{
			name: "empty protocol, protocol setting and dialect",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, testingv2.OrderCreatedEventType),
				testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
				testingv2.WithFinalizers([]string{}),
			},
			want: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					testingv2.OrderCreatedEventType: {
						reconcilertestingv2.BeExistingSubscription(),
						reconcilertestingv2.BeValidSubscription(),
						reconcilertestingv2.BeJetStreamSubscriptionWithSubject(
							testingv2.EventTypePrefixEmpty,
							testingv2.OrderCreatedEventType,
							eventingv1alpha2.TypeMatchingExact,
							ens.jetStreamBackend.Config,
						),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// The new t instance needs to be used for making assertions. If ignored, errors like this will be printed:
			//	test executed panic(nil) or runtime.Goexit: subtest may have called FailNow on a parent test
			ens := ens
			ens.G = gomega.NewGomegaWithT(t)

			t.Log("creating the k8s subscription")
			sub := reconcilertestingv2.CreateSubscription(ens.TestEnsemble, tc.givenSubscriptionOpts...)

			t.Log("testing the k8s subscription")
			reconcilertestingv2.TestSubscriptionOnK8s(ens.TestEnsemble, sub, tc.want.K8sSubscription...)

			t.Log("testing the k8s events")
			reconcilertestingv2.TestEventsOnK8s(ens.TestEnsemble, tc.want.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.want.NatsSubscriptions {
				log.Printf("eventType: %v", eventType)
				testSubscriptionOnNATS(ens, sub, eventType, matchers...)
			}

			t.Log("testing the deletion of the subscription")
			testSubscriptionDeletion(ens, sub)

			t.Log("testing the deletion of the NATS subscription(s)")
			for _, eventType := range sub.Spec.Types {
				ensureNATSSubscriptionIsDeleted(ens, sub, eventType)
			}
		})
	}
	t.Cleanup(ens.Cancel)
}

// TestChangeSubscription tests if existing subscriptions are reconciled properly after getting changed.
func TestChangeSubscription(t *testing.T) {
	ens := setupTestEnsemble(t)
	defer cleanupResources(ens)

	var testCases = []struct {
		name                  string
		givenSubscriptionOpts []testingv2.SubscriptionOpt
		wantBefore            reconcilertestingv2.Want
		changeSubscription    func(subscription *eventingv1alpha2.Subscription)
		wantAfter             reconcilertestingv2.Want
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
				testingv2.WithSourceAndType(testingv2.EventSource, reconcilertestingv2.NewCleanEventType("0")),
				// valid sink
				testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveSubscriptionReady(),
					// for each filter we want to have a clean event type
					testingv2.HaveTypes([]string{
						reconcilertestingv2.NewCleanEventType("0"),
					}),
				},
				K8sEvents: nil,
				// ensure that each filter results in a NATS consumer
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					reconcilertestingv2.NewCleanEventType("0"): {
						reconcilertestingv2.BeExistingSubscription(),
						reconcilertestingv2.BeValidSubscription(),
						reconcilertestingv2.BeJetStreamSubscriptionWithSubject(
							testingv2.EventTypePrefix, reconcilertestingv2.NewCleanEventType("0"),
							eventingv1alpha2.TypeMatchingExact, ens.jetStreamBackend.Config,
						),
					},
				},
			},
			changeSubscription: func(subscription *eventingv1alpha2.Subscription) {
				testingv2.AddEventType(reconcilertestingv2.NewCleanEventType("1"), subscription)

				// induce an error by making the sink invalid
				subscription.Spec.Sink = testingv2.ValidSinkURL(subscription.Namespace, "invalid")
			},
			wantAfter: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveSubscriptionNotReady(),
					// for each filter we want to have a clean event type
					testingv2.HaveTypes([]string{
						reconcilertestingv2.NewCleanEventType("0"),
						reconcilertestingv2.NewCleanEventType("1"),
					}),
				},
				K8sEvents: nil,
				// ensure that each filter results in a NATS consumer
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					reconcilertestingv2.NewCleanEventType("0"): {
						reconcilertestingv2.BeExistingSubscription(),
					},
					// the newly added filter is not synced to NATS as the sink is invalid
					reconcilertestingv2.NewCleanEventType("1"): {
						gomega.Not(reconcilertestingv2.BeExistingSubscription()),
					},
				},
			},
		},
		{
			name: "CleanEventTypes; change filters",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, reconcilertestingv2.NewCleanEventType("0")),
				testingv2.WithSourceAndType(testingv2.EventSource, reconcilertestingv2.NewCleanEventType("1")),
				testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveTypes([]string{
						reconcilertestingv2.NewCleanEventType("0"),
						reconcilertestingv2.NewCleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha2.Subscription) {
				// change all the filters by adding "alpha" to the event type
				for i, eventType := range subscription.Spec.Types {
					subscription.Spec.Types[i] = fmt.Sprintf("%salpha", eventType)
				}
			},
			wantAfter: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveTypes([]string{
						reconcilertestingv2.NewCleanEventType("0alpha"),
						reconcilertestingv2.NewCleanEventType("1alpha"),
					}),
				},
			},
		},
		{
			name: "CleanEventTypes; delete a filter",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, reconcilertestingv2.NewCleanEventType("0")),
				testingv2.WithSourceAndType(testingv2.EventSource, reconcilertestingv2.NewCleanEventType("1")),
				testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveTypes([]string{
						reconcilertestingv2.NewCleanEventType("0"),
						reconcilertestingv2.NewCleanEventType("1"),
					}),
				},
			},
			changeSubscription: func(subscription *eventingv1alpha2.Subscription) {
				subscription.Spec.Types = subscription.Spec.Types[:1]
			},
			wantAfter: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveTypes([]string{
						reconcilertestingv2.NewCleanEventType("0"),
					}),
				},
			},
		},
		{
			name: "change configuration",
			givenSubscriptionOpts: []testingv2.SubscriptionOpt{
				testingv2.WithTypeMatchingExact(),
				testingv2.WithSourceAndType(testingv2.EventSource, testingv2.OrderCreatedEventType),
				testingv2.WithMaxInFlight(ens.DefaultSubscriptionConfig.MaxInFlightMessages),
				testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
			},
			wantBefore: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveMaxInFlight(ens.DefaultSubscriptionConfig.MaxInFlightMessages),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					testingv2.OrderCreatedEventType: {
						reconcilertestingv2.BeExistingSubscription(),
						reconcilertestingv2.BeValidSubscription(),
						reconcilertestingv2.BeNatsSubWithMaxPending(ens.DefaultSubscriptionConfig.MaxInFlightMessages),
					},
				},
			},
			changeSubscription: func(subscription *eventingv1alpha2.Subscription) {
				subscription.Spec.Config = map[string]string{
					eventingv1alpha2.MaxInFlightMessages: "101",
				}
			},
			wantAfter: reconcilertestingv2.Want{
				K8sSubscription: []gomegatypes.GomegaMatcher{
					testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
					testingv2.HaveMaxInFlight(101),
				},
				NatsSubscriptions: map[string][]gomegatypes.GomegaMatcher{
					testingv2.OrderCreatedEventType: {
						reconcilertestingv2.BeExistingSubscription(),
						reconcilertestingv2.BeValidSubscription(),
						reconcilertestingv2.BeJetStreamSubscriptionWithSubject(
							testingv2.EventTypePrefix, testingv2.OrderCreatedEventType, eventingv1alpha2.TypeMatchingExact,
							ens.jetStreamBackend.Config,
						),
						reconcilertestingv2.BeNatsSubWithMaxPending(101),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// The new t instance needs to be used for making assertions. If ignored, errors like this will be printed:
			// testing.go:1336: test executed panic(nil) or runtime.Goexit: subtest may have called FailNow on a parent test
			ens := ens
			ens.G = gomega.NewGomegaWithT(t)

			// given
			t.Log("creating the k8s subscription")
			sub := reconcilertestingv2.CreateSubscription(ens.TestEnsemble, tc.givenSubscriptionOpts...)

			t.Log("testing the k8s subscription")
			reconcilertestingv2.TestSubscriptionOnK8s(ens.TestEnsemble, sub, tc.wantBefore.K8sSubscription...)

			t.Log("testing the k8s events")
			reconcilertestingv2.TestEventsOnK8s(ens.TestEnsemble, tc.wantBefore.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.wantBefore.NatsSubscriptions {
				testSubscriptionOnNATS(ens, sub, eventType, matchers...)
			}

			// when
			t.Log("change and update the subscription")
			reconcilertestingv2.EventuallyUpdateSubscriptionOnK8s(ens.Ctx, ens.TestEnsemble,
				sub, func(sub *eventingv1alpha2.Subscription) error {
					tc.changeSubscription(sub)
					return ens.K8sClient.Update(ens.Ctx, sub)
				})

			// then
			t.Log("testing the k8s subscription")
			reconcilertestingv2.TestSubscriptionOnK8s(ens.TestEnsemble, sub, tc.wantAfter.K8sSubscription...)

			t.Log("testing the k8s events")
			reconcilertestingv2.TestEventsOnK8s(ens.TestEnsemble, tc.wantAfter.K8sEvents...)

			t.Log("testing the nats subscriptions")
			for eventType, matchers := range tc.wantAfter.NatsSubscriptions {
				testSubscriptionOnNATS(ens, sub, eventType, matchers...)
			}

			t.Log("testing the deletion of the subscription")
			testSubscriptionDeletion(ens, sub)

			t.Log("testing the deletion of the NATS subscription(s)")
			for _, filter := range sub.Spec.Types {
				ensureNATSSubscriptionIsDeleted(ens, sub, filter)
			}
		})
	}
	t.Cleanup(ens.Cancel)
}

// TestEmptyEventTypePrefix tests if a subscription is reconciled properly if the EventTypePrefix is empty.
func TestEmptyEventTypePrefix(t *testing.T) {
	ens := setupTestEnsemble(t)
	defer cleanupResources(ens)

	// when
	sub := reconcilertestingv2.CreateSubscription(ens.TestEnsemble,
		testingv2.WithTypeMatchingExact(),
		testingv2.WithSourceAndType(emptyEventSource, testingv2.OrderCreatedEventTypePrefixEmpty),
		testingv2.WithSinkURLFromSvc(ens.SubscriberSvc),
	)

	// then
	reconcilertestingv2.TestSubscriptionOnK8s(ens.TestEnsemble, sub,
		testingv2.HaveTypes([]string{testingv2.OrderCreatedEventTypePrefixEmpty}),
		testingv2.HaveCondition(testingv2.DefaultReadyCondition()),
		testingv2.HaveSubscriptionReady(),
	)

	expectedNatsSubscription := []gomegatypes.GomegaMatcher{
		reconcilertestingv2.BeExistingSubscription(),
		reconcilertestingv2.BeValidSubscription(),
		reconcilertestingv2.BeJetStreamSubscriptionWithSubject(testingv2.EventSource,
			testingv2.OrderCreatedEventTypePrefixEmpty, eventingv1alpha2.TypeMatchingExact, ens.jetStreamBackend.Config),
	}

	testSubscriptionOnNATS(ens, sub, testingv2.OrderCreatedEventTypePrefixEmpty, expectedNatsSubscription...)

	testSubscriptionDeletion(ens, sub)
	ensureNATSSubscriptionIsDeleted(ens, sub, testingv2.OrderCreatedEventTypePrefixEmpty)

	t.Cleanup(ens.Cancel)
}
