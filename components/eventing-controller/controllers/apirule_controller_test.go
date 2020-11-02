package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/testing"
	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"

	// gcp auth etc.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("APIRule Reconciliation Tests", func() {
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

	When("Subscription status is OK and APIRule status is also OK", func() {
		It("Should set APIRule status in Subscription to OK", func() {
			//subscriptionName := "test-valid-subscription-1"
			//ctx := context.Background()
			//givenSubscription := fixtureValidSubscription(subscriptionName, namespaceName)
			//ensureSubscriptionCreated(givenSubscription, ctx)
			//subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}
			//
			//By("Setting a finalizer")
			//var subscription = &eventingv1alpha1.Subscription{}
			//getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
			//	testing.HaveSubscriptionName(subscriptionName),
			//	testing.HaveSubscriptionFinalizer(SubscriptionFinalizer),
			//))
			//
			//By("Setting a subscribed condition")
			//subscriptionCreatedCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, v1.ConditionTrue)
			//getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
			//	testing.HaveSubscriptionName(subscriptionName),
			//	testing.HaveCondition(subscriptionCreatedCondition),
			//))
			//
			//By("Emitting a subscription created event")
			//var subscriptionEvents = v1.EventList{}
			//subscriptionCreatedEvent := v1.Event{
			//	Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionCreated),
			//	Message: "",
			//	Type:    v1.EventTypeNormal,
			//}
			//getK8sEvents(&subscriptionEvents, subscription.Namespace).Should(testing.HaveEvent(subscriptionCreatedEvent))
			//
			//By("Setting a subscription active condition")
			//subscriptionActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionActive, v1.ConditionTrue)
			//getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
			//	testing.HaveSubscriptionName(subscriptionName),
			//	testing.HaveCondition(subscriptionActiveCondition),
			//))
			//
			//By("Emitting a subscription active event")
			//subscriptionActiveEvent := v1.Event{
			//	Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionActive),
			//	Message: "",
			//	Type:    v1.EventTypeNormal,
			//}
			//getK8sEvents(&subscriptionEvents, subscription.Namespace).Should(testing.HaveEvent(subscriptionActiveEvent))
			//
			//By("Creating a BEB Subscription")
			//var bebSubscription bebtypes.Subscription
			//Eventually(func() bool {
			//	for r, payloadObject := range beb.Requests {
			//		if testing.IsBebSubscriptionCreate(r, *beb.BebConfig) {
			//			bebSubscription = payloadObject.(bebtypes.Subscription)
			//			receivedSubscriptionName := bebSubscription.Name
			//			// ensure the correct subscription was created
			//			return subscriptionName == receivedSubscriptionName
			//		}
			//	}
			//	return false
			//}).Should(BeTrue())
			//
			//By("Marking it as ready")
			//getSubscription(subscription, subscriptionLookupKey, ctx).Should(testing.HaveSubscriptionReady())
		})
	})

	When("Subscription status is OK and APIRule status is not OK", func() {
		It("Should set APIRule status in Subscription to false", func() {
			//	subscriptionName := "test-subscription-beb-not-status-not-ready"
			//	ctx := context.Background()
			//
			//	oldSvc := NewSubscriberSvc("webhook-old", namespaceName)
			//	ensureSubscriberSvcCreated(oldSvc, ctx)
			//
			//	newSvc := NewSubscriberSvc("webhook-new", namespaceName)
			//	ensureSubscriberSvcCreated(newSvc, ctx)
			//
			//	apiRuleForOldSvc := handlers.NewAPIRule(handlers.WithoutPath, handlers.WithGateway, handlers.WithService, handlers.WithStatusReady)
			//	apiRuleForOldSvc.Namespace = namespaceName
			//	apiRuleForOldSvc.Labels = map[string]string{
			//		ControllerServiceLabelKey:  oldSvc.Name,
			//		ControllerIdentityLabelKey: ControllerIdentityLabelValue,
			//	}
			//	ensureAPIRuleCreated(apiRuleForOldSvc, ctx)
			//
			//	apiRuleForNewSvc := handlers.NewAPIRule(handlers.WithoutPath, handlers.WithGateway, handlers.WithService, handlers.WithStatusReady)
			//	apiRuleForNewSvc.Namespace = namespaceName
			//	apiRuleForNewSvc.Labels = map[string]string{
			//		ControllerServiceLabelKey:  newSvc.Name,
			//		ControllerIdentityLabelKey: ControllerIdentityLabelValue,
			//	}
			//	ensureAPIRuleCreated(apiRuleForNewSvc, ctx)
			//
			//	givenSubscription := NewSubscription(subscriptionName, namespaceName, WithFilter, WithWebhook)
			//	WithValidSink(oldSvc.Namespace, oldSvc.Name, givenSubscription)
			//
			//	ensureSubscriptionCreated(givenSubscription, ctx)
			//	subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}
			//
			//	By("Given subscription is ready")
			//	var subscription = &eventingv1alpha1.Subscription{}
			//	getSubscription(subscription, subscriptionLookupKey, ctx).Should(testing.HaveSubscriptionReady())
			//
			//	By("Updating the sink")
			//	subscription.Spec.Sink = fmt.Sprintf("http://%s.%s.svc.cluster.local", newSvc.Name, newSvc.Namespace)
			//	updateSubscription(subscription, ctx).Should(testing.HaveSubscriptionReady())
			//
			//	By("Updating the BEB Subscription with the new sink")
			//	bebCreationRequests := make([]bebtypes.Subscription, 0)
			//	getBebSubscriptionCreationRequests(bebCreationRequests).Should(And(
			//		ContainElement(MatchFields(IgnoreMissing|IgnoreExtras,
			//			Fields{
			//				"Name":       BeEquivalentTo(subscription.Name),
			//				"WebhookUrl": BeEquivalentTo(subscription.Spec.Sink),
			//			},
			//		))))
		})
	})

	When("BEB subscription creation failed", func() {
		It("Should not mark the subscription as ready", func() {
			subscriptionName := "test-subscription-beb-not-status-not-ready"
			ctx := context.Background()
			svc := NewSubscriberSvc("webhook-old", namespaceName)
			ensureSubscriberSvcCreated(svc, ctx)
			givenSubscription := NewSubscription(subscriptionName, namespaceName, WithWebhook, WithFilter)
			WithValidSink(svc.Name, svc.Namespace, givenSubscription)
			var subscription = &eventingv1alpha1.Subscription{}

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

			ensureSubscriptionCreated(givenSubscription, ctx)
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}

			By("Setting a subscription not created condition")
			subscriptionNotCreatedCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreationFailed, v1.ConditionFalse)
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
				testing.HaveSubscriptionName(subscriptionName),
				testing.HaveCondition(subscriptionNotCreatedCondition),
			))

			By("Marking it as not ready")
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
				testing.HaveSubscriptionName(subscriptionName),
				Not(testing.HaveSubscriptionReady()),
			))

			By("Deleting the object to not provoke more reconciliation requests")
			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
			getSubscription(subscription, subscriptionLookupKey, ctx).ShouldNot(testing.HaveSubscriptionFinalizer(SubscriptionFinalizer))
		})
	})

	When("BEB subscription status is not ready", func() {
		It("Should not mark the subscription as ready", func() {
			subscriptionName := "test-subscription-beb-not-status-not-ready"
			ctx := context.Background()
			svc := NewSubscriberSvc("webhook-old", namespaceName)
			ensureSubscriberSvcCreated(svc, ctx)
			givenSubscription := NewSubscription(subscriptionName, namespaceName, WithWebhook, WithFilter)
			WithValidSink(svc.Name, svc.Namespace, givenSubscription)
			var subscription = &eventingv1alpha1.Subscription{}
			isBebSubscriptionCreated := false

			By("preparing mock to simulate a non ready BEB subscription")
			beb.GetResponse = func(w http.ResponseWriter, subscriptionName string) {
				// until the BEB subscription creation call was performed, send successful get requests
				if !isBebSubscriptionCreated {
					testing.BebGetSuccess(w, subscriptionName)
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
				testing.BebCreateSuccess(w)
			}

			ensureSubscriptionCreated(givenSubscription, ctx)
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}

			By("Setting a subscription not active condition")
			subscriptionNotActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionNotActive, v1.ConditionFalse)
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
				testing.HaveSubscriptionName(subscriptionName),
				testing.HaveCondition(subscriptionNotActiveCondition),
			))

			By("Marking it as not ready")
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
				testing.HaveSubscriptionName(subscriptionName),
				Not(testing.HaveSubscriptionReady()),
			))

			By("Deleting the object to not provoke more reconciliation requests")
			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
			getSubscription(subscription, subscriptionLookupKey, ctx).ShouldNot(testing.HaveSubscriptionFinalizer(SubscriptionFinalizer))
		})
	})

	When("Deleting a valid Subscription", func() {
		It("Should reconcile the Subscription", func() {

			subscriptionName := "test-delete-valid-subscription-1"
			ctx := context.Background()
			givenSubscription := fixtureValidSubscription(subscriptionName, namespaceName)
			processedBebRequests := 0
			svc := NewSubscriberSvc("webhook", namespaceName)
			var subscription = &eventingv1alpha1.Subscription{}
			ensureSubscriberSvcCreated(svc, ctx)
			ensureSubscriptionCreated(givenSubscription, ctx)
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}

			Context("Given the subscription is ready", func() {

				getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
					testing.HaveSubscriptionName(subscriptionName),
					testing.HaveSubscriptionReady(),
				))

				By("Creating a BEB Subscription")
				var bebSubscription bebtypes.Subscription
				Eventually(func() bool {
					for r, payloadObject := range beb.Requests {
						if testing.IsBebSubscriptionCreate(r, *beb.BebConfig) {
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
			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())

			By("Deleting the BEB Subscription")
			Eventually(func() bool {
				i := -1
				for r := range beb.Requests {
					i++
					// only consider requests against beb after the subscription creation request
					if i <= processedBebRequests {
						continue
					}
					if testing.IsBebSubscriptionDelete(r) {
						receivedSubscriptionName := testing.GetRestAPIObject(r.URL)
						// ensure the correct subscription was created
						return subscriptionName == receivedSubscriptionName
					}
				}
				return false
			}).Should(BeTrue())

			By("Removing the finalizer")
			getSubscription(subscription, subscriptionLookupKey, ctx).ShouldNot(testing.HaveSubscriptionFinalizer(SubscriptionFinalizer))

			By("Emitting some k8s events")
			var subscriptionEvents = v1.EventList{}
			subscriptionDeletedEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionDeleted),
				Message: "",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, subscription.Namespace).Should(testing.HaveEvent(subscriptionDeletedEvent))
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
				subscription := fixtureValidSubscription("schema-filter-missing", "")
				subscription.Spec.Filter = nil
				return subscription
			}()),
		Entry("protocolsettings missing",
			func() *eventingv1alpha1.Subscription {
				subscription := fixtureValidSubscription("schema-filter-missing", "")
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
			ensureSubscriptionCreated(subscription, ctx)
		},
		Entry("protocolsettings.webhookauth missing",
			func() *eventingv1alpha1.Subscription {
				subscription := fixtureValidSubscription("schema-filter-missing", "")
				subscription.Spec.ProtocolSettings.WebhookAuth = nil
				return subscription
			}()),
	)
})
