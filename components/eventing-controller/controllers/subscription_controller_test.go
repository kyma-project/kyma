package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/constants"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	testingeventing "github.com/kyma-project/kyma/components/eventing-controller/testing"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/testing"
	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"

	// gcp auth etc.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// +kubebuilder:scaffold:imports
)

const (
	subscriptionNamespacePrefix = "test-"
	subscriptionID              = "test-subs-1"
	bigPollingInterval          = 5 * time.Second
	bigTimeOut                  = time.Second * 60
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

	When("Creating a valid Subscription with webhook with creation of APIRule", func() {

	})
	When("Creating a valid Subscription without webhook with already existing APIRule", func() {})
	When("Creating a valid Subscription with invalid APIRule", func() {})
	When("Subscription changed with creation of APIRule", func() {})
	When("Subscription changed in sink path with update of APIRule", func() {
		It("Should append to the array of rules in APIRule", func() {})
	})
	When("Subscription changed in sink path with update of APIRule", func() {
		It("Should shrink the array of rules in APIRule", func() {})
	})
	When("Subscription changed in sink port with update of APIRule", func() {
		It("Should add the Subscription to the ownerreferences of APIRule and remove it from the old APIRule ", func() {})
	})
	FWhen("Creating a valid Subscription(with webhook) with already existing APIRule", func() {
		It("Should reconcile the Subscription", func() {
			subscriptionName := "test-valid-subscription-1"
			ctx := context.Background()

			// Ensuring subscriber svc
			subscriberSvc := testingeventing.NewSubscriberSvc("webhook", namespaceName)
			ensureSubscriberSvcCreated(subscriberSvc, ctx)

			// Ensuring existing APIRule
			apiRule := testingeventing.NewAPIRuleWithOwnRef(testingeventing.WithoutPath, testingeventing.WithGateway, testingeventing.WithService, testingeventing.WithStatusReady)
			apiRule.Namespace = namespaceName
			apiRule.Labels = map[string]string{
				constants.ControllerServiceLabelKey:  subscriberSvc.Name,
				constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
			}
			ensureAPIRuleCreated(apiRule, ctx)

			givenSubscription := testingeventing.NewSubscription(subscriptionName, namespaceName, testingeventing.WithFilter, testingeventing.WithWebhook)
			testingeventing.WithValidSink(namespaceName, subscriberSvc.Name, givenSubscription)
			ensureSubscriptionCreated(givenSubscription, ctx)

			By("Setting a finalizer")
			//var subscription = &eventingv1alpha1.Subscription{}
			getSubscription(givenSubscription, ctx).Should(And(
				testingeventing.HaveSubscriptionName(subscriptionName),
				testingeventing.HaveSubscriptionFinalizer(SubscriptionFinalizer),
			))

			By("Setting a subscribed condition")
			subscriptionCreatedCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, v1.ConditionTrue)
			getSubscription(givenSubscription, ctx).Should(And(
				testingeventing.HaveSubscriptionName(subscriptionName),
				testingeventing.HaveCondition(subscriptionCreatedCondition),
			))

			By("Emitting a subscription created event")
			var subscriptionEvents = v1.EventList{}
			subscriptionCreatedEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionCreated),
				Message: "",
				Type:    v1.EventTypeNormal,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(testingeventing.HaveEvent(subscriptionCreatedEvent))

			By("Setting a subscription active condition")
			subscriptionActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionActive, v1.ConditionTrue)
			getSubscription(givenSubscription, ctx).Should(And(
				testingeventing.HaveSubscriptionName(subscriptionName),
				testingeventing.HaveCondition(subscriptionActiveCondition),
			))

			By("Emitting a subscription active event")
			subscriptionActiveEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionActive),
				Message: "",
				Type:    v1.EventTypeNormal,
			}
			getK8sEvents(&subscriptionEvents, givenSubscription.Namespace).Should(testingeventing.HaveEvent(subscriptionActiveEvent))

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
				}
				return false
			}).Should(BeTrue())

			By("Marking it as ready")
			getSubscription(givenSubscription, ctx).Should(testingeventing.HaveSubscriptionReady())

			By("Updating APIRule")
			getAPIRule(apiRule, ctx).Should(testingeventing.HaveValidAPIRule(givenSubscription))
		})
	})

	FWhen("Subscription changed with already existing APIRule", func() {
		It("Should update the BEB subscription", func() {
			subscriptionName := "test-subscription-beb-not-status-not-ready"
			ctx := context.Background()

			oldSvc := testingeventing.NewSubscriberSvc("webhook-old", namespaceName)
			ensureSubscriberSvcCreated(oldSvc, ctx)

			newSvc := testingeventing.NewSubscriberSvc("webhook-new", namespaceName)
			ensureSubscriberSvcCreated(newSvc, ctx)

			givenSubscription := testingeventing.NewSubscription(subscriptionName, namespaceName, testingeventing.WithFilter, testingeventing.WithWebhook)
			testingeventing.WithValidSink(oldSvc.Namespace, oldSvc.Name, givenSubscription)
			ensureSubscriptionCreated(givenSubscription, ctx)

			apiRuleForOldSvc := testingeventing.NewAPIRule(givenSubscription, testingeventing.WithoutPath, testingeventing.WithGateway, testingeventing.WithService, testingeventing.WithStatusReady)
			apiRuleForOldSvc.Namespace = namespaceName
			apiRuleForOldSvc.Labels = map[string]string{
				constants.ControllerServiceLabelKey:  oldSvc.Name,
				constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
			}
			ensureAPIRuleCreated(apiRuleForOldSvc, ctx)

			apiRuleForNewSvc := testingeventing.NewAPIRule(givenSubscription, testingeventing.WithoutPath, testingeventing.WithGateway, testingeventing.WithService, testingeventing.WithStatusReady)
			apiRuleForNewSvc.Namespace = namespaceName
			apiRuleForNewSvc.Labels = map[string]string{
				constants.ControllerServiceLabelKey:  newSvc.Name,
				constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
			}
			ensureAPIRuleCreated(apiRuleForNewSvc, ctx)

			By("Given subscription is ready")
			var subscription = &eventingv1alpha1.Subscription{}
			getSubscription(subscription, ctx).Should(testingeventing.HaveSubscriptionReady())
			//waitForSubscriptionReady(subscriptionLookupKey, ctx)

			By("Updating the sink")
			subscription.Spec.Sink = fmt.Sprintf("http://%s.%s.svc.cluster.local", newSvc.Name, newSvc.Namespace)
			updateSubscription(subscription, ctx).Should(testingeventing.HaveSubscriptionReady())

			By("Updating the BEB Subscription with the new sink")
			bebCreationRequests := make([]bebtypes.Subscription, 0)
			getBebSubscriptionCreationRequests(bebCreationRequests).Should(And(
				ContainElement(MatchFields(IgnoreMissing|IgnoreExtras,
					Fields{
						"Name": BeEquivalentTo(subscription.Name),
						//"WebhookUrl": BeEquivalentTo(subscription.Spec.Sink),
					},
				))))
		})
	})

	When("BEB subscription creation failed", func() {
		It("Should not mark the subscription as ready", func() {
			subscriptionName := "test-subscription-beb-not-status-not-ready"
			ctx := context.Background()
			svc := testingeventing.NewSubscriberSvc("webhook-old", namespaceName)
			ensureSubscriberSvcCreated(svc, ctx)
			givenSubscription := testingeventing.NewSubscription(subscriptionName, namespaceName, testingeventing.WithWebhook, testingeventing.WithFilter)
			testingeventing.WithValidSink(svc.Name, svc.Namespace, givenSubscription)
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

			By("Setting a subscription not created condition")
			subscriptionNotCreatedCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreationFailed, v1.ConditionFalse)
			getSubscription(subscription, ctx).Should(And(
				testingeventing.HaveSubscriptionName(subscriptionName),
				testingeventing.HaveCondition(subscriptionNotCreatedCondition),
			))

			By("Marking it as not ready")
			getSubscription(subscription, ctx).Should(And(
				testingeventing.HaveSubscriptionName(subscriptionName),
				Not(testingeventing.HaveSubscriptionReady()),
			))

			By("Deleting the object to not provoke more reconciliation requests")
			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
			getSubscription(subscription, ctx).ShouldNot(testingeventing.HaveSubscriptionFinalizer(SubscriptionFinalizer))
		})
	})

	When("BEB subscription status is not ready", func() {
		It("Should not mark the subscription as ready", func() {
			subscriptionName := "test-subscription-beb-not-status-not-ready"
			ctx := context.Background()
			svc := testingeventing.NewSubscriberSvc("webhook-old", namespaceName)
			ensureSubscriberSvcCreated(svc, ctx)
			givenSubscription := testingeventing.NewSubscription(subscriptionName, namespaceName, testingeventing.WithWebhook, testingeventing.WithFilter)
			testingeventing.WithValidSink(svc.Name, svc.Namespace, givenSubscription)
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

			By("Setting a subscription not active condition")
			subscriptionNotActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionNotActive, v1.ConditionFalse)
			getSubscription(subscription, ctx).Should(And(
				testingeventing.HaveSubscriptionName(subscriptionName),
				testingeventing.HaveCondition(subscriptionNotActiveCondition),
			))

			By("Marking it as not ready")
			getSubscription(subscription, ctx).Should(And(
				testingeventing.HaveSubscriptionName(subscriptionName),
				Not(testingeventing.HaveSubscriptionReady()),
			))

			By("Deleting the object to not provoke more reconciliation requests")
			Expect(k8sClient.Delete(ctx, subscription)).Should(BeNil())
			getSubscription(subscription, ctx).ShouldNot(testingeventing.HaveSubscriptionFinalizer(SubscriptionFinalizer))
		})
	})

	When("Deleting a valid Subscription", func() {
		It("Should reconcile the Subscription", func() {

			subscriptionName := "test-delete-valid-subscription-1"
			ctx := context.Background()
			givenSubscription := testingeventing.FixtureValidSubscription(subscriptionName, namespaceName, subscriptionID)
			processedBebRequests := 0
			svc := testingeventing.NewSubscriberSvc("webhook", namespaceName)
			var subscription = &eventingv1alpha1.Subscription{}
			ensureSubscriberSvcCreated(svc, ctx)
			ensureSubscriptionCreated(givenSubscription, ctx)

			Context("Given the subscription is ready", func() {

				getSubscription(subscription, ctx).Should(And(
					testingeventing.HaveSubscriptionName(subscriptionName),
					testingeventing.HaveSubscriptionReady(),
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
			getSubscription(subscription, ctx).ShouldNot(testingeventing.HaveSubscriptionFinalizer(SubscriptionFinalizer))

			By("Emitting some k8s events")
			var subscriptionEvents = v1.EventList{}
			subscriptionDeletedEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionDeleted),
				Message: "",
				Type:    v1.EventTypeWarning,
			}
			getK8sEvents(&subscriptionEvents, subscription.Namespace).Should(testingeventing.HaveEvent(subscriptionDeletedEvent))
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
				subscription := testingeventing.FixtureValidSubscription("schema-filter-missing", "", subscriptionID)
				subscription.Spec.Filter = nil
				return subscription
			}()),
		Entry("protocolsettings missing",
			func() *eventingv1alpha1.Subscription {
				subscription := testingeventing.FixtureValidSubscription("schema-filter-missing", "", subscriptionID)
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
				subscription := testingeventing.FixtureValidSubscription("schema-filter-missing", "", subscriptionID)
				subscription.Spec.ProtocolSettings.WebhookAuth = nil
				return subscription
			}()),
	)
})

// getSubscription fetches a subscription using the lookupKey and allows to make assertions on it
func getSubscription(subscription *eventingv1alpha1.Subscription, ctx context.Context) AsyncAssertion {
	return Eventually(func() eventingv1alpha1.Subscription {
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			log.Printf("failed to fetch subscription(%s): %v", lookupKey.String(), err)
			return eventingv1alpha1.Subscription{}
		}
		fmt.Println(">>>", subscription.Status.Ready)
		return *subscription
	}, bigTimeOut, bigPollingInterval)
}

