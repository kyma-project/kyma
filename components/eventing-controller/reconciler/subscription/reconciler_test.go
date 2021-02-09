package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// gcp auth etc.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application/fake"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	subscriptionNamespacePrefix = "test-"
	bigPollingInterval          = 3 * time.Second
	bigTimeOut                  = 40 * time.Second
	smallTimeOut                = 5 * time.Second
	smallPollingInterval        = 1 * time.Second
	domain                      = "domain.com"
)

var (
	acceptableMethods = []string{http.MethodPost, http.MethodOptions}
)

var _ = Describe("Subscription Reconciliation Tests", func() {
	var namespaceName string

	// enable me for debugging
	// SetDefaultEventuallyTimeout(time.Minute)
	// SetDefaultEventuallyPollingInterval(time.Second)

	BeforeEach(func() {
		namespaceName = getUniqueNamespaceName()
		// we need to reset the http requests which the mock captured
		beb.Reset()
	})

	AfterEach(func() {
		// detailed request logs
		logf.Log.V(1).Info("beb requests", "number", len(beb.Requests))

		i := 0
		for req, payloadObject := range beb.Requests {
			reqDescription := fmt.Sprintf("method: %q, url: %q, payload object: %+v", req.Method, req.RequestURI, payloadObject)
			fmt.Printf("request[%d]: %s\n", i, reqDescription)
			i++
		}

		// print all subscriptions in the namespace for debugging purposes
		if err := printSubscriptions(namespaceName); err != nil {
			logf.Log.Error(err, "error while printing subscriptions")
		}
	})

	When("Creating a Subscription with invalid Sink", func() {
		It("Should set the Subscription APIRule status to not ready", func() {
			ctx := context.Background()
			subscriptionName := "sub-create-with-invalid-sink"
			// Ensuring subscriber svc
			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
			ensureSubscriberSvcCreated(subscriberSvc, ctx)

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithNotCleanEventTypeFilter, reconcilertesting.WithWebhookAuthForBEB)
			givenSubscription.Spec.Sink = "invalid"
			ensureSubscriptionCreated(givenSubscription, ctx)

			By("Setting APIRule status to False")
			subscriptionAPIReadyFalseCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionAPIRuleStatus, eventingv1alpha1.ConditionReasonAPIRuleStatusNotReady, v1.ConditionFalse, "")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(subscriptionAPIReadyFalseCondition),
			))

			By("Deleting the object to not provoke more reconciliation requests")
			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
			getSubscription(givenSubscription, ctx).ShouldNot(reconcilertesting.HaveSubscriptionFinalizer(Finalizer))

			By("Emitting a Subscription deleted event")
			var subscriptionEvents = v1.EventList{}
			subscriptionDeletedEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionDeleted),
				Message: "",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionDeletedEvent))

			By("Sending valid webhook requests only")
			validRequests, invalidRequests := countRequestsForWebhooks(reconcilertesting.EventType, reconcilertesting.EventTypeNotClean)
			Expect(validRequests).Should(BeZero())
			Expect(invalidRequests).Should(BeZero())
		})
	})

	// TODO: https://github.com/kyma-project/kyma/issues/9950
	PWhen("Fixing an invalid Subscription (related to APIRule)", func() {
		It("Should finally mark the subscription as ready", func() {})
	})

	PWhen("Two Subscriptions using the same sink but different path", func() {
		It("Should update the APIRule accordingly ", func() {})
		// check APIRule has two owner refs and two paths
		// delete one subscription
		// check APIRule has one owner ref and one path
		// delete last subscription
		// check APIRule is gone
	})

	PWhen("Creating two Subscriptions using different sinks", func() {
		It("Should mark both Subscriptions as ready and create two different APIRules", func() {})
		When("Unifying both Subscription sinks", func() {
			It("Should mark both Subscriptions as ready and end up with only one APIRule", func() {})
		})
	})

	When("Creating a valid Subscription", func() {
		It("Should mark the Subscription as ready", func() {
			subscriptionName := "test-valid-subscription-1"
			ctx := context.Background()

			// Ensuring subscriber svc
			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
			ensureSubscriberSvcCreated(subscriberSvc, ctx)

			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithNotCleanEventTypeFilter, reconcilertesting.WithWebhookAuthForBEB)
			reconcilertesting.WithValidSink(namespaceName, subscriberSvc.Name, givenSubscription)
			ensureSubscriptionCreated(givenSubscription, ctx)

			By("Creating a valid APIRule")
			getAPIRuleForASvc(subscriberSvc, ctx).Should(reconcilertesting.HaveNotEmptyAPIRule())

			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
			ensureAPIRuleStatusUpdatedWithStatusReady(&apiRuleCreated, ctx).Should(BeNil())

			By("Setting APIRule status in Subscription to Ready")
			subscriptionAPIReadyCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionAPIRuleStatus, eventingv1alpha1.ConditionReasonAPIRuleStatusReady, v1.ConditionTrue, "")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(subscriptionAPIReadyCondition),
			))

			By("Setting a finalizer")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveSubscriptionFinalizer(Finalizer),
			))

			By("Setting a subscribed condition")
			subscriptionCreatedCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, v1.ConditionTrue, "")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(subscriptionCreatedCondition),
			))

			By("Emitting a subscription created event")
			var subscriptionEvents = v1.EventList{}
			subscriptionCreatedEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionCreated),
				Message: "",
				Type:    v1.EventTypeNormal,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionCreatedEvent))

			By("Setting a subscription active condition")
			subscriptionActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionActive, v1.ConditionTrue, "")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(subscriptionActiveCondition),
			))

			By("Emitting a subscription active event")
			subscriptionActiveEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionActive),
				Message: "",
				Type:    v1.EventTypeNormal,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionActiveEvent))

			By("Creating a BEB Subscription")
			var bebSubscription bebtypes.Subscription
			Eventually(func() bool {
				for r, payloadObject := range beb.Requests {
					if reconcilertesting.IsBebSubscriptionCreate(r, *beb.BebConfig) {
						bebSubscription = payloadObject.(bebtypes.Subscription)
						receivedSubscriptionName := bebSubscription.Name
						// ensure the correct subscription was created
						return subscriptionName == receivedSubscriptionName
					}
				}
				return false
			}).Should(BeTrue())

			By("Updating APIRule")
			apiRule := &apigatewayv1alpha1.APIRule{
				ObjectMeta: metav1.ObjectMeta{
					Name:      givenSubscription.Status.APIRuleName,
					Namespace: givenSubscription.Namespace,
				},
			}
			expectedLabels := map[string]string{
				constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
				constants.ControllerServiceLabelKey:  subscriberSvc.Name,
			}
			getAPIRule(apiRule, ctx).Should(And(
				reconcilertesting.HaveAPIRuleOwnersRefs(givenSubscription.UID),
				reconcilertesting.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, "/"),
				reconcilertesting.HaveAPIRuleGateway(constants.ClusterLocalAPIGateway),
				reconcilertesting.HaveAPIRuleLabels(expectedLabels),
				reconcilertesting.HaveAPIRuleService(subscriberSvc.Name, 443, domain),
			))

			By("Marking it as ready")
			getSubscription(givenSubscription, ctx).Should(reconcilertesting.HaveSubscriptionReady())

			By("Sending valid webhook requests only")
			validRequests, invalidRequests := countRequestsForWebhooks(reconcilertesting.EventType, reconcilertesting.EventTypeNotClean)
			Expect(validRequests).Should(reconcilertesting.BeGreaterThanOrEqual(1))
			Expect(invalidRequests).Should(BeZero())
		})
	})

	When("Subscription sink name is changed", func() {
		It("Should update the BEB subscription webhookURL by creating a new APIRule", func() {
			subscriptionName := "test-subscription-sink-name-changed"
			ctx := context.Background()

			// prepare objects
			serviceOld := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithWebhookAuthForBEB, reconcilertesting.WithNotCleanEventTypeFilter)
			reconcilertesting.WithValidSink(serviceOld.Namespace, serviceOld.Name, givenSubscription)

			// create them and wait for Subscription to be ready
			readySubscription, apiRule := createSubscriptionObjectsAndWaitForReadiness(givenSubscription, serviceOld, ctx)

			By("Updating the sink")
			serviceNew := reconcilertesting.NewSubscriberSvc("webhook-new", namespaceName)
			ensureSubscriberSvcCreated(serviceNew, ctx)
			reconcilertesting.WithValidSink(serviceNew.Namespace, serviceNew.Name, readySubscription)
			updateSubscription(readySubscription, ctx).Should(reconcilertesting.HaveSubscriptionReady())
			getSubscription(readySubscription, ctx).ShouldNot(reconcilertesting.HaveAPIRuleName(apiRule.Name))

			apiRuleNew := &apigatewayv1alpha1.APIRule{ObjectMeta: metav1.ObjectMeta{Name: readySubscription.Status.APIRuleName, Namespace: namespaceName}}
			getAPIRule(apiRuleNew, ctx).Should(And(
				reconcilertesting.HaveNotEmptyHost(),
				reconcilertesting.HaveNotEmptyAPIRule(),
			))
			reconcilertesting.WithStatusReady(apiRuleNew)
			updateAPIRuleStatus(apiRuleNew, ctx).ShouldNot(HaveOccurred())

			By("BEB Subscription has the same webhook URL")
			bebCreationRequests := make([]bebtypes.Subscription, 0)
			getBebSubscriptionCreationRequests(bebCreationRequests).Should(And(
				ContainElement(MatchFields(IgnoreMissing|IgnoreExtras,
					Fields{
						"Name":       BeEquivalentTo(readySubscription.Name),
						"WebhookUrl": ContainSubstring(*apiRuleNew.Spec.Service.Host),
					},
				))))

			By("Cleanup not used APIRule")
			getAPIRule(apiRule, ctx).ShouldNot(reconcilertesting.HaveNotEmptyAPIRule())

			By("Sending valid webhook requests only")
			validRequests, invalidRequests := countRequestsForWebhooks(reconcilertesting.EventType, reconcilertesting.EventTypeNotClean)
			Expect(validRequests).Should(reconcilertesting.BeGreaterThanOrEqual(1))
			Expect(invalidRequests).Should(BeZero())
		})
	})

	When("Subscription1 sink is changed to reuse Subscription2 APIRule", func() {
		It("Should delete APIRule for Subscription1 and use APIRule2 from Subscription2 instead", func() {
			ctx := context.Background()

			// prepare objects
			// create them and wait for Subscription to be ready
			subscriptionName1 := "test-subscription-1"
			service1 := reconcilertesting.NewSubscriberSvc("webhook-1", namespaceName)
			subscription1 := reconcilertesting.NewSubscription(subscriptionName1, namespaceName, reconcilertesting.WithWebhookAuthForBEB, reconcilertesting.WithNotCleanEventTypeFilter)
			subscription1.Spec.Sink = fmt.Sprintf("https://%s.%s.svc.cluster.local/path1", service1.Name, service1.Namespace)
			readySubscription1, apiRule1 := createSubscriptionObjectsAndWaitForReadiness(subscription1, service1, ctx)
			subscriptionName2 := "test-subscription-2"
			service2 := reconcilertesting.NewSubscriberSvc("webhook-2", namespaceName)
			subscription2 := reconcilertesting.NewSubscription(subscriptionName2, namespaceName, reconcilertesting.WithWebhookAuthForBEB, reconcilertesting.WithNotCleanEventTypeFilter)
			subscription2.Spec.Sink = fmt.Sprintf("https://%s.%s.svc.cluster.local/path2", service2.Name, service2.Namespace)
			readySubscription2, apiRule2 := createSubscriptionObjectsAndWaitForReadiness(subscription2, service2, ctx)

			By("Updating the sink to use same port and service as Subscription 2")
			newSink := fmt.Sprintf("https://%s.%s.svc.cluster.local/path1", service2.Name, service2.Namespace)
			readySubscription1.Spec.Sink = newSink
			updateSubscription(readySubscription1, ctx).Should(reconcilertesting.HaveSubscriptionSink(newSink))

			By("Reusing APIRule from Subscription 2")
			getSubscription(readySubscription1, ctx).Should(reconcilertesting.HaveAPIRuleName(apiRule2.Name))

			By("Get the reused APIRule (from subscription 2)")
			apiRuleNew := &apigatewayv1alpha1.APIRule{ObjectMeta: metav1.ObjectMeta{Name: readySubscription1.Status.APIRuleName, Namespace: namespaceName}}
			getAPIRule(apiRuleNew, ctx).Should(And(
				reconcilertesting.HaveNotEmptyHost(),
				reconcilertesting.HaveNotEmptyAPIRule(),
			))

			By("Ensuring the reused APIRule has 2 OwnerReferences and 2 paths")
			getAPIRule(apiRule2, ctx).Should(And(
				reconcilertesting.HaveAPIRuleOwnersRefs(readySubscription1.UID, readySubscription2.UID),
				reconcilertesting.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, "/path1"),
				reconcilertesting.HaveAPIRuleSpecRules(acceptableMethods, object.OAuthHandlerName, "/path2"),
			))

			By("Deleting APIRule from Subscription 1")
			getAPIRule(apiRule1, ctx).ShouldNot(reconcilertesting.HaveNotEmptyAPIRule())

			By("Sending valid webhook requests only")
			validRequests, invalidRequests := countRequestsForWebhooks(reconcilertesting.EventType, reconcilertesting.EventTypeNotClean)
			Expect(validRequests).Should(reconcilertesting.BeGreaterThanOrEqual(2))
			Expect(invalidRequests).Should(BeZero())
		})
	})

	When("BEB subscription creation failed", func() {
		It("Should not mark the subscription as ready", func() {
			subscriptionName := "test-subscription-beb-not-status-not-ready"
			ctx := context.Background()

			// Ensuring subscriber svc
			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
			ensureSubscriberSvcCreated(subscriberSvc, ctx)

			// Create subscription
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithWebhookAuthForBEB, reconcilertesting.WithNotCleanEventTypeFilter)
			reconcilertesting.WithValidSink(subscriberSvc.Namespace, subscriberSvc.Name, givenSubscription)
			ensureSubscriptionCreated(givenSubscription, ctx)

			By("preparing mock to simulate creation of BEB subscription failing on BEB side")
			beb.CreateResponse = func(w http.ResponseWriter) {
				// ups ... server returns 500
				w.WriteHeader(http.StatusInternalServerError)
				s := bebtypes.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "sorry, but this mock does not let you create a BEB subscription",
				}
				err := json.NewEncoder(w).Encode(s)
				Expect(err).ShouldNot(HaveOccurred())
			}

			By("Creating a valid APIRule")
			getAPIRuleForASvc(subscriberSvc, ctx).Should(reconcilertesting.HaveNotEmptyAPIRule())

			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
			ensureAPIRuleStatusUpdatedWithStatusReady(&apiRuleCreated, ctx).Should(BeNil())

			By("Setting a subscription not created condition")
			subscriptionNotCreatedCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreationFailed, v1.ConditionFalse, "")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(subscriptionNotCreatedCondition),
			))

			By("Marking subscription as not ready")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				Not(reconcilertesting.HaveSubscriptionReady()),
			))

			By("Deleting the object to not provoke more reconciliation requests")
			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
			getSubscription(givenSubscription, ctx).ShouldNot(reconcilertesting.HaveSubscriptionFinalizer(Finalizer))

			By("Sending valid webhook requests only")
			validRequests, invalidRequests := countRequestsForWebhooks(reconcilertesting.EventType, reconcilertesting.EventTypeNotClean)
			Expect(validRequests).Should(reconcilertesting.BeGreaterThanOrEqual(1))
			Expect(invalidRequests).Should(BeZero())
		})
	})

	When("BEB subscription is set to paused after creation", func() {
		It("Should not mark the subscription as active", func() {
			subscriptionName := "test-subscription-beb-not-status-not-ready-2"
			ctx := context.Background()

			// Ensuring subscriber subscriberSvc
			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
			ensureSubscriberSvcCreated(subscriberSvc, ctx)

			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithWebhookAuthForBEB, reconcilertesting.WithNotCleanEventTypeFilter)
			reconcilertesting.WithValidSink(subscriberSvc.Namespace, subscriberSvc.Name, givenSubscription)

			isBebSubscriptionCreated := false

			By("preparing mock to simulate a non ready BEB subscription")
			beb.GetResponse = func(w http.ResponseWriter, subscriptionName string) {
				// until the BEB subscription creation call was performed, send successful get requests
				if !isBebSubscriptionCreated {
					reconcilertesting.BebGetSuccess(w, subscriptionName)
				} else {
					// after the BEB subscription was created, set the status to paused
					w.WriteHeader(http.StatusOK)
					s := bebtypes.Subscription{
						Name: subscriptionName,
						// ups ... BEB Subscription status is now paused
						SubscriptionStatus: bebtypes.SubscriptionStatusPaused,
					}
					err := json.NewEncoder(w).Encode(s)
					Expect(err).ShouldNot(HaveOccurred())
				}
			}
			beb.CreateResponse = func(w http.ResponseWriter) {
				isBebSubscriptionCreated = true
				reconcilertesting.BebCreateSuccess(w)
			}

			// Create subscription
			ensureSubscriptionCreated(givenSubscription, ctx)

			By("Creating a valid APIRule")
			getAPIRuleForASvc(subscriberSvc, ctx).Should(reconcilertesting.HaveNotEmptyAPIRule())

			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
			ensureAPIRuleStatusUpdatedWithStatusReady(&apiRuleCreated, ctx).Should(BeNil())

			By("Setting APIRule status to Ready")
			subscriptionAPIReadyCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionAPIRuleStatus, eventingv1alpha1.ConditionReasonAPIRuleStatusReady, v1.ConditionTrue, "")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(subscriptionAPIReadyCondition),
			))

			By("Setting a subscription not active condition")
			subscriptionNotActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionNotActive, v1.ConditionFalse, "")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				reconcilertesting.HaveCondition(subscriptionNotActiveCondition),
			))

			By("Marking it as not ready")
			getSubscription(givenSubscription, ctx).Should(And(
				reconcilertesting.HaveSubscriptionName(subscriptionName),
				Not(reconcilertesting.HaveSubscriptionReady()),
			))

			By("Deleting the object to not provoke more reconciliation requests")
			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())
			getSubscription(givenSubscription, ctx).ShouldNot(reconcilertesting.HaveSubscriptionFinalizer(Finalizer))

			By("Sending valid webhook requests only")
			validRequests, invalidRequests := countRequestsForWebhooks(reconcilertesting.EventType, reconcilertesting.EventTypeNotClean)
			Expect(validRequests).Should(reconcilertesting.BeGreaterThanOrEqual(1))
			Expect(invalidRequests).Should(BeZero())
		})
	})

	When("Deleting a valid Subscription", func() {
		It("Should reconcile the Subscription", func() {
			subscriptionName := "test-delete-valid-subscription-1"
			ctx := context.Background()
			givenSubscription := reconcilertesting.NewSubscription(subscriptionName, namespaceName, reconcilertesting.WithWebhookAuthForBEB, reconcilertesting.WithNotCleanEventTypeFilter)
			processedBebRequests := 0

			// Create service
			subscriberSvc := reconcilertesting.NewSubscriberSvc("webhook", namespaceName)
			ensureSubscriberSvcCreated(subscriberSvc, ctx)

			// Create subscription
			reconcilertesting.WithValidSink(subscriberSvc.Namespace, subscriberSvc.Name, givenSubscription)
			ensureSubscriptionCreated(givenSubscription, ctx)

			By("Creating a valid APIRule")
			getAPIRuleForASvc(subscriberSvc, ctx).Should(reconcilertesting.HaveNotEmptyAPIRule())

			By("Updating the APIRule(replicating apigateway controller) status to be Ready")
			apiRuleCreated := filterAPIRulesForASvc(getAPIRules(ctx, subscriberSvc), subscriberSvc)
			ensureAPIRuleStatusUpdatedWithStatusReady(&apiRuleCreated, ctx).Should(BeNil())

			Context("Given the subscription is ready", func() {
				getSubscription(givenSubscription, ctx).Should(And(
					reconcilertesting.HaveSubscriptionName(subscriptionName),
					reconcilertesting.HaveSubscriptionReady(),
				))

				By("Creating a BEB Subscription")
				var bebSubscription bebtypes.Subscription
				Eventually(func() bool {
					for r, payloadObject := range beb.Requests {
						if reconcilertesting.IsBebSubscriptionCreate(r, *beb.BebConfig) {
							bebSubscription = payloadObject.(bebtypes.Subscription)
							receivedSubscriptionName := bebSubscription.Name
							// ensure the correct subscription was created
							return subscriptionName == receivedSubscriptionName
						}
						processedBebRequests++
					}
					return false
				}).Should(BeTrue())
			})

			By("Deleting the Subscription")
			Expect(k8sClient.Delete(ctx, givenSubscription)).Should(BeNil())

			By("Removing the Subscription")
			getSubscription(givenSubscription, ctx).Should(reconcilertesting.IsAnEmptySubscription())

			By("Deleting the BEB Subscription")
			Eventually(func() bool {
				i := -1
				for r := range beb.Requests {
					i++
					// only consider requests against beb after the subscription creation request
					if i <= processedBebRequests {
						continue
					}
					if reconcilertesting.IsBebSubscriptionDelete(r) {
						receivedSubscriptionName := reconcilertesting.GetRestAPIObject(r.URL)
						// ensure the correct subscription was created
						return subscriptionName == receivedSubscriptionName
					}
					// TODO: ensure that the remaining beb calls are neither create nor delete (means no new beb subscription is created)
				}
				return false
			}).Should(BeTrue())

			By("Removing the APIRule")
			// TODO: Check for disappearing of APIRule which does not work in kubebuilder suite
			//getAPIRule(apiRule, ctx).Should(IsAnEmptyAPIRule())

			By("Emitting some k8s events")
			var subscriptionEvents = v1.EventList{}
			subscriptionDeletedEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionDeleted),
				Message: "",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(reconcilertesting.HaveEvent(subscriptionDeletedEvent))

			By("Sending valid webhook requests only")
			validRequests, invalidRequests := countRequestsForWebhooks(reconcilertesting.EventType, reconcilertesting.EventTypeNotClean)
			Expect(validRequests).Should(reconcilertesting.BeGreaterThanOrEqual(1))
			Expect(invalidRequests).Should(BeZero())
		})
	})

	DescribeTable("Schema tests: ensuring required fields are not treated as optional",
		func(subscription *eventingv1alpha1.Subscription) {
			ctx := context.Background()
			subscription.Namespace = namespaceName

			By("Letting the APIServer reject the custom resource")
			ensureSubscriptionCreationFails(subscription, ctx)
		},
		Entry("filter missing",
			func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("schema-filter-missing", "")
				subscription.Spec.Filter = nil
				return subscription
			}()),
		Entry("protocolsettings missing",
			func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("schema-protocolsettings-missing", "", reconcilertesting.WithWebhookAuthForBEB)
				subscription.Spec.ProtocolSettings = nil
				return subscription
			}()),
	)

	DescribeTable("Schema tests: ensuring optional fields are not treated as required",
		func(subscription *eventingv1alpha1.Subscription) {
			ctx := context.Background()
			namespaceName := getUniqueNamespaceName()
			subscription.Namespace = namespaceName

			By("Letting the APIServer reject the custom resource")
			ensureSubscriptionCreationFails(subscription, ctx)
		},
		Entry("protocolsettings.webhookauth missing",
			func() *eventingv1alpha1.Subscription {
				subscription := reconcilertesting.NewSubscription("schema-protocolsettings-missing", "", reconcilertesting.WithWebhookAuthForBEB)
				subscription.Spec.ProtocolSettings.WebhookAuth = nil
				return subscription
			}()),
	)
})

