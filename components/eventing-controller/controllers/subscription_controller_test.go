package controllers

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/testing"
	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/types"

	// gcp auth etc.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// +kubebuilder:scaffold:imports
)

const (
	subscriptionNamespacePrefix = "test-"
	subscriptionID              = "test-subs-1"
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

	When("Creating a valid Subscription", func() {
		It("Should reconcile the Subscription", func() {
			subscriptionName := "test-valid-subscription-1"
			ctx := context.Background()
			givenSubscription := fixtureValidSubscription(subscriptionName, namespaceName)
			ensureSubscriptionCreated(givenSubscription, ctx)
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}

			By("Setting a finalizer")
			var subscription = &eventingv1alpha1.Subscription{}
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
				testing.HaveSubscriptionName(subscriptionName),
				testing.HaveSubscriptionFinalizer(FinalizerName),
			))

			By("Setting a subscribed condition")
			subscriptionCreatedCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, eventingv1alpha1.ConditionReasonSubscriptionCreated, v1.ConditionTrue)
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
				testing.HaveSubscriptionName(subscriptionName),
				testing.HaveCondition(subscriptionCreatedCondition),
			))

			By("Emitting a subscription created event")
			var subscriptionEvents = v1.EventList{}
			subscriptionCreatedEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionCreated),
				Message: "",
				Type:    v1.EventTypeNormal,
			}
			getK8sEvents(&subscriptionEvents, subscription.Namespace).Should(testing.HaveEvent(subscriptionCreatedEvent))

			By("Setting a subscription active condition")
			subscriptionActiveCondition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscriptionActive, eventingv1alpha1.ConditionReasonSubscriptionActive, v1.ConditionTrue)
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
				testing.HaveSubscriptionName(subscriptionName),
				testing.HaveCondition(subscriptionActiveCondition),
			))

			By("Emitting a subscription active event")
			subscriptionActiveEvent := v1.Event{
				Reason:  string(eventingv1alpha1.ConditionReasonSubscriptionActive),
				Message: "",
				Type:    v1.EventTypeNormal,
			}
			getK8sEvents(&subscriptionEvents, subscription.Namespace).Should(testing.HaveEvent(subscriptionActiveEvent))

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
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(testing.HaveSubscriptionReady())
		})
	})

	When("Subscription hash did change", func() {
		It("Should update the BEB subscription", func() {
		})
	})

	When("Subscription hash did not change", func() {
		It("Should not create a POST or DELETE call against BEB", func() {
		})
	})

	When("BEB Subscription is not active", func() {
		It("Should not mark the subscription as ready", func() {
		})
	})

	When("BEB subscription creation failed", func() {
		It("Should not mark the subscription as ready", func() {
			By("Marking it as not ready")
			By("Setting a subscription not created condition")
		})
	})

	When("BEB subscription status is not ready", func() {
		It("Should not mark the subscription as ready", func() {
			By("Marking it as not ready")
			By("Setting a subscription not active condition")
		})
	})

	When("Deleting a valid Subscription", func() {
		It("Should reconcile the Subscription", func() {

			subscriptionName := "test-delete-valid-subscription-1"
			ctx := context.Background()
			givenSubscription := fixtureValidSubscription(subscriptionName, namespaceName)
			processedBebRequests := 0
			var subscription = &eventingv1alpha1.Subscription{}
			ensureSubscriptionCreated(givenSubscription, ctx)
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}

			Context("Given the subscription is ready", func() {

				getSubscription(subscription, subscriptionLookupKey, ctx).Should(And(
					testing.HaveSubscriptionName(subscriptionName),
					testing.HaveSubscriptionFinalizer(FinalizerName),
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
			getSubscription(subscription, subscriptionLookupKey, ctx).ShouldNot(testing.HaveSubscriptionFinalizer(FinalizerName))

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

// getSubscription fetches a subscription using the lookupKey and allows to make assertions on it
func getSubscription(subscription *eventingv1alpha1.Subscription, lookupKey types.NamespacedName, ctx context.Context) AsyncAssertion {
	return Eventually(func() eventingv1alpha1.Subscription {
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
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

// ensureSubscriptionCreated creates a Subscription in the k8s cluster. If a custom namespace is used, it will be created as well.
func ensureSubscriptionCreated(subscription *eventingv1alpha1.Subscription, ctx context.Context) {
	if subscription.Namespace != "default " {
		// create testing namespace
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			err := k8sClient.Create(ctx, namespace)
			fmt.Println(err)
			Expect(err).ShouldNot(HaveOccurred())
		}
	}
	// create subscription
	err := k8sClient.Create(ctx, subscription)
	Expect(err).Should(BeNil())
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
			testing.IsK8sUnprocessableEntity(),
		),
	)
}

// fixtureValidSubscription returns a valid subscription
func fixtureValidSubscription(name, namespace string) *eventingv1alpha1.Subscription {
	return &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventingv1alpha1.SubscriptionSpec{
			ID:       subscriptionID,
			Protocol: "BEB",
			ProtocolSettings: &eventingv1alpha1.ProtocolSettings{
				ContentMode:     eventingv1alpha1.ProtocolSettingsContentModeBinary,
				ExemptHandshake: true,
				Qos:             "AT-LEAST_ONCE",
				WebhookAuth: &eventingv1alpha1.WebhookAuth{
					Type:         "oauth2",
					GrantType:    "client_credentials",
					ClientId:     "xxx",
					ClientSecret: "xxx",
					TokenUrl:     "https://oauth2.xxx.com/oauth2/token",
					Scope:        []string{"guid-identifier"},
				},
			},
			Sink: "https://webhook.xxx.com",
			Filter: &eventingv1alpha1.BebFilters{
				Dialect: "beb",
				Filters: []*eventingv1alpha1.BebFilter{
					{
						EventSource: &eventingv1alpha1.Filter{
							Type:     "exact",
							Property: "source",
							Value:    "/default/kyma/myinstance",
						},
						EventType: &eventingv1alpha1.Filter{
							Type:     "exact",
							Property: "type",
							Value:    "kyma.ev2.poc.event1.v1",
						},
					},
				},
			},
		},
	}
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
	logf.Log.V(1).Info("subscriptions", "subscriptions", subscriptionList)
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
	By(fmt.Sprintf("Using unique namespace name: %s", namespaceName))
	return namespaceName
}