// TODO change the function name
func getSubscription2(lookupKey types.NamespacedName, ctx context.Context) *eventingv1alpha1.Subscription {
	subscription := &eventingv1alpha1.Subscription{}
	if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
		return subscription
	}
	return nil
}

func updateSubscription(subscription *eventingv1alpha1.Subscription, ctx context.Context) AsyncAssertion {
	return Eventually(func() eventingv1alpha1.Subscription {
		if err := k8sClient.Update(ctx, subscription); err != nil {
			return eventingv1alpha1.Subscription{}
		}
		return *subscription
	}, time.Second*10, time.Second)
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

// ensureAPIRuleCreated creates an APIRule with status READY
func ensureAPIRuleCreated(apiRule *apigatewayv1alpha1.APIRule, ctx context.Context) {
	By(fmt.Sprintf("Ensuring the test namespace %q is created", apiRule.Namespace))
	if apiRule.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(apiRule.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			if !k8serrors.IsAlreadyExists(err) {
				fmt.Println(err)
				Expect(err).ShouldNot(HaveOccurred())
			}
		}
	}

	By(fmt.Sprintf("Ensuring the APIRule %q is created", apiRule.Name))
	// create subscription
	err := k8sClient.Create(ctx, apiRule)
	if !k8serrors.IsAlreadyExists(err) {
		fmt.Println(err)
		Expect(err).Should(BeNil())
	}
	// update status only once when APIRule creation succeeds
	if err == nil {
		err = k8sClient.Status().Update(ctx, apiRule)
		fmt.Println(err)
		Expect(err).Should(BeNil())
	}
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
			if testing.IsBebSubscriptionCreate(r, *beb.BebConfig) {
				bebSubscription := payloadObject.(bebtypes.Subscription)
				bebSubscriptions = append(bebSubscriptions, bebSubscription)
			}
		}
		return bebSubscriptions
	})
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
			testingeventing.IsK8sUnprocessableEntity(),
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
	return Eventually(func() *apigatewayv1alpha1.APIRule {
		lookUpKey := types.NamespacedName{
			Namespace: apiRule.Namespace,
			Name:      apiRule.Name,
		}
		err := k8sClient.Get(ctx, lookUpKey, apiRule)
		Expect(err).ToNot(HaveOccurred())
		return apiRule
	})
}