func updateAPIRuleStatus(apiRule *apigatewayv1alpha1.APIRule, ctx context.Context) AsyncAssertion {
	return Eventually(func() error {
		return k8sClient.Status().Update(ctx, apiRule)
	}, bigTimeOut, bigPollingInterval)
}

// getSubscription fetches a subscription using the lookupKey and allows to make assertions on it
func getSubscription(subscription *eventingv1alpha1.Subscription, ctx context.Context) AsyncAssertion {
	return Eventually(func() *eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			log.Printf("failed to fetch subscription(%s): %v", lookupKey.String(), err)
			return &eventingv1alpha1.Subscription{}
		}
		log.Printf("[Subscription] name:%s ns:%s apiRule:%s", subscription.Name, subscription.Namespace, subscription.Status.APIRuleName)
		return subscription
	}, bigTimeOut, bigPollingInterval)
}

// getK8sEvents returns all kubernetes events for the given namespace.
// The result can be used in a gomega assertion.
func getK8sEvents(eventList *v1.EventList, namespace string) AsyncAssertion {
	ctx := context.TODO()
	return Eventually(func() v1.EventList {
		err := k8sClient.List(ctx, eventList, client.InNamespace(namespace))
		if err != nil {
			return v1.EventList{}
		}
		return *eventList
	})
}

