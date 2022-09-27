package eventmesh_test

//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"log"
//	"net/http"
//	"path/filepath"
//	"strings"
//	"testing"
//	"time"
//
//	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
//
//	"github.com/avast/retry-go/v3"
//	"github.com/go-logr/zapr"
//	apigatewayv1beta1 "github.com/kyma-incubator/api-gateway/api/v1beta1"
//	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//	. "github.com/onsi/gomega/gstruct"
//	corev1 "k8s.io/api/core/v1"
//	k8serrors "k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	k8slabels "k8s.io/apimachinery/pkg/labels"
//	"k8s.io/apimachinery/pkg/types"
//	"k8s.io/client-go/kubernetes/scheme"
//	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
//	"k8s.io/client-go/rest"
//	ctrl "sigs.k8s.io/controller-runtime"
//	"sigs.k8s.io/controller-runtime/pkg/client"
//	"sigs.k8s.io/controller-runtime/pkg/envtest"
//	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
//	logf "sigs.k8s.io/controller-runtime/pkg/log"
//
//	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
//	eventmeshreconciler "github.com/kyma-project/kyma/components/eventing-controller/controllers/subscriptionv2/eventmesh"
//	"github.com/kyma-project/kyma/components/eventing-controller/logger"
//	backendbeb "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/beb"
//	backendeventmesh "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/eventmesh"
//	sink "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/sink/v2"
//	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/utils"
//	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
//	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
//	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
//	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
//	reconcilertestingv1 "github.com/kyma-project/kyma/components/eventing-controller/testing"
//	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing/v2"
//)
//
//const (
//	testEnvStartDelay           = time.Minute
//	testEnvStartAttempts        = 10
//	beforeSuiteTimeoutInSeconds = testEnvStartAttempts * 60
//	subscriptionNamespacePrefix = "test-"
//	bigPollingInterval          = 3 * time.Second
//	bigTimeOut                  = 40 * time.Second
//	smallTimeOut                = 5 * time.Second
//	smallPollingInterval        = 1 * time.Second
//	domain                      = "domain.com"
//)
//
//var (
//	acceptableMethods = []string{http.MethodPost, http.MethodOptions}
//	k8sCancelFn       context.CancelFunc
//)
//
//var _ = Describe("Subscription Reconciliation Tests", func() {
//	var namespaceName string
//	var testID = 0
//	var ctx context.Context
//
//	// enable me for debugging
//	// SetDefaultEventuallyTimeout(time.Minute)
//	// SetDefaultEventuallyPollingInterval(time.Second)
//
//	BeforeEach(func() {
//		namespaceName = fmt.Sprintf("%s%d", subscriptionNamespacePrefix, testID)
//		// we need to reset the http requests which the mock captured
//		beb.Reset()
//
//		// Context
//		ctx = context.Background()
//	})
//
//	AfterEach(func() {
//		// detailed request logs
//		logf.Log.V(1).Info("beb requests", "number", beb.Requests.Len())
//
//		i := 0
//
//		beb.Requests.ReadEach(
//			func(req *http.Request, payload interface{}) {
//				reqDescription := fmt.Sprintf("method: %q, url: %q, payload object: %+v", req.Method, req.RequestURI, payload)
//				fmt.Printf("request[%d]: %s\n", i, reqDescription)
//				i++
//			})
//
//		// print all subscriptions in the namespace for debugging purposes
//		if err := printSubscriptions(namespaceName); err != nil {
//			logf.Log.Error(err, "print subscriptions failed")
//		}
//		testID++
//	})
//
//	When("Updating the clean event types in the Subscription status", func() {
//		It("should mark the Subscription as ready", func() {
//			// create a subscriber service
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//
//			// create a Subscription
//			subscriptionName := "test-valid-subscription-1"
//			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
//				reconcilertesting.WithDefaultSource(),
//				reconcilertesting.WithEmptyTypes(),
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
//			)
//			ensureSubscriptionCreated(ctx, subscription)
//
//			// @TODO: Update this tests once validation webhook is implemented
//			// validation should reject the subscription CR if types is empty
//			//By("a Subscription without types", func() {
//			//	By("should fail to create subscription if types is empty", func() {
//			//		getSubscription(ctx, subscription).Should(And(
//			//			reconcilertesting.HaveSubscriptionName(subscriptionName),
//			//			reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
//			//				eventingv1alpha2.ConditionSubscriptionActive,
//			//				eventingv1alpha2.ConditionReasonSubscriptionActive,
//			//				corev1.ConditionTrue, "")),
//			//			reconcilertesting.HaveCleanEventTypesEmpty(),
//			//		))
//			//	})
//			//})
//
//			By("Creation of Subscription with types", func() {
//				subscribeToEventTypes := []string{
//					fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
//					fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
//				}
//
//				expectedStatusTypes := []eventingv1alpha2.EventType{
//					{
//						OriginalType: fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1EventNotClean),
//						CleanType:    fmt.Sprintf("%s0", reconcilertesting.OrderCreatedV1Event),
//					}, {
//						OriginalType: fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1EventNotClean),
//						CleanType:    fmt.Sprintf("%s1", reconcilertesting.OrderCreatedV1Event),
//					},
//				}
//
//				By("adding filters to a Subscription", func() {
//					for _, f := range subscribeToEventTypes {
//						addFilter := reconcilertesting.WithEventType(f)
//						addFilter(subscription)
//					}
//					ensureSubscriptionUpdated(ctx, subscription)
//				})
//
//				By("checking if Subscription status has 'types' with the correct values", func() {
//					getSubscription(ctx, subscription).Should(And(
//						reconcilertesting.HaveSubscriptionName(subscriptionName),
//						reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
//							eventingv1alpha2.ConditionSubscriptionActive,
//							eventingv1alpha2.ConditionReasonSubscriptionActive,
//							corev1.ConditionTrue, "")),
//						reconcilertesting.HaveCleanEventTypes(expectedStatusTypes),
//					))
//				})
//			})
//
//			By("Updating Subscription types", func() {
//				expectedStatusTypes := []eventingv1alpha2.EventType{
//					{
//						OriginalType: fmt.Sprintf("%s0alpha", reconcilertesting.OrderCreatedV1EventNotClean),
//						CleanType:    fmt.Sprintf("%s0alpha", reconcilertesting.OrderCreatedV1Event),
//					}, {
//						OriginalType: fmt.Sprintf("%s1alpha", reconcilertesting.OrderCreatedV1EventNotClean),
//						CleanType:    fmt.Sprintf("%s1alpha", reconcilertesting.OrderCreatedV1Event),
//					},
//				}
//
//				By("updating the Subscription types", func() {
//					for i, _ := range subscription.Spec.Types {
//						subscription.Spec.Types[i] = fmt.Sprintf("%salpha", subscription.Spec.Types[i])
//					}
//					ensureSubscriptionUpdated(ctx, subscription)
//				})
//
//				By("checking if Subscription status has 'types' with the correct cleaned and updated values", func() {
//					getSubscription(ctx, subscription).Should(And(
//						reconcilertesting.HaveSubscriptionName(subscriptionName),
//						reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
//							eventingv1alpha2.ConditionSubscriptionActive,
//							eventingv1alpha2.ConditionReasonSubscriptionActive,
//							corev1.ConditionTrue, "")),
//						reconcilertesting.HaveCleanEventTypes(expectedStatusTypes),
//					))
//				})
//			})
//
//			By("Deletion Subscription types", func() {
//				expectedStatusTypes := []eventingv1alpha2.EventType{
//					{
//						OriginalType: fmt.Sprintf("%s0alpha", reconcilertesting.OrderCreatedV1EventNotClean),
//						CleanType:    fmt.Sprintf("%s0alpha", reconcilertesting.OrderCreatedV1Event),
//					},
//				}
//
//				By("deleting Subscription types", func() {
//					subscription.Spec.Types = subscription.Spec.Types[:1]
//					ensureSubscriptionUpdated(ctx, subscription)
//				})
//
//				By("checking if Subscription status has 'types' with the correct cleaned values", func() {
//					getSubscription(ctx, subscription).Should(And(
//						reconcilertesting.HaveSubscriptionName(subscriptionName),
//						reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
//							eventingv1alpha2.ConditionSubscriptionActive,
//							eventingv1alpha2.ConditionReasonSubscriptionActive,
//							corev1.ConditionTrue, "")),
//						reconcilertesting.HaveCleanEventTypes(expectedStatusTypes),
//					))
//				})
//			})
//
//			// @TODO: Do we need this test? as we append the preffix inside reconciler
//			// and user wont provide prefix
//			//By("Updating the Subscription filter with an invalid prefix", func() {
//			//	By("updating the Subscription filter", func() {
//			//		subscription.Spec.Types[0] = fmt.Sprintf("invalid%s", reconcilertesting.OrderCreatedEventType)
//			//		ensureSubscriptionUpdated(ctx, subscription)
//			//	})
//			//
//			//	By("checking if Subscription status has no 'cleanEventTypes' and has ConditionSubscribed status set to false", func() {
//			//		getSubscription(ctx, subscription).Should(And(
//			//			reconcilertesting.HaveSubscriptionName(subscriptionName),
//			//			reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
//			//				eventingv1alpha2.ConditionSubscribed,
//			//				eventingv1alpha2.ConditionReasonSubscriptionCreationFailed,
//			//				corev1.ConditionFalse, "prefix not found")),
//			//			reconcilertesting.HaveCleanEventTypesEmpty(),
//			//		))
//			//	})
//			//})
//
//			// @TODO: Do we need this test? as we append the preffix inside reconciler
//			//By("Updating the invalid Subscription filter with a valid prefix", func() {
//			//	bebSubscriptionName := "test-validc1971deb47ad85e05c215e9271f111b0d3051283"
//			//	By("updating the Subscription filter", func() {
//			//		subscription.Spec.Types[0] = reconcilertesting.OrderCreatedEventType
//			//		ensureSubscriptionUpdated(ctx, subscription)
//			//	})
//			//
//			//	By("checking if Subscription status has 'cleanEventTypes' with the correct filter value", func() {
//			//		getSubscription(ctx, subscription).Should(And(
//			//			reconcilertesting.HaveSubscriptionName(subscriptionName),
//			//			reconcilertesting.HaveCondition(eventingv1alpha2.MakeCondition(
//			//				eventingv1alpha2.ConditionSubscribed,
//			//				eventingv1alpha2.ConditionReasonSubscriptionCreated,
//			//				corev1.ConditionTrue, fmt.Sprintf("BEB-subscription-name=%s", bebSubscriptionName))),
//			//			reconcilertesting.HaveCleanEventTypes([]eventingv1alpha2.EventType{
//			//				{
//			//					OriginalType: reconcilertesting.OrderCreatedEventType,
//			//					CleanType:    reconcilertesting.OrderCreatedEventType,
//			//				},
//			//			}),
//			//		))
//			//	})
//			//})
//		})
//	})
//
//	When("Creating a Subscription with invalid Sink and fixing it", func() {
//		createAndFixTheSubscription := func(sinkFormat string) {
//			// Ensuring subscriber svc
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//			// Create subscription with invalid sink
//			subscriptionName := "sub-create-with-invalid-sink"
//			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
//				reconcilertesting.WithDefaultSource(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithWebhookAuthForBEB(),
//				// The following sink is invalid because it is missing a scheme.
//				reconcilertesting.WithSinkMissingScheme(subscriberSvc.Namespace, subscriberSvc.Name),
//			)
//			ensureSubscriptionCreated(ctx, givenSubscription)
//
//			By("Setting APIRule status to False")
//			subscriptionAPIReadyFalseCondition := eventingv1alpha2.MakeCondition(
//				eventingv1alpha2.ConditionAPIRuleStatus,
//				eventingv1alpha2.ConditionReasonAPIRuleStatusNotReady,
//				corev1.ConditionFalse,
//				sink.MissingSchemeErrMsg,
//			)
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionAPIReadyFalseCondition),
//			))
//
//			By("Fixing the Subscription with a valid Sink")
//			path := "/path1"
//			validSink := fmt.Sprintf(sinkFormat, subscriberSvc.Name, subscriberSvc.Namespace, path)
//			givenSubscription.Spec.Sink = validSink
//			updateSubscription(ctx, givenSubscription).Should(reconcilertesting.HaveSubscriptionSink(validSink))
//
//			By("Creating a valid APIRule")
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//			apiRuleUpdated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
//			getAPIRule(ctx, &apiRuleUpdated).Should(And(
//				reconcilertestingv1.HaveNotEmptyHost(),
//				reconcilertestingv1.HaveNotEmptyAPIRule(),
//				reconcilertestingv1.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, path),
//				reconcilertestingv1.HaveAPIRuleOwnersRefs(givenSubscription.UID),
//			))
//
//			By("Updating the APIRule status to be Ready")
//			ensureAPIRuleStatusUpdatedWithStatusReady(ctx, &apiRuleUpdated).Should(Succeed())
//
//			By("Setting a Subscription active condition")
//			subscriptionActiveCondition := eventingv1alpha2.MakeCondition(
//				eventingv1alpha2.ConditionSubscriptionActive,
//				eventingv1alpha2.ConditionReasonSubscriptionActive,
//				corev1.ConditionTrue,
//				"",
//			)
//
//			By("Setting APIRule status in Subscription to Ready")
//			subscriptionAPIReadyCondition := eventingv1alpha2.MakeCondition(
//				eventingv1alpha2.ConditionAPIRuleStatus,
//				eventingv1alpha2.ConditionReasonAPIRuleStatusReady,
//				corev1.ConditionTrue,
//				"",
//			)
//
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionActiveCondition),
//				reconcilertesting.HaveAPIRuleName(apiRuleUpdated.Name),
//				reconcilertesting.HaveCondition(subscriptionAPIReadyCondition),
//				reconcilertesting.HaveSubscriptionReady(),
//			))
//
//			// @TODO: Fix why this check is failing
//			//By("Sending at least one creation requests for the Subscription")
//			//_, postRequests, _ = countBEBRequests(nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//			//Expect(postRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//		}
//		It("Should update the Subscription APIRule status from not ready to ready with sink containing port number", func() {
//			createAndFixTheSubscription("https://%s.%s.svc.cluster.local:8080%s")
//		})
//		It("Should update the Subscription APIRule status from not ready to ready", func() {
//			createAndFixTheSubscription("https://%s.%s.svc.cluster.local%s")
//		})
//	})
//
//	When("Creating a Subscription with empty protocol, protocolsettings and dialect", func() {
//		It("Should reconcile the Subscription", func() {
//			subscriptionName := "test-valid-subscription-1"
//
//			// Ensuring subscriber svc
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//
//			// Creating subscription with empty protocol, protocolsettings and dialect
//			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
//			)
//			ensureSubscriptionCreated(ctx, givenSubscription)
//
//			By("Creating a valid APIRule")
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
//			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
//			ensureAPIRuleStatusUpdatedWithStatusReady(ctx, &apiRuleCreated).Should(BeNil())
//
//			By("Setting APIRule status in Subscription to Ready")
//			subscriptionAPIReadyCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionAPIRuleStatus, eventingv1alpha2.ConditionReasonAPIRuleStatusReady, corev1.ConditionTrue, "")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionAPIReadyCondition),
//			))
//
//			By("Setting a finalizer")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer),
//			))
//
//			By("Setting a subscribed condition")
//			message := eventingv1alpha2.CreateMessageForConditionReasonSubscriptionCreated(nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//			subscriptionCreatedCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscribed, eventingv1alpha2.ConditionReasonSubscriptionCreated, corev1.ConditionTrue, message)
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionCreatedCondition),
//			))
//
//			By("Emitting a subscription created event")
//			var subscriptionEvents = corev1.EventList{}
//			subscriptionCreatedEvent := corev1.Event{
//				Reason:  string(eventingv1alpha2.ConditionReasonSubscriptionCreated),
//				Message: message,
//				Type:    corev1.EventTypeNormal,
//			}
//			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertestingv1.HaveEvent(subscriptionCreatedEvent))
//
//			By("Setting a subscription active condition")
//			subscriptionActiveCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscriptionActive, eventingv1alpha2.ConditionReasonSubscriptionActive, corev1.ConditionTrue, "")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionActiveCondition),
//			))
//
//			By("Emitting a subscription active event")
//			subscriptionActiveEvent := corev1.Event{
//				Reason:  string(eventingv1alpha2.ConditionReasonSubscriptionActive),
//				Message: "",
//				Type:    corev1.EventTypeNormal,
//			}
//			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertestingv1.HaveEvent(subscriptionActiveEvent))
//
//			By("Creating a BEB Subscription")
//			Eventually(wasSubscriptionCreated(givenSubscription)).Should(BeTrue())
//		})
//	})
//
//	When("Two Subscriptions using different Sinks are made to use the same Sink and then both are deleted", func() {
//		It("Should update the APIRule accordingly and then remove the APIRule", func() {
//			// Service
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//
//			// Subscriptions
//			subscription1Path := "/path1"
//			subscription1Name := "test-delete-valid-subscription-1"
//			subscription1 := reconcilertesting.NewSubscription(subscription1Name, namespaceName,
//				reconcilertesting.WithDefaultSource(),
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvcAndPath(subscriberSvc, subscription1Path),
//			)
//			ensureSubscriptionCreated(ctx, subscription1)
//
//			subscription2Path := "/path2"
//			subscription2Name := "test-delete-valid-subscription-2"
//			subscription2 := reconcilertesting.NewSubscription(subscription2Name, namespaceName,
//				reconcilertesting.WithDefaultSource(),
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvcAndPath(subscriberSvc, subscription2Path),
//			)
//			ensureSubscriptionCreated(ctx, subscription2)
//
//			By("Creating a valid APIRule")
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			By("Updating the APIRule status to be Ready")
//			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
//			ensureAPIRuleStatusUpdatedWithStatusReady(ctx, &apiRuleCreated)
//
//			By("Using the same APIRule for both Subscriptions")
//			getSubscription(ctx, subscription1).Should(reconcilertesting.HaveAPIRuleName(apiRuleCreated.Name))
//			getSubscription(ctx, subscription2).Should(reconcilertesting.HaveAPIRuleName(apiRuleCreated.Name))
//
//			By("Ensuring the APIRule has 2 OwnerReferences and 2 paths")
//			getAPIRule(ctx, &apiRuleCreated).Should(And(
//				reconcilertestingv1.HaveNotEmptyHost(),
//				reconcilertestingv1.HaveNotEmptyAPIRule(),
//				reconcilertestingv1.HaveAPIRuleOwnersRefs(subscription1.UID, subscription2.UID),
//				reconcilertestingv1.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, subscription1Path),
//				reconcilertestingv1.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, subscription2Path),
//			))
//
//			By("Deleting the first Subscription")
//			Expect(k8sClient.Delete(ctx, subscription1)).Should(BeNil())
//
//			By("Removing the Subscription")
//			getSubscription(ctx, subscription1).Should(reconcilertesting.IsAnEmptySubscription())
//
//			By("Emitting a k8s Subscription deleted event")
//			var subscriptionEvents = corev1.EventList{}
//			subscriptionDeletedEvent := corev1.Event{
//				Reason:  string(eventingv1alpha2.ConditionReasonSubscriptionDeleted),
//				Message: "",
//				Type:    corev1.EventTypeWarning,
//			}
//			getK8sEvents(&subscriptionEvents, subscription1.Namespace).Should(reconcilertestingv1.HaveEvent(subscriptionDeletedEvent))
//
//			By("Ensuring the APIRule has 1 OwnerReference and 1 path")
//			getAPIRule(ctx, &apiRuleCreated).Should(And(
//				reconcilertestingv1.HaveNotEmptyHost(),
//				reconcilertestingv1.HaveNotEmptyAPIRule(),
//				reconcilertestingv1.HaveAPIRuleOwnersRefs(subscription2.UID),
//				reconcilertestingv1.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, subscription2Path),
//			))
//
//			By("Ensuring the deleted Subscription is removed as Owner from the APIRule")
//			getAPIRule(ctx, &apiRuleCreated).ShouldNot(And(
//				reconcilertestingv1.HaveAPIRuleOwnersRefs(subscription1.UID),
//				reconcilertestingv1.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, subscription1Path),
//			))
//
//			By("Deleting the second Subscription")
//			Expect(k8sClient.Delete(ctx, subscription2)).Should(BeNil())
//
//			By("Removing the Subscription")
//			getSubscription(ctx, subscription2).Should(reconcilertesting.IsAnEmptySubscription())
//
//			By("Emitting a k8s Subscription deleted event")
//			subscriptionDeletedEvent = corev1.Event{
//				Reason:  string(eventingv1alpha2.ConditionReasonSubscriptionDeleted),
//				Message: "",
//				Type:    corev1.EventTypeWarning,
//			}
//			getK8sEvents(&subscriptionEvents, subscription2.Namespace).Should(reconcilertestingv1.HaveEvent(subscriptionDeletedEvent))
//
//			By("Removing the APIRule")
//			Expect(apiRuleCreated.GetDeletionTimestamp).NotTo(BeNil())
//
//			// @TODO: Check why this failing
//			//By("Sending at least one creation and one deletion request for each subscription")
//			//_, creationRequestsSubscription1, deletionRequestsSubscription1 := countBEBRequests(nameMapper.MapSubscriptionName(subscription1.Name, subscription1.Namespace))
//			//Expect(creationRequestsSubscription1).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//			//Expect(deletionRequestsSubscription1).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//			//
//			//_, creationRequestsSubscription2, deletionRequestsSubscription2 := countBEBRequests(nameMapper.MapSubscriptionName(subscription2.Name, subscription2.Namespace))
//			//Expect(creationRequestsSubscription2).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//			//Expect(deletionRequestsSubscription2).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//		})
//	})
//
//	When("Creating a valid Subscription", func() {
//		It("Should mark the Subscription as ready", func() {
//			subscriptionName := "test-valid-subscription-1"
//
//			// Ensuring subscriber svc
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//
//			givenSubscription := reconcilertesting.NewSubscription(
//				subscriptionName, namespaceName,
//				reconcilertesting.WithDefaultSource(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
//			)
//			ensureSubscriptionCreated(ctx, givenSubscription)
//
//			By("Creating a valid APIRule")
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
//			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
//			ensureAPIRuleStatusUpdatedWithStatusReady(ctx, &apiRuleCreated).Should(BeNil())
//
//			By("Setting APIRule status in Subscription to Ready")
//			subscriptionAPIReadyCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionAPIRuleStatus, eventingv1alpha2.ConditionReasonAPIRuleStatusReady, corev1.ConditionTrue, "")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionAPIReadyCondition),
//			))
//
//			By("Setting a finalizer")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer),
//			))
//
//			By("Setting a subscribed condition")
//			message := eventingv1alpha2.CreateMessageForConditionReasonSubscriptionCreated(nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//			subscriptionCreatedCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscribed, eventingv1alpha2.ConditionReasonSubscriptionCreated, corev1.ConditionTrue, message)
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionCreatedCondition),
//			))
//
//			By("Emitting a subscription created event")
//			var subscriptionEvents = corev1.EventList{}
//			subscriptionCreatedEvent := corev1.Event{
//				Reason:  string(eventingv1alpha2.ConditionReasonSubscriptionCreated),
//				Message: message,
//				Type:    corev1.EventTypeNormal,
//			}
//			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertestingv1.HaveEvent(subscriptionCreatedEvent))
//
//			By("Setting a subscription active condition")
//			subscriptionActiveCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscriptionActive, eventingv1alpha2.ConditionReasonSubscriptionActive, corev1.ConditionTrue, "")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionActiveCondition),
//			))
//
//			By("Emitting a subscription active event")
//			subscriptionActiveEvent := corev1.Event{
//				Reason:  string(eventingv1alpha2.ConditionReasonSubscriptionActive),
//				Message: "",
//				Type:    corev1.EventTypeNormal,
//			}
//			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertestingv1.HaveEvent(subscriptionActiveEvent))
//
//			By("Creating a BEB Subscription")
//			Eventually(wasSubscriptionCreated(givenSubscription)).Should(BeTrue())
//
//			By("Updating APIRule")
//			apiRule := &apigatewayv1beta1.APIRule{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      givenSubscription.Status.Backend.APIRuleName,
//					Namespace: givenSubscription.Namespace,
//				},
//			}
//			expectedLabels := map[string]string{
//				constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
//				constants.ControllerServiceLabelKey:  subscriberSvc.Name,
//			}
//			getAPIRule(ctx, apiRule).Should(And(
//				reconcilertestingv1.HaveAPIRuleOwnersRefs(givenSubscription.UID),
//				reconcilertestingv1.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, "/"),
//				reconcilertestingv1.HaveAPIRuleGateway(constants.ClusterLocalAPIGateway),
//				reconcilertestingv1.HaveAPIRuleLabels(expectedLabels),
//				reconcilertestingv1.HaveAPIRuleService(subscriberSvc.Name, 443, domain),
//			))
//
//			By("Marking it as ready")
//			getSubscription(ctx, givenSubscription).Should(reconcilertesting.HaveSubscriptionReady())
//
//			// @TODO: check why this is failing
//			//By("Sending at least one creation request for the Subscription")
//			//_, creationRequests, _ := countBEBRequests(nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//			//Expect(creationRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//		})
//	})
//
//	When("Subscription sink name is changed", func() {
//		It("Should update the BEB subscription webhookURL by creating a new APIRule", func() {
//			subscriptionName := "test-subscription-sink-name-changed"
//
//			// prepare objects
//			serviceOld := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
//				reconcilertesting.WithDefaultSource(),
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvc(serviceOld),
//			)
//
//			// create them and wait for Subscription to be ready
//			readySubscription, apiRule := createSubscriptionObjectsAndWaitForReadiness(ctx, givenSubscription, serviceOld)
//
//			By("Updating the sink")
//			serviceNew := reconcilertesting.NewSubscriberSvc("webhook-new", namespaceName)
//			ensureSubscriberSvcCreated(ctx, serviceNew)
//			reconcilertesting.SetSink(serviceNew.Namespace, serviceNew.Name, readySubscription)
//			updateSubscription(ctx, readySubscription).Should(reconcilertesting.HaveSubscriptionReady())
//			getSubscription(ctx, readySubscription).ShouldNot(reconcilertesting.HaveAPIRuleName(apiRule.Name))
//
//			apiRuleNew := &apigatewayv1beta1.APIRule{ObjectMeta: metav1.ObjectMeta{Name: readySubscription.Status.Backend.APIRuleName, Namespace: namespaceName}}
//			getAPIRule(ctx, apiRuleNew).Should(And(
//				reconcilertestingv1.HaveNotEmptyHost(),
//				reconcilertestingv1.HaveNotEmptyAPIRule(),
//			))
//			reconcilertesting.MarkReady(apiRuleNew)
//			updateAPIRuleStatus(ctx, apiRuleNew).ShouldNot(HaveOccurred())
//
//			By("BEB Subscription has the same webhook URL")
//			bebCreationRequests := make([]bebtypes.Subscription, 0)
//			getBEBSubscriptionCreationRequests(bebCreationRequests).Should(And(
//				ContainElement(MatchFields(IgnoreMissing|IgnoreExtras,
//					Fields{
//						"Name":       BeEquivalentTo(nameMapper.MapSubscriptionName(readySubscription.Name, readySubscription.Namespace)),
//						"WebhookURL": ContainSubstring(*apiRuleNew.Spec.Host),
//					},
//				))))
//
//			By("Cleanup not used APIRule")
//			getAPIRule(ctx, apiRule).ShouldNot(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			// @TODO: check why this is failing
//			//By("Sending at least one creation request")
//			//_, creationRequests, _ := countBEBRequests(nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//			//Expect(creationRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//		})
//	})
//
//	When("Subscription1 sink is changed to reuse Subscription2 APIRule", func() {
//		It("Should delete APIRule for Subscription1 and use APIRule2 from Subscription2 instead", func() {
//			// prepare objects
//			// create them and wait for Subscription to be ready
//			subscriptionName1 := "test-subscription-1"
//			service1 := reconcilertesting.NewSubscriberSvc("webhook-1", namespaceName)
//			subscription1 := reconcilertesting.NewSubscription(subscriptionName1, namespaceName,
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvcAndPath(service1, "/path1"),
//			)
//			readySubscription1, apiRule1 := createSubscriptionObjectsAndWaitForReadiness(ctx, subscription1, service1)
//
//			subscriptionName2 := "test-subscription-2"
//			service2 := reconcilertesting.NewSubscriberSvc("webhook-2", namespaceName)
//			subscription2 := reconcilertesting.NewSubscription(subscriptionName2, namespaceName,
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvcAndPath(service2, "/path2"),
//			)
//			readySubscription2, apiRule2 := createSubscriptionObjectsAndWaitForReadiness(ctx, subscription2, service2)
//
//			By("Updating the sink to use same port and service as Subscription 2")
//			newSink := fmt.Sprintf("https://%s.%s.svc.cluster.local/path1", service2.Name, service2.Namespace)
//			readySubscription1.Spec.Sink = newSink
//			updateSubscription(ctx, readySubscription1).Should(reconcilertesting.HaveSubscriptionSink(newSink))
//
//			By("Reusing APIRule from Subscription 2")
//			getSubscription(ctx, readySubscription1).Should(reconcilertesting.HaveAPIRuleName(apiRule2.Name))
//
//			By("Get the reused APIRule (from subscription 2)")
//			apiRuleNew := &apigatewayv1beta1.APIRule{ObjectMeta: metav1.ObjectMeta{Name: readySubscription1.Status.Backend.APIRuleName, Namespace: namespaceName}}
//			getAPIRule(ctx, apiRuleNew).Should(And(
//				reconcilertestingv1.HaveNotEmptyHost(),
//				reconcilertestingv1.HaveNotEmptyAPIRule(),
//			))
//
//			By("Ensuring the reused APIRule has 2 OwnerReferences and 2 paths")
//			getAPIRule(ctx, apiRule2).Should(And(
//				reconcilertestingv1.HaveAPIRuleOwnersRefs(readySubscription1.UID, readySubscription2.UID),
//				reconcilertestingv1.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, "/path1"),
//				reconcilertestingv1.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, "/path2"),
//			))
//
//			By("Deleting APIRule from Subscription 1")
//			getAPIRule(ctx, apiRule1).ShouldNot(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			// @TODO: check why this is failing
//			//By("Sending at least one creation request for Subscription 1")
//			//_, creationRequests, _ := countBEBRequests(nameMapper.MapSubscriptionName(subscription1.Name, subscription1.Namespace))
//			//Expect(creationRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//			//
//			//By("Sending at least one creation request for Subscription 2")
//			//_, creationRequests, _ = countBEBRequests(nameMapper.MapSubscriptionName(subscription2.Name, subscription2.Namespace))
//			//Expect(creationRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//		})
//	})
//
//	When("BEB subscription creation failed", func() {
//		It("Should not mark the subscription as ready", func() {
//			subscriptionName := "test-subscription-beb-not-status-not-ready"
//
//			// Ensuring subscriber svc
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//
//			// Create subscription
//			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
//			)
//			ensureSubscriptionCreated(ctx, givenSubscription)
//
//			By("preparing mock to simulate creation of BEB subscription failing on BEB side")
//			beb.CreateResponse = func(w http.ResponseWriter) {
//				// ups ... server returns 500
//				w.WriteHeader(http.StatusInternalServerError)
//				s := bebtypes.Response{
//					StatusCode: http.StatusInternalServerError,
//					Message:    "sorry, but this mock does not let you create a BEB subscription",
//				}
//				err := json.NewEncoder(w).Encode(s)
//				Expect(err).ShouldNot(HaveOccurred())
//			}
//
//			By("Creating a valid APIRule")
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
//			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
//			ensureAPIRuleStatusUpdatedWithStatusReady(ctx, &apiRuleCreated).Should(BeNil())
//
//			By("Setting a subscription not created condition")
//			message := "create subscription failed: 500; 500 Internal Server Error;{\"message\":\"sorry, but this mock does not let you create a BEB subscription\"}\n"
//			subscriptionNotCreatedCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscribed, eventingv1alpha2.ConditionReasonSubscriptionCreationFailed, corev1.ConditionFalse, message)
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionNotCreatedCondition),
//			))
//
//			By("Marking subscription as not ready")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				Not(reconcilertesting.HaveSubscriptionReady()),
//			))
//
//			By("Deleting the object to not provoke more reconciliation requests")
//			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
//			getSubscription(ctx, givenSubscription).ShouldNot(reconcilertesting.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer))
//
//			// @TODO: check why this is failing
//			//By("Sending at least one creation request for the Subscription")
//			//_, creationRequests, _ := countBEBRequests(nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//			//Expect(creationRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//		})
//	})
//
//	When("BEB subscription is set to paused after creation", func() {
//		It("Should not mark the subscription as active", func() {
//			subscriptionName := "test-subscription-beb-not-status-not-ready-2"
//
//			// Ensuring subscriber subscriberSvc
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//
//			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
//			)
//
//			isBEBSubscriptionCreated := false
//
//			By("preparing mock to simulate a non ready BEB subscription")
//			beb.GetResponse = func(w http.ResponseWriter, subscriptionName string) {
//				// until the BEB subscription creation call was performed, send successful get requests
//				if !isBEBSubscriptionCreated {
//					reconcilertestingv1.BEBGetSuccess(w, nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//				} else {
//					// after the BEB subscription was created, set the status to paused
//					w.WriteHeader(http.StatusOK)
//					s := bebtypes.Subscription{
//						Name: nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace),
//						// ups ... BEB Subscription status is now paused
//						SubscriptionStatus: bebtypes.SubscriptionStatusPaused,
//					}
//					err := json.NewEncoder(w).Encode(s)
//					Expect(err).ShouldNot(HaveOccurred())
//				}
//			}
//			beb.CreateResponse = func(w http.ResponseWriter) {
//				isBEBSubscriptionCreated = true
//				reconcilertestingv1.BEBCreateSuccess(w)
//			}
//
//			// Create subscription
//			ensureSubscriptionCreated(ctx, givenSubscription)
//
//			By("Creating a valid APIRule")
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
//			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
//			ensureAPIRuleStatusUpdatedWithStatusReady(ctx, &apiRuleCreated).Should(BeNil())
//
//			By("Setting APIRule status to Ready")
//			subscriptionAPIReadyCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionAPIRuleStatus, eventingv1alpha2.ConditionReasonAPIRuleStatusReady, corev1.ConditionTrue, "")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionAPIReadyCondition),
//			))
//
//			By("Setting a subscription not active condition")
//			subscriptionNotActiveCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscriptionActive,
//				eventingv1alpha2.ConditionReasonSubscriptionNotActive, corev1.ConditionFalse, "Waiting for subscription to be active")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionNotActiveCondition),
//			))
//
//			By("Marking it as not ready")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				Not(reconcilertesting.HaveSubscriptionReady()),
//			))
//
//			By("Deleting the object to not provoke more reconciliation requests")
//			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
//			getSubscription(ctx, givenSubscription).ShouldNot(reconcilertesting.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer))
//
//			// @TODO: check why this is failing
//			//By("Sending at least one creation request for the Subscription")
//			//_, creationRequests, _ := countBEBRequests(nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//			//Expect(creationRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//		})
//	})
//
//	When("BEB subscription is set to have `lastFailedDelivery` and `lastFailedDeliveryReason`='Webhook endpoint response code: 401' after creation", func() {
//		It("Should not mark the subscription as ready", func() {
//			subscriptionName := "test-subscription-beb-status-not-ready-3"
//			lastFailedDeliveryReason := "Webhook endpoint response code: 401"
//
//			// Ensuring subscriber subscriberSvc
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//
//			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
//			)
//
//			isBEBSubscriptionCreated := false
//
//			By("preparing mock to simulate a non ready BEB subscription")
//			beb.GetResponse = func(w http.ResponseWriter, subscriptionName string) {
//				// until the BEB subscription creation call was performed, send successful get requests
//				if !isBEBSubscriptionCreated {
//					reconcilertestingv1.BEBGetSuccess(w, nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//				} else {
//					// after the BEB subscription was created, set lastFailedDelivery
//					w.WriteHeader(http.StatusOK)
//					s := bebtypes.Subscription{
//						Name:                     nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace),
//						SubscriptionStatus:       bebtypes.SubscriptionStatusActive,
//						LastSuccessfulDelivery:   time.Now().Format(time.RFC3339),                       // "now",
//						LastFailedDelivery:       time.Now().Add(10 * time.Second).Format(time.RFC3339), // "now + 10s"
//						LastFailedDeliveryReason: lastFailedDeliveryReason,
//					}
//					err := json.NewEncoder(w).Encode(s)
//					Expect(err).ShouldNot(HaveOccurred())
//				}
//			}
//			beb.CreateResponse = func(w http.ResponseWriter) {
//				isBEBSubscriptionCreated = true
//				reconcilertestingv1.BEBCreateSuccess(w)
//			}
//
//			// Create subscription
//			ensureSubscriptionCreated(ctx, givenSubscription)
//
//			By("Creating a valid APIRule")
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
//			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
//			ensureAPIRuleStatusUpdatedWithStatusReady(ctx, &apiRuleCreated).Should(BeNil())
//
//			By("Setting APIRule status to Ready")
//			subscriptionAPIReadyCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionAPIRuleStatus, eventingv1alpha2.ConditionReasonAPIRuleStatusReady, corev1.ConditionTrue, "")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionAPIReadyCondition),
//			))
//
//			By("Setting a subscription active condition")
//			subscriptionActiveCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionSubscriptionActive, eventingv1alpha2.ConditionReasonSubscriptionActive, corev1.ConditionTrue, "")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionActiveCondition),
//			))
//
//			By("Setting a subscription webhook failed condition")
//			subscriptionWebhookCallFailedCondition := eventingv1alpha2.MakeCondition(eventingv1alpha2.ConditionWebhookCallStatus, eventingv1alpha2.ConditionReasonWebhookCallStatus, corev1.ConditionFalse, lastFailedDeliveryReason)
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				reconcilertesting.HaveCondition(subscriptionWebhookCallFailedCondition),
//			))
//
//			By("Marking it as not ready")
//			getSubscription(ctx, givenSubscription).Should(And(
//				reconcilertesting.HaveSubscriptionName(subscriptionName),
//				Not(reconcilertesting.HaveSubscriptionReady()),
//			))
//
//			By("Deleting the object to not provoke more reconciliation requests")
//			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
//			getSubscription(ctx, givenSubscription).ShouldNot(reconcilertesting.HaveSubscriptionFinalizer(eventingv1alpha2.Finalizer))
//
//			// @TODO: check why this is failing
//			//By("Sending at least one creation request for the Subscription")
//			//_, creationRequests, _ := countBEBRequests(nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//			//Expect(creationRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//		})
//	})
//
//	When("Deleting a valid Subscription", func() {
//		It("Should reconcile the Subscription", func() {
//
//			// Create service
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//
//			// Create subscription
//			subscriptionName := "test-delete-valid-subscription-1"
//			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
//			)
//			ensureSubscriptionCreated(ctx, givenSubscription)
//
//			By("Creating a valid APIRule")
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
//			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
//			ensureAPIRuleStatusUpdatedWithStatusReady(ctx, &apiRuleCreated).Should(BeNil())
//
//			By("Given the subscription is ready", func() {
//				getSubscription(ctx, givenSubscription).Should(And(
//					reconcilertesting.HaveSubscriptionName(subscriptionName),
//					reconcilertesting.HaveSubscriptionReady(),
//				))
//
//				By("Creating a BEB Subscription")
//				Eventually(wasSubscriptionCreated(givenSubscription)).Should(BeTrue())
//			})
//
//			By("Deleting the Subscription")
//			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
//
//			By("Removing the Subscription")
//			getSubscription(ctx, givenSubscription).Should(reconcilertesting.IsAnEmptySubscription())
//
//			By("Deleting the BEB Subscription")
//			Eventually(wasSubscriptionCreated(givenSubscription)).Should(BeTrue())
//
//			By("Removing the APIRule")
//			Expect(apiRuleCreated.GetDeletionTimestamp).NotTo(BeNil())
//
//			By("Emitting some k8s events")
//			var subscriptionEvents = corev1.EventList{}
//			subscriptionDeletedEvent := corev1.Event{
//				Reason:  string(eventingv1alpha2.ConditionReasonSubscriptionDeleted),
//				Message: "",
//				Type:    corev1.EventTypeWarning,
//			}
//			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertestingv1.HaveEvent(subscriptionDeletedEvent))
//
//			// @TODO: check why this is failing
//			//By("Sending at least one creation and one deletion request for the Subscription")
//			//_, creationRequests, deletionRequests := countBEBRequests(nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace))
//			//Expect(creationRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//			//Expect(deletionRequests).Should(reconcilertestingv1.BeGreaterThanOrEqual(1))
//		})
//	})
//
//	When("Deleting BEB Subscription manually", func() {
//		It("Should recreate BEB Subscription again", func() {
//
//			var kymaSubscription *eventingv1alpha2.Subscription
//			kymaSubscriptionName := "test-subscription"
//
//			By("Setup Kyma Subscription required resources", func() {
//				var svc *corev1.Service
//				By("Creating Subscriber service", func() {
//					svc = reconcilertesting.NewSubscriberSvc("test-service", namespaceName)
//					ensureSubscriberSvcCreated(ctx, svc)
//				})
//
//				By("Creating Kyma Subscription", func() {
//					kymaSubscription = reconcilertesting.NewSubscription(kymaSubscriptionName, namespaceName,
//						reconcilertesting.WithWebhookAuthForBEB(),
//						reconcilertesting.WithOrderCreatedFilter(),
//						reconcilertesting.WithSinkURLFromSvc(svc),
//					)
//					ensureSubscriptionCreated(ctx, kymaSubscription)
//				})
//
//				By("Creating APIRule", func() {
//					getAPIRuleForASvc(ctx, svc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//				})
//
//				By("Updating APIRule status to be ready", func() {
//					apiRule := filterAPIRulesForASvc(getAPIRules(ctx, svc), svc)
//					ensureAPIRuleStatusUpdatedWithStatusReady(ctx, &apiRule).Should(BeNil())
//				})
//			})
//
//			By("Check Kyma Subscription ready", func() {
//				By("Checking BEB mock server creation requests to contain Subscription creation request", func() {
//					Eventually(wasSubscriptionCreated(kymaSubscription)).Should(BeTrue())
//				})
//
//				By("Checking Kyma Subscription ready condition to be true", func() {
//					getSubscription(ctx, kymaSubscription).Should(And(
//						reconcilertesting.HaveSubscriptionName(kymaSubscriptionName),
//						reconcilertesting.HaveSubscriptionReady(),
//					))
//				})
//			})
//
//			By("Delete BEB Subscription", func() {
//				By("Deleting its entry in BEB mock internal cache", func() {
//					beb.Subscriptions.DeleteSubscriptionsByName(nameMapper.MapSubscriptionName(kymaSubscription.Name, kymaSubscription.Namespace))
//				})
//			})
//
//			By("Trigger Kyma Subscription reconciliation request", func() {
//				By("Labeling Kyma Subscription", func() {
//					labels := map[string]string{"reconcile": "true"}
//					kymaSubscription.Labels = labels
//					updateSubscription(ctx, kymaSubscription).Should(reconcilertesting.HaveSubscriptionLabels(labels))
//				})
//			})
//
//			// @TODO: check why this is failing
//			//By("Check BEB Subscription was recreated", func() {
//			//	By("Checking BEB mock server received a second creation request", func() {
//			//		Eventually(func() int {
//			//			_, countPost, _ := countBEBRequests(nameMapper.MapSubscriptionName(kymaSubscription.Name, kymaSubscription.Namespace))
//			//			return countPost
//			//		}, bigTimeOut, bigPollingInterval).Should(Equal(2))
//			//	})
//			//})
//		})
//	})
//
//	When("Removing APIRule of a subscription", func() {
//		It("Should recreate the APIRule", func() {
//			subscriptionName := "test-sub-apirule-recreation"
//
//			By("Creating a valid subscription")
//			// Ensuring subscriber svc
//			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
//			ensureSubscriberSvcCreated(ctx, subscriberSvc)
//
//			subscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName,
//				reconcilertesting.WithOrderCreatedFilter(),
//				reconcilertesting.WithWebhookAuthForBEB(),
//				reconcilertesting.WithSinkURLFromSvc(subscriberSvc),
//			)
//			ensureSubscriptionCreated(ctx, subscription)
//
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//
//			By("Finding and removing the matching APIRule")
//			apiRules := getAPIRules(ctx, subscriberSvc)
//			apiRule := filterAPIRulesForASvc(apiRules, subscriberSvc)
//			Expect(apiRule).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//			Expect(k8sClient.Delete(ctx, &apiRule)).ShouldNot(HaveOccurred())
//
//			// wait until it is removed
//			apiRuleKey := client.ObjectKey{
//				Namespace: apiRule.Namespace,
//				Name:      apiRule.Name,
//			}
//			Eventually(func() bool {
//				apiRule := new(apigatewayv1beta1.APIRule)
//				err := k8sClient.Get(ctx, apiRuleKey, apiRule)
//				return k8serrors.IsNotFound(err)
//			}).Should(BeTrue())
//
//			By("Ensuring a new APIRule is created")
//			getAPIRuleForASvc(ctx, subscriberSvc).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//		})
//	})
//
//	// @TODO: Update this tests once validation webhook is implemented
//	//DescribeTable("Schema tests: ensuring required fields are not treated as optional",
//	//	func(subscription *eventingv1alpha2.Subscription) {
//	//		subscription.Namespace = namespaceName
//	//
//	//		By("Letting the APIServer reject the custom resource")
//	//		ensureSubscriptionCreationFails(ctx, subscription)
//	//	},
//	//	Entry("types missing",
//	//		func() *eventingv1alpha2.Subscription {
//	//			subscription := reconcilertesting.NewSubscription("schema-filter-missing", "")
//	//			subscription.Spec.Types = nil
//	//			return subscription
//	//		}()),
//	//)
//
//	// @TODO: Update this tests once validation webhook is implemented
//	//DescribeTable("Schema tests: ensuring optional fields are not treated as required",
//	//	func(subscription *eventingv1alpha2.Subscription) {
//	//		subscription.Namespace = namespaceName
//	//
//	//		By("Letting the APIServer reject the custom resource")
//	//		ensureSubscriptionCreationFails(ctx, subscription)
//	//	},
//	//	Entry("protocolsettings.webhookauth missing",
//	//		func() *eventingv1alpha2.Subscription {
//	//			subscription := reconcilertesting.NewSubscription("schema-protocolsettings-missing", "",
//	//				reconcilertesting.WithWebhookAuthForBEB(),
//	//				reconcilertesting.WithProtocolBEB(),
//	//				reconcilertesting.WithProtocolSettings(
//	//					reconcilertesting.NewProtocolSettings(
//	//						reconcilertesting.WithBinaryContentMode(),
//	//						reconcilertesting.WithExemptHandshake(),
//	//						reconcilertesting.WithAtLeastOnceQOS()),
//	//				),
//	//			)
//	//			return subscription
//	//		}()),
//	//)
//})
//
//func updateAPIRuleStatus(ctx context.Context, apiRule *apigatewayv1beta1.APIRule) AsyncAssertion {
//	return Eventually(func() error {
//		return k8sClient.Status().Update(ctx, apiRule)
//	}, bigTimeOut, bigPollingInterval)
//}
//
//// getSubscription fetches a subscription using the lookupKey and allows making assertions on it.
//func getSubscription(ctx context.Context, subscription *eventingv1alpha2.Subscription) AsyncAssertion {
//	return Eventually(func() *eventingv1alpha2.Subscription {
//		lookupKey := types.NamespacedName{
//			Namespace: subscription.Namespace,
//			Name:      subscription.Name,
//		}
//		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
//			log.Printf("fetch subscription %s failed: %v", lookupKey.String(), err)
//			return &eventingv1alpha2.Subscription{}
//		}
//		log.Printf("[Subscription] name:%s ns:%s apiRule:%s", subscription.Name, subscription.Namespace, subscription.Status.Backend.APIRuleName)
//		return subscription
//	}, bigTimeOut, bigPollingInterval)
//}
//
//// getK8sEvents returns all kubernetes events for the given namespace.
//// The result can be used in a gomega assertion.
//func getK8sEvents(eventList *corev1.EventList, namespace string) AsyncAssertion {
//	ctx := context.TODO()
//	return Eventually(func() corev1.EventList {
//		err := k8sClient.List(ctx, eventList, client.InNamespace(namespace))
//		if err != nil {
//			return corev1.EventList{}
//		}
//		return *eventList
//	})
//}
//
//// ensureAPIRuleStatusUpdated updates the status fof the APIRule(mocking APIGateway controller).
//func ensureAPIRuleStatusUpdatedWithStatusReady(ctx context.Context, apiRule *apigatewayv1beta1.APIRule) AsyncAssertion {
//	By(fmt.Sprintf("Ensuring the APIRule %q is updated", apiRule.Name))
//
//	return Eventually(func() error {
//
//		lookupKey := types.NamespacedName{
//			Namespace: apiRule.Namespace,
//			Name:      apiRule.Name,
//		}
//		err := k8sClient.Get(ctx, lookupKey, apiRule)
//		if err != nil {
//			return err
//		}
//		newAPIRule := apiRule.DeepCopy()
//		reconcilertesting.MarkReady(newAPIRule)
//		err = k8sClient.Status().Update(ctx, newAPIRule)
//		if err != nil {
//			return err
//		}
//		return nil
//	}, bigTimeOut, bigPollingInterval)
//}
//
//// ensureSubscriptionCreated creates a Subscription in the k8s cluster. If a custom namespace is used, it will be created as well.
//func ensureSubscriptionCreated(ctx context.Context, subscription *eventingv1alpha2.Subscription) {
//	By(fmt.Sprintf("Ensuring the test namespace %q is created", subscription.Namespace))
//	if subscription.Namespace != "default " {
//		// create testing namespace
//		namespace := fixtureNamespace(subscription.Namespace)
//		err := k8sClient.Create(ctx, namespace)
//		if !k8serrors.IsAlreadyExists(err) {
//			Expect(err).ShouldNot(HaveOccurred())
//		}
//	}
//
//	By(fmt.Sprintf("Ensuring the subscription %q is created", subscription.Name))
//	Expect(k8sClient.Create(ctx, subscription)).Should(Succeed())
//}
//
//// ensureSubscriptionUpdated conducts an update of a Subscription.
//func ensureSubscriptionUpdated(ctx context.Context, subscription *eventingv1alpha2.Subscription) {
//	By(fmt.Sprintf("Ensuring the subscription %q is updated", subscription.Name))
//	Expect(k8sClient.Update(ctx, subscription)).Should(Succeed())
//}
//
//// ensureSubscriberSvcCreated creates a Service in the k8s cluster. If a custom namespace is used, it will be created as well.
//func ensureSubscriberSvcCreated(ctx context.Context, svc *corev1.Service) {
//	By(fmt.Sprintf("Ensuring the test namespace %q is created", svc.Namespace))
//	if svc.Namespace != "default " {
//		// create testing namespace
//		namespace := fixtureNamespace(svc.Namespace)
//		err := k8sClient.Create(ctx, namespace)
//		if !k8serrors.IsAlreadyExists(err) {
//			Expect(err).ShouldNot(HaveOccurred())
//		}
//	}
//
//	By(fmt.Sprintf("Ensuring the subscriber service %q is created", svc.Name))
//	Expect(k8sClient.Create(ctx, svc)).Should(Succeed())
//}
//
//// getBEBSubscriptionCreationRequests filters the http requests made against BEB and returns the BEB Subscriptions.
//func getBEBSubscriptionCreationRequests(bebSubscriptions []bebtypes.Subscription) AsyncAssertion {
//	return Eventually(func() []bebtypes.Subscription {
//		for req, sub := range beb.Requests.GetSubscriptions() {
//			if reconcilertestingv1.IsBEBSubscriptionCreate(req) {
//				bebSubscriptions = append(bebSubscriptions, sub)
//			}
//		}
//		return bebSubscriptions
//	}, bigTimeOut, bigPollingInterval)
//}
//
//// ensureSubscriptionCreationFails creates a Subscription in the k8s cluster and ensures that it is rejected because of invalid schema.
//func ensureSubscriptionCreationFails(ctx context.Context, subscription *eventingv1alpha2.Subscription) {
//	if subscription.Namespace != "default " {
//		namespace := fixtureNamespace(subscription.Namespace)
//		Expect(k8sClient.Create(ctx, namespace)).Should(Succeed())
//	}
//	Expect(k8sClient.Create(ctx, subscription)).Should(
//		And(
//			// prevent nil-pointer stacktrace
//			Not(BeNil()),
//			reconcilertestingv1.IsK8sUnprocessableEntity(),
//		),
//	)
//}
//
//func fixtureNamespace(name string) *corev1.Namespace {
//	namespace := corev1.Namespace{
//		TypeMeta: metav1.TypeMeta{
//			Kind:       "Namespace",
//			APIVersion: "v1",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name: name,
//		},
//	}
//	return &namespace
//}
//
//// printSubscriptions prints all subscriptions in the given namespace.
//func printSubscriptions(namespace string) error {
//	// print subscription details
//	ctx := context.TODO()
//	subscriptionList := eventingv1alpha2.SubscriptionList{}
//	if err := k8sClient.List(ctx, &subscriptionList, client.InNamespace(namespace)); err != nil {
//		logf.Log.V(1).Info("error while getting subscription list", "error", err)
//		return err
//	}
//	subscriptions := make([]string, 0)
//	for _, sub := range subscriptionList.Items {
//		subscriptions = append(subscriptions, sub.Name)
//	}
//	log.Printf("subscriptions: %v", subscriptions)
//	return nil
//}
//
//func getAPIRule(ctx context.Context, apiRule *apigatewayv1beta1.APIRule) AsyncAssertion {
//	return Eventually(func() apigatewayv1beta1.APIRule {
//		lookUpKey := types.NamespacedName{
//			Namespace: apiRule.Namespace,
//			Name:      apiRule.Name,
//		}
//		if err := k8sClient.Get(ctx, lookUpKey, apiRule); err != nil {
//			log.Printf("fetch APIRule %s failed: %v", lookUpKey.String(), err)
//			return apigatewayv1beta1.APIRule{}
//		}
//		return *apiRule
//	}, bigTimeOut, bigPollingInterval)
//}
//
//func filterAPIRulesForASvc(apiRules *apigatewayv1beta1.APIRuleList, svc *corev1.Service) apigatewayv1beta1.APIRule {
//	if len(apiRules.Items) == 1 && *apiRules.Items[0].Spec.Service.Name == svc.Name {
//		return apiRules.Items[0]
//	}
//	return apigatewayv1beta1.APIRule{}
//}
//
//func getAPIRules(ctx context.Context, svc *corev1.Service) *apigatewayv1beta1.APIRuleList {
//	labels := map[string]string{
//		constants.ControllerServiceLabelKey:  svc.Name,
//		constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
//	}
//	apiRules := &apigatewayv1beta1.APIRuleList{}
//	err := k8sClient.List(ctx, apiRules, &client.ListOptions{
//		LabelSelector: k8slabels.SelectorFromSet(labels),
//		Namespace:     svc.Namespace,
//	})
//	Expect(err).Should(BeNil())
//	return apiRules
//}
//
//func getAPIRuleForASvc(ctx context.Context, svc *corev1.Service) AsyncAssertion {
//	return Eventually(func() apigatewayv1beta1.APIRule {
//		apiRules := getAPIRules(ctx, svc)
//		return filterAPIRulesForASvc(apiRules, svc)
//	}, smallTimeOut, smallPollingInterval)
//}
//
//func updateSubscription(ctx context.Context, subscription *eventingv1alpha2.Subscription) AsyncAssertion {
//	return Eventually(func() *eventingv1alpha2.Subscription {
//		if err := k8sClient.Update(ctx, subscription); err != nil {
//			return &eventingv1alpha2.Subscription{}
//		}
//		return subscription
//	}, time.Second*10, time.Second)
//}
//
//// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//// Test Suite setup ////////////////////////////////////////////////////////////////////////////////////////////////////
//// //////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
//// These tests use Ginkgo (BDD-style Go controllertesting framework). Refer to
//// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.
//
//// TODO: make configurable
//// but how?
//const (
//	useExistingCluster       = false
//	attachControlPlaneOutput = false
//)
//
//var (
//	k8sClient  client.Client
//	testEnv    *envtest.Environment
//	beb        *reconcilertestingv1.BEBMock
//	nameMapper utils.NameMapper
//	mock       *reconcilertestingv1.BEBMock
//)
//
//func TestAPIs(t *testing.T) {
//	RegisterFailHandler(Fail)
//
//	RunSpecsWithDefaultAndCustomReporters(t,
//		"Controller Suite",
//		[]Reporter{printer.NewlineReporter{}})
//}
//
//var _ = BeforeSuite(func(done Done) {
//	By("bootstrapping test environment")
//	useExistingCluster := useExistingCluster
//	testEnv = &envtest.Environment{
//		CRDDirectoryPaths: []string{
//			filepath.Join("../../../", "config", "crd", "bases"),
//			filepath.Join("../../../", "config", "crd", "basesv1alpha2"),
//			filepath.Join("../../../", "config", "crd", "external"),
//			filepath.Join("../../../", "config", "crd", "patches"), // patches/cainjection_in_subscriptions.yaml
//		},
//		AttachControlPlaneOutput: attachControlPlaneOutput,
//		UseExistingCluster:       &useExistingCluster,
//	}
//
//	var err error
//	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
//	Expect(err).To(BeNil())
//	logf.SetLogger(zapr.NewLogger(defaultLogger.WithContext().Desugar()))
//
//	var cfg *rest.Config
//	err = retry.Do(func() error {
//		defer func() {
//			if r := recover(); r != nil {
//				log.Println("panic recovered:", r)
//			}
//		}()
//
//		cfg, err = testEnv.Start()
//		return err
//	},
//		retry.Delay(testEnvStartDelay),
//		retry.DelayType(retry.FixedDelay),
//		retry.Attempts(testEnvStartAttempts),
//		retry.OnRetry(func(n uint, err error) {
//			log.Printf("[%v] try failed to start testenv: %s", n, err)
//			if stopErr := testEnv.Stop(); stopErr != nil {
//				log.Printf("failed to stop testenv: %s", stopErr)
//			}
//		}),
//	)
//
//	Expect(err).ToNot(HaveOccurred())
//	Expect(cfg).ToNot(BeNil())
//
//	err = eventingv1alpha2.AddToScheme(scheme.Scheme)
//	Expect(err).NotTo(HaveOccurred())
//
//	err = apigatewayv1beta1.AddToScheme(scheme.Scheme)
//	Expect(err).NotTo(HaveOccurred())
//	// +kubebuilder:scaffold:scheme
//
//	mock = startBEBMock()
//	// client, err := client.New()
//	// Source: https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html
//	syncPeriod := time.Second * 2
//	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
//		Scheme:             scheme.Scheme,
//		SyncPeriod:         &syncPeriod,
//		MetricsBindAddress: "localhost:9090",
//	})
//	Expect(err).ToNot(HaveOccurred())
//	envConf := env.Config{
//		BEBAPIURL:                mock.MessagingURL,
//		ClientID:                 "foo-id",
//		ClientSecret:             "foo-secret",
//		TokenEndpoint:            mock.TokenURL,
//		WebhookActivationTimeout: 0,
//		WebhookTokenEndpoint:     "foo-token-endpoint",
//		Domain:                   domain,
//		EventTypePrefix:          reconcilertesting.EventTypePrefix,
//		BEBNamespace:             "/default/ns",
//		Qos:                      string(bebtypes.QosAtLeastOnce),
//		EnableNewCRDVersion: true,
//	}
//
//	credentials := &backendbeb.OAuth2ClientCredentials{
//		ClientID:     "foo-client-id",
//		ClientSecret: "foo-client-secret",
//	}
//
//	// prepare
//	eventMeshCleaner := cleaner.NewEventMeshCleaner(defaultLogger)
//	nameMapper = utils.NewBEBSubscriptionNameMapper(domain, backendeventmesh.MaxEventMeshSubscriptionNameLength)
//	eventMeshHandler := backendeventmesh.NewEventMesh(credentials, nameMapper, defaultLogger)
//
//	recorder := k8sManager.GetEventRecorderFor("eventing-controller")
//	sinkValidator := sink.NewValidator(context.Background(), k8sManager.GetClient(), recorder, defaultLogger)
//	err = eventmeshreconciler.NewReconciler(context.Background(), k8sManager.GetClient(), defaultLogger,
//		recorder, envConf, eventMeshCleaner, eventMeshHandler, credentials, nameMapper, sinkValidator).SetupUnmanaged(k8sManager)
//	Expect(err).ToNot(HaveOccurred())
//
//	go func() {
//		defer GinkgoRecover()
//		var ctx context.Context
//		ctx, k8sCancelFn = context.WithCancel(ctrl.SetupSignalHandler())
//		err = k8sManager.Start(ctx)
//		Expect(err).ToNot(HaveOccurred())
//	}()
//
//	k8sClient = k8sManager.GetClient()
//	Expect(k8sClient).ToNot(BeNil())
//
//	close(done)
//}, beforeSuiteTimeoutInSeconds)
//
//var _ = AfterSuite(func() {
//	By("tearing down the test environment")
//	if k8sCancelFn != nil {
//		k8sCancelFn()
//	}
//	err := testEnv.Stop()
//	mock.Stop()
//	Expect(err).ToNot(HaveOccurred())
//})
//
//// startBEBMock starts the beb mock and configures the controller process to use it.
//func startBEBMock() *reconcilertestingv1.BEBMock {
//	By("Preparing BEB Mock")
//	beb = reconcilertestingv1.NewBEBMock()
//	beb.Start()
//	return beb
//}
//
//// createSubscriptionObjectsAndWaitForReadiness creates the given Subscription and the given Service. It then performs the following steps:
//// - wait until an APIRule is linked in the Subscription
//// - mark the APIRule as ready
//// - wait until the Subscription is ready
//// - as soon as both the APIRule and Subscription are ready, the function returns both objects.
//func createSubscriptionObjectsAndWaitForReadiness(ctx context.Context, givenSubscription *eventingv1alpha2.Subscription, service *corev1.Service) (*eventingv1alpha2.Subscription, *apigatewayv1beta1.APIRule) {
//	ensureSubscriberSvcCreated(ctx, service)
//	ensureSubscriptionCreated(ctx, givenSubscription)
//
//	By("Given subscription with none empty APIRule name")
//	sub := &eventingv1alpha2.Subscription{ObjectMeta: metav1.ObjectMeta{Name: givenSubscription.Name, Namespace: givenSubscription.Namespace}}
//	// wait for APIRule to be set in Subscription
//	getSubscription(ctx, sub).Should(reconcilertesting.HaveNoneEmptyAPIRuleName())
//	apiRule := &apigatewayv1beta1.APIRule{ObjectMeta: metav1.ObjectMeta{Name: sub.Status.Backend.APIRuleName, Namespace: sub.Namespace}}
//	getAPIRule(ctx, apiRule).Should(reconcilertestingv1.HaveNotEmptyAPIRule())
//	reconcilertesting.MarkReady(apiRule)
//	updateAPIRuleStatus(ctx, apiRule).ShouldNot(HaveOccurred())
//
//	By("Given subscription is ready")
//	getSubscription(ctx, sub).Should(reconcilertesting.HaveSubscriptionReady())
//
//	return sub, apiRule
//}
//
//// countBEBRequests returns how many requests for a given subscription are sent for each HTTP method
////
////nolint:unparam
//func countBEBRequests(subscriptionName string) (countGet, countPost, countDelete int) {
//	countGet, countPost, countDelete = 0, 0, 0
//	beb.Requests.ReadEach(
//		func(request *http.Request, payload interface{}) {
//			switch method := request.Method; method {
//			case http.MethodGet:
//				if strings.Contains(request.URL.Path, subscriptionName) {
//					countGet++
//				}
//			case http.MethodPost:
//				if sub, ok := payload.(bebtypes.Subscription); ok {
//					if len(sub.Events) > 0 {
//						for _, event := range sub.Events {
//							if event.Type == reconcilertesting.OrderCreatedEventType && sub.Name == subscriptionName {
//								countPost++
//							}
//						}
//					}
//				}
//			case http.MethodDelete:
//				if strings.Contains(request.URL.Path, subscriptionName) {
//					countDelete++
//				}
//			}
//		})
//	return countGet, countPost, countDelete
//}
//
//// wasSubscriptionCreated returns a func that determines if a given Subscription was actually created.
//func wasSubscriptionCreated(givenSubscription *eventingv1alpha2.Subscription) func() bool {
//	return func() bool {
//		for request, name := range beb.Requests.GetSubscriptionNames() {
//			if reconcilertestingv1.IsBEBSubscriptionCreate(request) {
//				return nameMapper.MapSubscriptionName(givenSubscription.Name, givenSubscription.Namespace) == name
//			}
//		}
//		return false
//	}
//}
