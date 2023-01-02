package test

import (
	"context"
	"fmt"

	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventMeshtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"

	testingv2 "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/reconcilertesting"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	reconcilertestingv1 "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/stretchr/testify/assert"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
	eventmeshsubmatchers "github.com/kyma-project/kyma/components/eventing-controller/testing/v2/matchers/eventmeshsub"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"

	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

const (
	invalidSinkErrMsg = "failed to validate subscription sink URL. " +
		"It is not a valid cluster local svc: Service \"invalid\" not found"
	testName = "test"
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
		givenSubscriptionFunc func(namespace string) *eventingv1alpha2.Subscription
		wantError             error
	}{
		{
			name: "should fail to create subscription with invalid event source",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithStandardTypeMatching(),
					reconcilertesting.WithSource(""),
					reconcilertesting.WithOrderCreatedV1Event(),
					reconcilertesting.WithValidSink(namespace, "svc"),
				)
			},
			wantError: testingv2.GenerateInvalidSubscriptionError(testName,
				eventingv1alpha2.EmptyErrDetail, eventingv1alpha2.SourcePath),
		},
		{
			name: "should fail to create subscription with invalid event types",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithStandardTypeMatching(),
					reconcilertesting.WithSource("source"),
					reconcilertesting.WithTypes([]string{}),
					reconcilertesting.WithValidSink(namespace, "svc"),
				)
			},
			wantError: testingv2.GenerateInvalidSubscriptionError(testName,
				eventingv1alpha2.EmptyErrDetail, eventingv1alpha2.TypesPath),
		},
		{
			name: "should fail to create subscription with invalid sink",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithStandardTypeMatching(),
					reconcilertesting.WithSource("source"),
					reconcilertesting.WithOrderCreatedV1Event(),
					reconcilertesting.WithSink("https://svc2.test.local"),
				)
			},
			wantError: testingv2.GenerateInvalidSubscriptionError(testName,
				eventingv1alpha2.SuffixMissingErrDetail, eventingv1alpha2.SinkPath),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// create unique namespace for this test run
			testNamespace := getTestNamespace()
			ensureNamespaceCreated(ctx, t, testNamespace)

			// update namespace information in given test assets
			givenSubscription := tc.givenSubscriptionFunc(testNamespace)

			// attempt to create subscription
			ensureK8sResourceNotCreated(ctx, t, givenSubscription, tc.wantError)
		})
	}
}