// ensureAPIRuleStatusUpdated updates the status fof the APIRule(mocking APIGateway controller)
func ensureAPIRuleStatusUpdatedWithStatusReady(apiRule *apigatewayv1alpha1.APIRule, ctx context.Context) AsyncAssertion {
	By(fmt.Sprintf("Ensuring the APIRule %q is updated", apiRule.Name))

	return Eventually(func() error {

		lookupKey := types.NamespacedName{
			Namespace: apiRule.Namespace,
			Name:      apiRule.Name,
		}
		err := k8sClient.Get(ctx, lookupKey, apiRule)
		if err != nil {
			return err
		}
		newAPIRule := apiRule.DeepCopy()
		reconcilertesting.WithStatusReady(newAPIRule)
		err = k8sClient.Status().Update(ctx, newAPIRule)
		if err != nil {
			return err
		}
		log.Printf("apirule is updated: %v", apiRule)
		return nil
	}, bigTimeOut, bigPollingInterval)
}

// ensureSubscriptionCreated creates a Subscription in the k8s cluster. If a custom namespace is used, it will be created as well.
func ensureSubscriptionCreated(subscription *eventingv1alpha1.Subscription, ctx context.Context) {
	By(fmt.Sprintf("Ensuring the test namespace %q is created", subscription.Namespace))
	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				Expect(err).ShouldNot(HaveOccurred())
			}
		}
	}

	By(fmt.Sprintf("Ensuring the subscription %q is created", subscription.Name))
	// create subscription
	err := k8sClient.Create(ctx, subscription)
	Expect(err).Should(BeNil())
}

