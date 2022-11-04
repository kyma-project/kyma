package test

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	reconcilertestingv1 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/stretchr/testify/assert"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	sink "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink/v2"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"

	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

// TestMain pre-hook and post-hook to run before and after all tests
func TestMain(m *testing.M) {
	// Note: The setup will provision a single K8s env and
	// all the tests need to create and use a separate namespace

	// setup env test
	if err := setupSuite(); err != nil {
		fmt.Printf("failed to start test suite")
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

func Test_CreateSubscription(t *testing.T) {
	var testCases = []struct {
		name                     string
		givenSubscriptionFunc    func(namespace string) *eventingv1alpha2.Subscription
		wantSubscriptionMatchers gomegatypes.GomegaMatcher
	}{
		{
			name: "should fail to create subscription if types is empty",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription("test", namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithEmptyTypes(),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "test")),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionNotReady(),
				reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
					eventingv1alpha2.ConditionSubscribed,
					eventingv1alpha2.ConditionReasonSubscriptionCreationFailed,
					corev1.ConditionFalse, "Types are required")),
				reconcilertesting.HaveCleanEventTypesEmpty(),
			),
		},
		{
			name: "should succeed to create subscription if types are non-empty",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription("test", namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "test")),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
					eventingv1alpha2.ConditionSubscriptionActive,
					eventingv1alpha2.ConditionReasonSubscriptionActive,
					corev1.ConditionTrue, "")),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1Event),
					}, {
						OriginalType: fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := gomega.NewGomegaWithT(t)
			ctx := context.Background()

			// create unique namespace for this test run
			testNamespace := getTestNamespace()
			ensureNamespaceCreated(ctx, t, testNamespace)

			// update namespace information in given test assets
			givenSubscription := tc.givenSubscriptionFunc(testNamespace)

			// create a subscriber service
			subscriberSvc := reconcilertesting.NewSubscriberSvc(givenSubscription.Name, testNamespace)
			ensureK8sResourceCreated(ctx, t, subscriberSvc)

			// create subscription
			ensureK8sResourceCreated(ctx, t, givenSubscription)

			// check if the subscription is as required
			getSubscriptionAssert(ctx, g, givenSubscription).Should(tc.wantSubscriptionMatchers)

		})
	}
}

func Test_UpdateSubscription(t *testing.T) {
	var testCases = []struct {
		name                           string
		givenSubscriptionFunc          func(namespace string) *eventingv1alpha2.Subscription
		givenUpdateSubscriptionFunc    func(namespace string) *eventingv1alpha2.Subscription
		wantSubscriptionMatchers       gomegatypes.GomegaMatcher
		wantUpdateSubscriptionMatchers gomegatypes.GomegaMatcher
	}{
		{
			name: "should succeed to update subscription when event type is added",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription("test", namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "test")),
				)
			},
			givenUpdateSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription("test", namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "test")),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
					eventingv1alpha2.ConditionSubscriptionActive,
					eventingv1alpha2.ConditionReasonSubscriptionActive,
					corev1.ConditionTrue, "")),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
			wantUpdateSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
					eventingv1alpha2.ConditionSubscriptionActive,
					eventingv1alpha2.ConditionReasonSubscriptionActive,
					corev1.ConditionTrue, "")),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1Event),
					}, {
						OriginalType: fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
		},
		{
			name: "should succeed to update subscription when event types are updated",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription("test", namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "test")),
				)
			},
			givenUpdateSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription("test", namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0alpha", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1alpha", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "test")),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
					eventingv1alpha2.ConditionSubscriptionActive,
					eventingv1alpha2.ConditionReasonSubscriptionActive,
					corev1.ConditionTrue, "")),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1Event),
					}, {
						OriginalType: fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
			wantUpdateSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
					eventingv1alpha2.ConditionSubscriptionActive,
					eventingv1alpha2.ConditionReasonSubscriptionActive,
					corev1.ConditionTrue, "")),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: fmt.Sprintf("%s0alpha", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s0alpha", reconcilertesting.OrderCreatedV1Event),
					}, {
						OriginalType: fmt.Sprintf("%s1alpha", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s1alpha", reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
		},
		{
			name: "should succeed to update subscription when event type is deleted",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription("test", namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "test")),
				)
			},
			givenUpdateSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription("test", namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "test")),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
					eventingv1alpha2.ConditionSubscriptionActive,
					eventingv1alpha2.ConditionReasonSubscriptionActive,
					corev1.ConditionTrue, "")),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1Event),
					}, {
						OriginalType: fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
			wantUpdateSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
					eventingv1alpha2.ConditionSubscriptionActive,
					eventingv1alpha2.ConditionReasonSubscriptionActive,
					corev1.ConditionTrue, "")),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			g := gomega.NewGomegaWithT(t)
			ctx := context.Background()

			// create unique namespace for this test run
			testNamespace := getTestNamespace()
			ensureNamespaceCreated(ctx, t, testNamespace)

			// update namespace information in given test assets
			givenSubscription := tc.givenSubscriptionFunc(testNamespace)
			givenUpdateSubscription := tc.givenUpdateSubscriptionFunc(testNamespace)

			// create a subscriber service
			subscriberSvc := reconcilertesting.NewSubscriberSvc(givenSubscription.Name, testNamespace)
			ensureK8sResourceCreated(ctx, t, subscriberSvc)

			// create subscription
			ensureK8sResourceCreated(ctx, t, givenSubscription)
			createdSubscription := givenSubscription.DeepCopy()
			// check if the created subscription is correct
			getSubscriptionAssert(ctx, g, createdSubscription).Should(tc.wantSubscriptionMatchers)

			// update subscription
			givenUpdateSubscription.ResourceVersion = createdSubscription.ResourceVersion
			ensureK8sResourceUpdated(ctx, t, givenUpdateSubscription)

			// check if the updated subscription is correct
			getSubscriptionAssert(ctx, g, givenSubscription).Should(tc.wantUpdateSubscriptionMatchers)
		})
	}
}