func Test_CreateSubscription(t *testing.T) {
	t.Parallel()
	var testCases = []struct {
		name                     string
		givenSubscriptionFunc    func(namespace string) *eventingv1alpha2.Subscription
		wantSubscriptionMatchers gomegatypes.GomegaMatcher
		wantEventMeshSubMatchers gomegatypes.GomegaMatcher
		wantEventMeshSubCheck    bool
		wantAPIRuleCheck         bool
		wantSubCreatedEventCheck bool
		wantSubActiveEventCheck  bool
	}{
		{
			name: "should fail to create subscription if sink does not exist",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "invalid")),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionNotReady(),
				reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
					eventingv1alpha2.ConditionAPIRuleStatus,
					eventingv1alpha2.ConditionReasonAPIRuleStatusNotReady,
					corev1.ConditionFalse, invalidSinkErrMsg)),
			),
		},
		{
			name: "should succeed to create subscription if types are non-empty",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer),
				reconcilertesting.HaveSubscriptionActiveCondition(),
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
			wantEventMeshSubMatchers: gomega.And(
				eventmeshsubmatchers.HaveEvents(eventMeshtypes.Events{
					{
						Source: reconcilertesting.EventMeshNamespaceNS,
						Type: fmt.Sprintf("%s.%s.%s0", reconcilertesting.EventMeshPrefix,
							reconcilertesting.ApplicationName, reconcilertesting.OrderCreatedV1Event),
					},
					{
						Source: reconcilertesting.EventMeshNamespaceNS,
						Type: fmt.Sprintf("%s.%s.%s1", reconcilertesting.EventMeshPrefix,
							reconcilertesting.ApplicationName, reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
			wantEventMeshSubCheck: true,
			wantAPIRuleCheck:      true,
		},
		{
			name: "should succeed to create subscription with empty protocol and webhook settings",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithNotCleanType(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionActiveCondition(),
			),
			wantEventMeshSubMatchers: gomega.And(
				// should have default values for protocol and webhook auth
				eventmeshsubmatchers.HaveContentMode(emTestEnsemble.envConfig.ContentMode),
				eventmeshsubmatchers.HaveExemptHandshake(emTestEnsemble.envConfig.ExemptHandshake),
				eventmeshsubmatchers.HaveQoS(eventMeshtypes.Qos(emTestEnsemble.envConfig.Qos)),
				eventmeshsubmatchers.HaveWebhookAuth(eventMeshtypes.WebhookAuth{
					ClientID:     "foo-client-id",
					ClientSecret: "foo-client-secret",
					TokenURL:     emTestEnsemble.envConfig.WebhookTokenEndpoint,
					Type:         eventMeshtypes.AuthTypeClientCredentials,
					GrantType:    eventMeshtypes.GrantTypeClientCredentials,
				}),
			),
			wantAPIRuleCheck:      true,
			wantEventMeshSubCheck: true,
		},
		{
			name: "should succeed to create subscription with EXACT type matching",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithExactTypeMatching(),
					reconcilertesting.WithEventMeshExactType(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer),
				reconcilertesting.HaveSubscriptionActiveCondition(),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: reconcilertesting.EventMeshExactType,
						CleanType:    reconcilertesting.EventMeshExactType,
					},
				}),
			),
			wantEventMeshSubMatchers: gomega.And(
				eventmeshsubmatchers.HaveEvents(eventMeshtypes.Events{
					{
						Source: reconcilertesting.EventMeshNamespaceNS,
						Type:   reconcilertesting.EventMeshExactType,
					},
				}),
			),
			wantEventMeshSubCheck:    true,
			wantAPIRuleCheck:         true,
			wantSubCreatedEventCheck: true,
			wantSubActiveEventCheck:  true,
		},
		{
			name: "should mark a non-cleaned Subscription as ready",
			givenSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithNotCleanSource(),
					reconcilertesting.WithNotCleanType(),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer),
				reconcilertesting.HaveSubscriptionActiveCondition(),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: reconcilertesting.OrderCreatedV1EventNotClean,
						CleanType:    reconcilertesting.OrderCreatedV1Event,
					},
				}),
			),
			wantEventMeshSubMatchers: gomega.And(
				eventmeshsubmatchers.HaveEvents(eventMeshtypes.Events{
					{
						Source: reconcilertesting.EventMeshNamespaceNS,
						Type: fmt.Sprintf("%s.%s.%s", reconcilertesting.EventMeshPrefix,
							reconcilertesting.ApplicationName, reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
			wantEventMeshSubCheck: true,
			wantAPIRuleCheck:      true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
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

			if tc.wantAPIRuleCheck {
				// check if an APIRule was created for the subscription
				getAPIRuleForASvcAssert(ctx, g, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
			}

			if tc.wantEventMeshSubCheck {
				emSub := getEventMeshSubFromMock(givenSubscription.Name, givenSubscription.Namespace)
				g.Expect(emSub).ShouldNot(gomega.BeNil())
				g.Expect(emSub).Should(tc.wantEventMeshSubMatchers)
			}

			if tc.wantSubCreatedEventCheck {
				message := eventingv1alpha2.CreateMessageForConditionReasonSubscriptionCreated(
					emTestEnsemble.nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
				subscriptionCreatedEvent := corev1.Event{
					Reason:  string(eventingv1alpha2.ConditionReasonSubscriptionCreated),
					Message: message,
					Type:    corev1.EventTypeNormal,
				}
				ensureK8sEventReceived(t, subscriptionCreatedEvent, givenSubscription.Namespace)
			}

			if tc.wantSubActiveEventCheck {
				subscriptionActiveEvent := corev1.Event{
					Reason:  string(eventingv1alpha2.ConditionReasonSubscriptionActive),
					Message: "",
					Type:    corev1.EventTypeNormal,
				}
				ensureK8sEventReceived(t, subscriptionActiveEvent, givenSubscription.Namespace)
			}
		})
	}
}

func Test_UpdateSubscription(t *testing.T) {
	t.Parallel()
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
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionActiveCondition(),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
			givenUpdateSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantUpdateSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionActiveCondition(),
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
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionActiveCondition(),
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
			givenUpdateSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0alpha", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1alpha", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantUpdateSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionActiveCondition(),
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
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionActiveCondition(),
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
			givenUpdateSubscriptionFunc: func(namespace string) *eventingv1alpha2.Subscription {
				return reconcilertesting.NewSubscription(testName, namespace,
					reconcilertesting.WithDefaultSource(),
					reconcilertesting.WithTypes([]string{
						fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
					}),
					reconcilertesting.WithWebhookAuthForBEB(),
					reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, testName)),
				)
			},
			wantUpdateSubscriptionMatchers: gomega.And(
				reconcilertesting.HaveSubscriptionActiveCondition(),
				reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
					{
						OriginalType: fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
						CleanType:    fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1Event),
					},
				}),
			),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
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
			ensureK8sSubscriptionUpdated(ctx, t, givenUpdateSubscription)

			// check if the updated subscription is correct
			getSubscriptionAssert(ctx, g, givenSubscription).Should(tc.wantUpdateSubscriptionMatchers)
		})
	}
}