// ensureSubscriberSvcCreated creates a Service in the k8s cluster. If a custom namespace is used, it will be created as well.
func ensureSubscriberSvcCreated(svc *corev1.Service, ctx context.Context) {

	By(fmt.Sprintf("Ensuring the test namespace %q is created", svc.Namespace))
	if svc.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(svc.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				Expect(err).ShouldNot(HaveOccurred())
			}
		}
	}

	By(fmt.Sprintf("Ensuring the subscriber service %q is created", svc.Name))
	// create subscription
	err := k8sClient.Create(ctx, svc)
	Expect(err).Should(BeNil())
}

// getBebSubscriptionCreationRequests filters the http requests made against BEB and returns the BEB Subscriptions
func getBebSubscriptionCreationRequests(bebSubscriptions []bebtypes.Subscription) AsyncAssertion {

	return Eventually(func() []bebtypes.Subscription {

		for r, payloadObject := range beb.Requests {
			if reconcilertesting.IsBebSubscriptionCreate(r, *beb.BebConfig) {
				bebSubscription := payloadObject.(bebtypes.Subscription)
				bebSubscriptions = append(bebSubscriptions, bebSubscription)
			}
		}
		return bebSubscriptions
	}, bigTimeOut, bigPollingInterval)
}

// ensureSubscriptionCreationFails creates a Subscription in the k8s cluster and ensures that it is reject because of invalid schema
func ensureSubscriptionCreationFails(subscription *eventingv1alpha1.Subscription, ctx context.Context) {
	if subscription.Namespace != "default " {
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			Expect(k8sClient.Create(ctx, namespace)).Should(BeNil())
		}
	}
	Expect(k8sClient.Create(ctx, subscription)).Should(
		And(
			// prevent nil-pointer stacktrace
			Not(BeNil()),
			reconcilertesting.IsK8sUnprocessableEntity(),
		),
	)
}

