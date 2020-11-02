package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/handlers"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/util/intstr"

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
				testing.HaveSubscriptionFinalizer(SubscriptionFinalizer),
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

	FWhen("Subscription changed", func() {
		It("Should update the BEB subscription", func() {
			subscriptionName := "test-subscription-beb-not-status-not-ready"
			ctx := context.Background()

			oldSvc := NewSubscriberSvc("webhook-old", namespaceName)
			ensureSubscriberSvcCreated(oldSvc, ctx)

			newSvc := NewSubscriberSvc("webhook-new", namespaceName)
			ensureSubscriberSvcCreated(newSvc, ctx)

			// cheating
			givenSubscription := NewSubscription(subscriptionName, namespaceName, WithFilter, WithWebhook)
			WithValidSink(oldSvc.Namespace, oldSvc.Name, givenSubscription)
			ensureSubscriptionCreated(givenSubscription, ctx)

			apiRuleForOldSvc := handlers.NewAPIRule(givenSubscription, handlers.WithoutPath, handlers.WithGateway, handlers.WithService, handlers.WithStatusReady)
			apiRuleForOldSvc.Namespace = namespaceName
			apiRuleForOldSvc.Labels = map[string]string{
				ControllerServiceLabelKey:  oldSvc.Name,
				ControllerIdentityLabelKey: ControllerIdentityLabelValue,
			}
			ensureAPIRuleCreated(apiRuleForOldSvc, ctx)

			apiRuleForNewSvc := handlers.NewAPIRule(givenSubscription, handlers.WithoutPath, handlers.WithGateway, handlers.WithService, handlers.WithStatusReady)
			apiRuleForNewSvc.Namespace = namespaceName
			apiRuleForNewSvc.Labels = map[string]string{
				ControllerServiceLabelKey:  newSvc.Name,
				ControllerIdentityLabelKey: ControllerIdentityLabelValue,
			}
			ensureAPIRuleCreated(apiRuleForNewSvc, ctx)

			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}

			By("Given subscription is ready")
			var subscription = &eventingv1alpha1.Subscription{}
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(testing.HaveSubscriptionReady())
			//waitForSubscriptionReady(subscriptionLookupKey, ctx)

			By("Updating the sink")
			subscription.Spec.Sink = fmt.Sprintf("http://%s.%s.svc.cluster.local", newSvc.Name, newSvc.Namespace)
			updateSubscription(subscription, ctx).Should(testing.HaveSubscriptionReady())

			By("Updating the BEB Subscription with the new sink")
			bebCreationRequests := make([]bebtypes.Subscription, 0)
			getBebSubscriptionCreationRequests(bebCreationRequests).Should(And(
				ContainElement(MatchFields(IgnoreMissing|IgnoreExtras,
					Fields{
						"Name":       BeEquivalentTo(subscription.Name),
						"WebhookUrl": BeEquivalentTo(subscription.Spec.Sink),
					},
				))))
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

// getSubscription fetches a subscription using the lookupKey and allows to make assertions on it
func getSubscription(subscription *eventingv1alpha1.Subscription, lookupKey types.NamespacedName, ctx context.Context) AsyncAssertion {
	return Eventually(func() eventingv1alpha1.Subscription {
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return eventingv1alpha1.Subscription{}
		}
		fmt.Println(">>>", subscription.Status.Ready)
		return *subscription
	}, time.Second*60, time.Second)
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
			testing.IsK8sUnprocessableEntity(),
		),
	)
}

func NewSubscriberSvc(name, ns string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol: "TCP",
					Port:     80,
					TargetPort: intstr.IntOrString{
						IntVal: 8080,
					},
				},
			},
			Selector: map[string]string{
				"test": "test",
			},
		},
	}
}

type subOpt func(subscription *eventingv1alpha1.Subscription)

func NewSubscription(name, ns string, opts ...subOpt) *eventingv1alpha1.Subscription {
	newSub := &eventingv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: eventingv1alpha1.SubscriptionSpec{},
	}
	for _, o := range opts {
		o(newSub)
	}
	return newSub
}

func WithoutWebhook(s *eventingv1alpha1.Subscription) {
	s.Spec.Protocol = "BEB"
	s.Spec.ProtocolSettings = &eventingv1alpha1.ProtocolSettings{
		ContentMode:     eventingv1alpha1.ProtocolSettingsContentModeBinary,
		ExemptHandshake: true,
		Qos:             "AT-LEAST_ONCE",
	}
}

func WithWebhook(s *eventingv1alpha1.Subscription) {
	s.Spec.Protocol = "BEB"
	s.Spec.ProtocolSettings = &eventingv1alpha1.ProtocolSettings{
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
	}
}

func WithFilter(s *eventingv1alpha1.Subscription) {
	s.Spec.Filter = &eventingv1alpha1.BebFilters{
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
	}
}

func WithValidSink(svcNs, svcName string, s *eventingv1alpha1.Subscription) {
	s.Spec.Sink = fmt.Sprintf("https://%s.%s.svc.cluster.local", svcName, svcNs)
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
			Sink: fmt.Sprintf("https://webhook.%s.svc.cluster.local", namespace),
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