func Test_FixingSinkAndApiRule(t *testing.T) {
	t.Parallel()
	// common given test assets
	givenSubscriptionWithoutSinkFunc := func(namespace, name string) *eventingv1alpha2.Subscription {
		return reconcilertesting.NewSubscription(name, namespace,
			reconcilertesting.WithDefaultSource(),
			reconcilertesting.WithOrderCreatedV1Event(),
			reconcilertesting.WithWebhookAuthForBEB(),
			// The following sink is invalid because it has an invalid svc name
			reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(namespace, "invalid")),
		)
	}

	wantSubscriptionWithoutSinkMatchers := gomega.And(
		reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
			eventingv1alpha2.ConditionAPIRuleStatus,
			eventingv1alpha2.ConditionReasonAPIRuleStatusNotReady,
			corev1.ConditionFalse,
			invalidSinkErrMsg,
		)),
	)

	givenUpdateSubscriptionWithSinkFunc := func(namespace, name, sinkFormat, path string) *eventingv1alpha2.Subscription {
		return reconcilertesting.NewSubscription(name, namespace,
			reconcilertesting.WithDefaultSource(),
			reconcilertesting.WithOrderCreatedV1Event(),
			reconcilertesting.WithWebhookAuthForBEB(),
			reconcilertesting.WithSink(fmt.Sprintf(sinkFormat, name, namespace, path)),
		)
	}

	wantUpdateSubscriptionWithSinkMatchers := gomega.And(
		reconcilertesting.HaveSubscriptionReady(),
		reconcilertesting.HaveSubscriptionActiveCondition(),
		reconcilertesting.HaveAPIRuleTrueStatusCondition(),
	)

	// test cases
	var testCases = []struct {
		name            string
		givenSinkFormat string
	}{
		{
			name:            "should succeed to fix invalid sink with Url and port in subscription",
			givenSinkFormat: "https://%s.%s.svc.cluster.local:8080%s",
		},
		{
			name:            "should succeed to fix invalid sink with Url without port in subscription",
			givenSinkFormat: "https://%s.%s.svc.cluster.local%s",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
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
			givenSubscription := givenSubscriptionWithoutSinkFunc(testNamespace, subName)
			givenUpdateSubscription := givenUpdateSubscriptionWithSinkFunc(testNamespace, subName, tc.givenSinkFormat, sinkPath)

			// create a subscriber service
			subscriberSvc := reconcilertesting.NewSubscriberSvc(subName, testNamespace)
			ensureK8sResourceCreated(ctx, t, subscriberSvc)

			// create subscription
			ensureK8sResourceCreated(ctx, t, givenSubscription)
			createdSubscription := givenSubscription.DeepCopy()
			// check if the created subscription is correct
			getSubscriptionAssert(ctx, g, createdSubscription).Should(wantSubscriptionWithoutSinkMatchers)

			// update subscription with valid sink
			givenUpdateSubscription.ResourceVersion = createdSubscription.ResourceVersion
			ensureK8sSubscriptionUpdated(ctx, t, givenUpdateSubscription)

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
			getSubscriptionAssert(ctx, g, givenSubscription).Should(wantUpdateSubscriptionWithSinkMatchers)

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

// Test_SinkChangeAndAPIRule tests the Subscription sink change scenario.
// The reconciler should update the EventMesh subscription webhookURL by creating a new APIRule
// when the sink is changed.
func Test_SinkChangeAndAPIRule(t *testing.T) {
	t.Parallel()

	// given
	g := gomega.NewGomegaWithT(t)
	ctx := context.Background()

	// create unique namespace for this test run
	testNamespace := getTestNamespace()
	ensureNamespaceCreated(ctx, t, testNamespace)
	subName := fmt.Sprintf("test-sink-%s", testNamespace)

	givenSubscription := reconcilertesting.NewSubscription(subName, testNamespace,
		reconcilertesting.WithDefaultSource(),
		reconcilertesting.WithOrderCreatedV1Event(),
		reconcilertesting.WithSinkURL(reconcilertesting.ValidSinkURL(testNamespace, subName)),
	)

	// phase 1: Create a Subscription with ready APIRule and ready status.
	// create a subscriber service
	subscriberSvc := reconcilertesting.NewSubscriberSvc(subName, testNamespace)
	ensureK8sResourceCreated(ctx, t, subscriberSvc)
	// create subscription
	ensureK8sResourceCreated(ctx, t, givenSubscription)
	createdSubscription := givenSubscription.DeepCopy()

	// wait until the APIRule is assigned to the created subscription
	getSubscriptionAssert(ctx, g, createdSubscription).Should(reconcilertesting.HaveNoneEmptyAPIRuleName())

	// fetch the APIRule and update the status of the APIRule to ready (mocking APIGateway controller)
	// and wait until the created Subscription becomes ready
	apiRule := &apigatewayv1beta1.APIRule{ObjectMeta: metav1.ObjectMeta{
		Name: createdSubscription.Status.Backend.APIRuleName, Namespace: createdSubscription.Namespace}}
	getAPIRuleAssert(ctx, g, apiRule).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
	ensureAPIRuleStatusUpdatedWithStatusReady(ctx, t, apiRule)

	// check if the EventMesh Subscription has the correct webhook URL
	emSub := getEventMeshSubFromMock(givenSubscription.Name, givenSubscription.Namespace)
	g.Expect(emSub).ShouldNot(gomega.BeNil())
	g.Expect(emSub).Should(eventmeshsubmatchers.HaveWebhookURL(fmt.Sprintf("https://%s/", *apiRule.Spec.Host)))

	// phase 2: Update the Subscription sink and check if new APIRule is created.
	// create a subscriber service
	subscriberSvcNew := reconcilertesting.NewSubscriberSvc(fmt.Sprintf("%s2", subName), testNamespace)
	ensureK8sResourceCreated(ctx, t, subscriberSvcNew)

	// update subscription sink
	updatedSubscription := createdSubscription.DeepCopy()
	reconcilertesting.SetSink(subscriberSvcNew.Namespace, subscriberSvcNew.Name, updatedSubscription)
	ensureK8sSubscriptionUpdated(ctx, t, updatedSubscription)
	// wait until the APIRule details are updated in Subscription.
	getSubscriptionAssert(ctx, g, updatedSubscription).Should(reconcilertesting.HaveSubscriptionReady())
	getSubscriptionAssert(ctx, g, updatedSubscription).ShouldNot(reconcilertesting.HaveAPIRuleName(apiRule.Name))

	// fetch the new APIRule and update the status of the APIRule to ready (mocking APIGateway controller)
	// and wait until the created Subscription becomes ready
	apiRule = &apigatewayv1beta1.APIRule{ObjectMeta: metav1.ObjectMeta{
		Name: updatedSubscription.Status.Backend.APIRuleName, Namespace: updatedSubscription.Namespace}}
	getAPIRuleAssert(ctx, g, apiRule).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
	ensureAPIRuleStatusUpdatedWithStatusReady(ctx, t, apiRule)

	// check if the EventMesh Subscription has the correct webhook URL
	emSub = getEventMeshSubFromMock(givenSubscription.Name, givenSubscription.Namespace)
	g.Expect(emSub).ShouldNot(gomega.BeNil())
	g.Expect(emSub).Should(eventmeshsubmatchers.HaveWebhookURL(fmt.Sprintf("https://%s/", *apiRule.Spec.Host)))
}