func fixtureNamespace(name string) *v1.Namespace {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

// printSubscriptions prints all subscriptions in the given namespace
func printSubscriptions(namespace string) error {
	// print subscription details
	ctx := context.TODO()
	subscriptionList := eventingv1alpha1.SubscriptionList{}
	if err := k8sClient.List(ctx, &subscriptionList, client.InNamespace(namespace)); err != nil {
		logf.Log.V(1).Info("error while getting subscription list", "error", err)
		return err
	}
	subscriptions := make([]string, 0)
	for _, sub := range subscriptionList.Items {
		subscriptions = append(subscriptions, sub.Name)
	}
	log.Printf("subscriptions: %v", subscriptions)
	return nil
}

func generateTestSuiteID() int {
	var seededRand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	return seededRand.Int()
}

func getUniqueNamespaceName() string {
	testSuiteID := generateTestSuiteID()
	namespaceName := fmt.Sprintf("%s%d", subscriptionNamespacePrefix, testSuiteID)
	return namespaceName
}

func getAPIRule(apiRule *apigatewayv1alpha1.APIRule, ctx context.Context) AsyncAssertion {
	return Eventually(func() apigatewayv1alpha1.APIRule {
		lookUpKey := types.NamespacedName{
			Namespace: apiRule.Namespace,
			Name:      apiRule.Name,
		}
		if err := k8sClient.Get(ctx, lookUpKey, apiRule); err != nil {
			log.Printf("failed to fetch APIRule(%s): %v", lookUpKey.String(), err)
			return apigatewayv1alpha1.APIRule{}
		}
		return *apiRule
	}, bigTimeOut, bigPollingInterval)
}

func filterAPIRulesForASvc(apiRules *apigatewayv1alpha1.APIRuleList, svc *corev1.Service) apigatewayv1alpha1.APIRule {
	log.Printf("apirules got ::: %v", apiRules)
	if len(apiRules.Items) == 1 && *apiRules.Items[0].Spec.Service.Name == svc.Name {
		return apiRules.Items[0]
	}
	return apigatewayv1alpha1.APIRule{}
}

func getAPIRules(ctx context.Context, svc *corev1.Service) *apigatewayv1alpha1.APIRuleList {
	labels := map[string]string{
		constants.ControllerServiceLabelKey:  svc.Name,
		constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
	}
	apiRules := &apigatewayv1alpha1.APIRuleList{}
	err := k8sClient.List(ctx, apiRules, &client.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(labels),
		Namespace:     svc.Namespace,
	})
	Expect(err).Should(BeNil())
	return apiRules
}