func Test_FixingSinkAndApiRule(t *testing.T) {
	// common given test assets
	givenSubscriptionFunc := func(namespace, name string) *eventingv1alpha2.Subscription {
		return reconcilertesting.NewSubscription(name, namespace,
			reconcilertesting.WithDefaultSource(),
			reconcilertesting.WithOrderCreatedV1Event(),
			reconcilertesting.WithWebhookAuthForBEB(),
			// The following sink is invalid because it is missing a scheme.
			reconcilertesting.WithSinkMissingScheme(namespace, name),
		)
	}

	givenUpdateSubscriptionFunc := func(namespace, name, sinkFormat, path string) *eventingv1alpha2.Subscription {
		return reconcilertesting.NewSubscription(name, namespace,
			reconcilertesting.WithDefaultSource(),
			reconcilertesting.WithOrderCreatedV1Event(),
			reconcilertesting.WithWebhookAuthForBEB(),
			reconcilertesting.WithSink(fmt.Sprintf(sinkFormat, name, namespace, path)),
		)
	}

	wantSubscriptionMatchers := gomega.And(
		reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
			eventingv1alpha2.ConditionAPIRuleStatus,
			eventingv1alpha2.ConditionReasonAPIRuleStatusNotReady,
			corev1.ConditionFalse,
			sink.MissingSchemeErrMsg,
		)),
	)

	wantUpdateSubscriptionMatchers := gomega.And(
		reconcilertesting.HaveSubscriptionReady(),
		reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
			eventingv1alpha2.ConditionSubscriptionActive,
			eventingv1alpha2.ConditionReasonSubscriptionActive,
			corev1.ConditionTrue,
			"",
		)),
		reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
			eventingv1alpha2.ConditionAPIRuleStatus,
			eventingv1alpha2.ConditionReasonAPIRuleStatusReady,
			corev1.ConditionTrue,
			"",
		)),
	)

	// test cases
	var testCases = []struct {
		name            string
		givenSinkFormat string
	}{
		{
			name:            "should succeed to fix invalid sink with Url with port in subscription",
			givenSinkFormat: "https://%s.%s.svc.cluster.local:8080%s",
		},
		{
			name:            "should succeed to fix invalid sink with Url without port in subscription",
			givenSinkFormat: "https://%s.%s.svc.cluster.local%s",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			g := gomega.NewGomegaWithT(t)
			ctx := context.Background()

			// create unique namespace for this test run
			testNamespace := getTestNamespace()
			ensureNamespaceCreated(ctx, t, testNamespace)
			subName := fmt.Sprintf("test-sink-%s", testNamespace)
			sinkPath := "/path1"

			// update namespace information in given test assets
			givenSubscription := givenSubscriptionFunc(testNamespace, subName)
			givenUpdateSubscription := givenUpdateSubscriptionFunc(testNamespace, subName, tc.givenSinkFormat, sinkPath)

			// create a subscriber service
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subName, testNamespace)
			ensureK8sResourceCreated(ctx, t, subscriberSvc)

			// create subscription
			ensureK8sResourceCreated(ctx, t, givenSubscription)
			createdSubscription := givenSubscription.DeepCopy()
			// check if the created subscription is correct
			getSubscriptionAssert(ctx, g, createdSubscription).Should(wantSubscriptionMatchers)

			// update subscription with valid sink
			givenUpdateSubscription.ResourceVersion = createdSubscription.ResourceVersion
			ensureK8sResourceUpdated(ctx, t, givenUpdateSubscription)

			// check if an APIRule was created for the subscription
			getAPIRuleForASvcAssert(ctx, g, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())

			// check if the created APIRule is as required
			apiRules, err := getAPIRulesList(ctx, subscriberSvc)
			assert.NoError(t, err)
			apiRuleUpdated := filterAPIRulesForASvc(apiRules, subscriberSvc)
			getAPIRuleAssert(ctx, g, &apiRuleUpdated).Should(gomega.And(
				reconcilertestingv1.HaveNotEmptyHost(),
				reconcilertestingv1.HaveNotEmptyAPIRule(),
				reconcilertestingv1.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, sinkPath),
				reconcilertestingv1.HaveAPIRuleOwnersRefs(givenSubscription.UID),
			))

			// update the status of the APIRule to ready (mocking APIGateway controller)
			ensureAPIRuleStatusUpdatedWithStatusReady(ctx, t, &apiRuleUpdated)

			// check if the updated subscription is correct
			getSubscriptionAssert(ctx, g, givenSubscription).Should(wantUpdateSubscriptionMatchers)

			// check if the reconciled subscription has API rule in the status
			assert.Equal(t, givenSubscription.Status.Backend.APIRuleName, apiRuleUpdated.Name)

			// check if the EventMesh mock received requests
			_, postRequests, _ := countEventMeshRequests(
				emTestEnsemble.nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace),
				reconcilertesting.EventMeshOrderCreatedV1Type)

			assert.GreaterOrEqual(t, postRequests, 1)
		})
	}
}