func getAPIRuleForASvc(svc *corev1.Service, ctx context.Context) AsyncAssertion {
	return Eventually(func() apigatewayv1alpha1.APIRule {
		apiRules := getAPIRules(ctx, svc)
		log.Printf("apirules got ::: %v", apiRules)
		return filterAPIRulesForASvc(apiRules, svc)
	}, smallTimeOut, smallPollingInterval)
}

func updateSubscription(subscription *eventingv1alpha1.Subscription, ctx context.Context) AsyncAssertion {
	return Eventually(func() *eventingv1alpha1.Subscription {
		if err := k8sClient.Update(ctx, subscription); err != nil {
			return &eventingv1alpha1.Subscription{}
		}
		return subscription
	}, time.Second*10, time.Second)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test Suite setup ////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// These tests use Ginkgo (BDD-style Go controllertesting framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// TODO: make configurable
const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var beb *reconcilertesting.BebMock

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	useExistingCluster := useExistingCluster
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../", "config", "crd", "bases"),
			filepath.Join("../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apigatewayv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	// +kubebuilder:scaffold:scheme

	bebMock := startBebMock()
	//client, err := client.New()
	// Source: https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html
	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: ":9090",
	})
	Expect(err).ToNot(HaveOccurred())
	envConf := env.Config{
		BebApiUrl:                bebMock.MessagingURL,
		ClientID:                 "foo-id",
		ClientSecret:             "foo-secret",
		TokenEndpoint:            bebMock.TokenURL,
		WebhookActivationTimeout: 0,
		WebhookClientID:          "foo-client-id",
		WebhookClientSecret:      "foo-client-secret",
		WebhookTokenEndpoint:     "foo-token-endpoint",
		Domain:                   domain,
		EventTypePrefix:          reconcilertesting.EventTypePrefix,
	}

	// prepare application-lister
	app := applicationtest.NewApplication(reconcilertesting.ApplicationNameNotClean, nil)
	applicationLister := fake.NewApplicationListerOrDie(context.Background(), app)

	err = NewReconciler(
		k8sManager.GetClient(),
		applicationLister,
		k8sManager.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("Subscription"),
		k8sManager.GetEventRecorderFor("eventing-controller"),
		envConf,
	).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// startBebMock starts the beb mock and configures the controller process to use it
func startBebMock() *reconcilertesting.BebMock {
	By("Preparing BEB Mock")
	bebConfig := &config.Config{}
	beb = reconcilertesting.NewBebMock(bebConfig)
	bebURI := beb.Start()
	logf.Log.Info("beb mock listening at", "address", bebURI)
	tokenURL := fmt.Sprintf("%s%s", bebURI, reconcilertesting.TokenURLPath)
	messagingURL := fmt.Sprintf("%s%s", bebURI, reconcilertesting.MessagingURLPath)
	beb.TokenURL = tokenURL
	beb.MessagingURL = messagingURL
	bebConfig = config.GetDefaultConfig(messagingURL)
	beb.BebConfig = bebConfig
	return beb
}

// createSubscriptionObjectsAndWaitForReadiness creates the given Subscription and the given Service. It then performs the following steps:
// - wait until an APIRule is linked in the Subscription
// - mark the APIRule as ready
// - wait until the Subscription is ready
// - as soon as both the APIRule and Subscription are ready, the function returns both objects
func createSubscriptionObjectsAndWaitForReadiness(givenSubscription *eventingv1alpha1.Subscription, service *corev1.Service, ctx context.Context) (*eventingv1alpha1.Subscription, *apigatewayv1alpha1.APIRule) {

	ensureSubscriberSvcCreated(service, ctx)
	ensureSubscriptionCreated(givenSubscription, ctx)

	By("Given subscription with none empty APIRule name")
	subscription := &eventingv1alpha1.Subscription{ObjectMeta: metav1.ObjectMeta{Name: givenSubscription.Name, Namespace: givenSubscription.Namespace}}
	// wait for APIRule to be set in Subscription
	getSubscription(subscription, ctx).Should(reconcilertesting.HaveNoneEmptyAPIRuleName())
	apiRule := &apigatewayv1alpha1.APIRule{ObjectMeta: metav1.ObjectMeta{Name: subscription.Status.APIRuleName, Namespace: subscription.Namespace}}
	getAPIRule(apiRule, ctx).Should(reconcilertesting.HaveNotEmptyAPIRule())
	reconcilertesting.WithStatusReady(apiRule)
	updateAPIRuleStatus(apiRule, ctx).ShouldNot(HaveOccurred())

	By("Given subscription is ready")
	getSubscription(subscription, ctx).Should(reconcilertesting.HaveSubscriptionReady())

	return subscription, apiRule
}

// countRequestsForWebhooks returns how many webhook requests are sent for each given event-type
func countRequestsForWebhooks(eventType1, eventType2 string) (int, int) {
	count1, count2 := 0, 0
	for _, v := range beb.Requests {
		subscription, ok := v.(bebtypes.Subscription)
		if ok && len(subscription.Events) > 0 {
			for _, event := range subscription.Events {
				if event.Type == eventType1 {
					count1++
				} else if event.Type == eventType2 {
					count2++
				}
			}
		}
	}
	return count1, count2
}
