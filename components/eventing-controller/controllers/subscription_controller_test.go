package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/onsi/gomega/types"
	"golang.org/x/oauth2"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/config"
	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/types"
	// gcp auth etc.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// +kubebuilder:scaffold:imports
)

const (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250
	// subscriptionName      = "test-subs-1"
	// TODO(nachtmaar) switch back to custom namespace? there is a namespace deletion problem :/ or otherwise cleanup in each test respectively
	subscriptionNamespacePrefix = "test-"
	subscriptionID              = "test-subs-1"

	namespaceCleanupTimeout  = time.Second * 30
	namespaceCleanupInterval = time.Second * 1
)

func generateTestSuiteID() int {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	return seededRand.Int()
}

func getUniqueNamespaceName() string {
	testSuiteID := generateTestSuiteID()
	namespaceName := fmt.Sprintf("%s%d", subscriptionNamespacePrefix, testSuiteID)
	return namespaceName
}

var _ = Describe("Subscription", func() {
	var beb *bebMock
	var bebConfig *config.Config

	// enable me for debugging
	// SetDefaultEventuallyTimeout(time.Minute)
	// SetDefaultEventuallyPollingInterval(time.Second)

	BeforeEach(func() {

		By("Preparing BEB Mock")
		err := os.Setenv("CLIENT_ID", "foo")
		Expect(err).ShouldNot(HaveOccurred())
		err = os.Setenv("CLIENT_SECRET", "foo")
		Expect(err).ShouldNot(HaveOccurred())

		beb = &bebMock{}
		bebURI := beb.start()
		logf.Log.Info("beb mock listening at", "address", bebURI)
		authURL := fmt.Sprintf("%s%s", bebURI, urlAuth)
		messagingURL := fmt.Sprintf("%s%s", bebURI, urlMessagingApi)

		err = os.Setenv("TOKEN_ENDPOINT", authURL)
		Expect(err).ShouldNot(HaveOccurred())
		err = os.Setenv("BEB_API_URL", messagingURL)
		Expect(err).ShouldNot(HaveOccurred())

		bebConfig = config.GetDefaultConfig()
		beb.bebConfig = bebConfig
	})

	AfterEach(func() {
		logf.Log.Info("beb mock calls", "calls", fmt.Sprintf("%+v", beb.requests))
	})

	// TODO: test required fields are provided  but with wrong values => basically testing the CRD schema
	// TODO: test required fields are provided => basically testing the CRD schema
	Context("When creating a valid Subscription", func() {
		It("Should reconcile the Subscription", func() {
			namespaceName := getUniqueNamespaceName()
			subscriptionName := "test-valid-subscription-1"
			ctx := context.Background()
			givenSubscription := fixtureValidSubscription(subscriptionName, namespaceName)
			ensureSubscriptionCreated(givenSubscription, ctx)
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}

			By("Setting a finalizer")
			var subscription = &eventingv1alpha1.Subscription{}
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(
				Not(BeNil()),
				haveName(subscriptionName),
				haveFinalizer(FinalizerName),
			)

			By("Setting a Subscribed condition")
			condition := eventingv1alpha1.MakeCondition(eventingv1alpha1.ConditionSubscribed, "TODO", v1.ConditionTrue)
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(
				Not(BeNil()),
				haveName(subscriptionName),
				haveFinalizer(FinalizerName),
				haveCondition(condition),
			)

			By("Creating a BEB Subscription")
			// TODO(nachtmaar): check that an HTTP call against BEB was done

			By("Emitting some k8s events")
			var subscriptionEvents = v1.EventList{}
			// TODO: adjust to event we want to have
			event := v1.Event{
				Reason:  "todo",
				Message: "todo",
				Type:    v1.EventTypeNormal,
			}
			Eventually(func() v1.EventList {
				err := k8sClient.List(ctx, &subscriptionEvents, client.InNamespace(namespaceName))
				if err != nil {
					return v1.EventList{}
				}
				return subscriptionEvents

			}).Should(haveEvent(event))

		})
	})

	Context("When deleting a valid Subscription", func() {
		It("Should reconcile the Subscription", func() {
			namespaceName := getUniqueNamespaceName()
			By(fmt.Sprintf("Using unique namespace name: %s", namespaceName))

			subscriptionName := "test-delete-valid-subscription-1"
			ctx := context.Background()
			givenSubscription := fixtureValidSubscription(subscriptionName, namespaceName)
			ensureSubscriptionCreated(givenSubscription, ctx)
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: namespaceName}

			// ensure subscription is given
			var subscription = &eventingv1alpha1.Subscription{}
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(
				Not(BeNil()),
				haveName(subscriptionName),
				haveFinalizer(FinalizerName),
			)

			By("Deleting the BEB Subscription")
			// TODO(nachtmaar): check that an HTTP call against BEB was done

			By("Removing the finalizer")
			getSubscription(subscription, subscriptionLookupKey, ctx).Should(
				Not(BeNil()),
				haveName(subscriptionName),
				Not(haveFinalizer(FinalizerName)),
			)

			By("Emitting some k8s events")

			// TODO(nachtmaar):
		})
	})

	DescribeTable("Schema tests",

		func(subscription *eventingv1alpha1.Subscription) {
			ctx := context.Background()
			namespaceName := getUniqueNamespaceName()
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
		Entry("protocolsettings.webhookauth missing",
			func() *eventingv1alpha1.Subscription {
				subscription := fixtureValidSubscription("schema-filter-missing", "")
				subscription.Spec.ProtocolSettings.WebhookAuth = nil
				return subscription
			}()),
		// TODO: find a way to set values to nil or remove in raw format, currently not testable with this test impl.
		// Entry("protocol empty",
		// 	func() *eventingv1alpha1.Subscription {
		// 		subscription := fixtureValidSubscription("schema-filter-missing")
		// 		subscription.Spec.Protocol = ""
		// 		return subscription
		// 	}()),
	)

})

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
		// TODO: validate all fields from here in the controller
		Spec: eventingv1alpha1.SubscriptionSpec{
			Id:       subscriptionID,
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

// TODO: document
func getSubscription(subscription *eventingv1alpha1.Subscription, lookupKey types.NamespacedName, ctx context.Context) AsyncAssertion {
	return Eventually(func() *eventingv1alpha1.Subscription {
		if err := k8sClient.Get(ctx, lookupKey, subscription); err != nil {
			return nil
		}
		return subscription
	}, timeout, interval)
}

// ensureSubscriptionCreated creates a Subscription in the k8s cluster. If a custom namespace is used, it will be created as well.
func ensureSubscriptionCreated(subscription *eventingv1alpha1.Subscription, ctx context.Context) {
	if subscription.Namespace != "default " {
		namespace := fixtureNamespace(subscription.Namespace)
		// TODO:
		// err := k8sClient.Create(ctx, &namespace)
		// if e, ok := err.(*errors.StatusError); ok {
		// 	if e.ErrStatus.Code == 409 && e.ErrStatus.Reason == "AlreadyExists" {
		// 		fmt.Printf("ignorning that namespace already exists")

		// 	} else {
		// 		Expect(false)
		// 	}
		// }
		if namespace.Name != "default" {
			Expect(k8sClient.Create(ctx, namespace)).Should(Or(
				// ignore if namespaces is already created
				// isK8sAlreadyExistsError(),
				BeNil(),
			))
		}
	}
	Expect(k8sClient.Create(ctx, subscription)).Should(BeNil())
}

// ensureSubscriptionCreationFails creates a Subscription in the k8s cluster and ensures that it is reject because of invalid schema
func ensureSubscriptionCreationFails(subscription *eventingv1alpha1.Subscription, ctx context.Context) {
	if subscription.Namespace != "default " {
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			Expect(k8sClient.Create(ctx, namespace)).Should(Or(
				// ignore if namespaces is already created
				// isK8sAlreadyExistsError(),
				BeNil(),
			))
		}
	}
	Expect(k8sClient.Create(ctx, subscription)).Should(
		And(
			// prevent nil-pointer stacktrace
			Not(BeNil()),
			isK8sUnprocessableEntity(),
		),
	)
	// Expect(getK8sError(k8sClient.Create(ctx, subscription)).Status()).Should(isK8sUnprocessableEntity())
}

func getK8sError(err error) *errors.StatusError {
	switch e := err.(type) {
	case *errors.StatusError:
		return e
	default:
		return nil
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

// TODO: add subscription prefix or move to subscription package
// TODO: move matchers  to extra file ?
func haveName(name string) GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) string { return s.Name }, Equal(name))
}

func haveFinalizer(finalizer string) GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) []string { return s.ObjectMeta.Finalizers }, ContainElement(finalizer))
}

func haveCondition(condition eventingv1alpha1.Condition) GomegaMatcher {
	return WithTransform(func(s *eventingv1alpha1.Subscription) []eventingv1alpha1.Condition { return s.Status.Conditions }, ContainElement(condition))
}

func isK8sAlreadyExistsError() GomegaMatcher {
	return WithTransform(func(err *errors.StatusError) metav1.StatusReason { return err.ErrStatus.Reason }, Equal("AlreadyExists"))
}

func haveEvent(event v1.Event) GomegaMatcher {
	return WithTransform(func(l v1.EventList) []v1.Event { return l.Items }, ContainElement(MatchFields(IgnoreExtras|IgnoreMissing, Fields{
		"Reason":  Equal(event.Reason),
		"Message": Equal(event.Message),
		"Type":    Equal(event.Type),
	})))
}

func isK8sUnprocessableEntity() GomegaMatcher {
	// TODO: also check for status code 422
	//  <*errors.StatusError | 0xc0001330e0>: {
	//     ErrStatus: {
	//         TypeMeta: {Kind: "", APIVersion: ""},
	//         ListMeta: {
	//             SelfLink: "",
	//             ResourceVersion: "",
	//             Continue: "",
	//             RemainingItemCount: nil,
	//         },
	//         Status: "Failure",
	//         Message: "Subscription.eventing.kyma-project.io \"test-valid-subscription-1\" is invalid: spec.filter: Invalid value: \"null\": spec.filter in body must be of type object: \"null\"",
	//         Reason: "Invalid",
	//         Details: {
	//             Name: "test-valid-subscription-1",
	//             Group: "eventing.kyma-project.io",
	//             Kind: "Subscription",
	//             UID: "",
	//             Causes: [
	//                 {
	//                     Type: "FieldValueInvalid",
	//                     Message: "Invalid value: \"null\": spec.filter in body must be of type object: \"null\"",
	//                     Field: "spec.filter",
	//                 },
	//             ],
	//             RetryAfterSeconds: 0,
	//         },
	//         Code: 422,
	//     },
	// }
	return WithTransform(func(err *errors.StatusError) metav1.StatusReason { return err.ErrStatus.Reason }, Equal(metav1.StatusReasonInvalid))
}

// func isK8sKnotFoundError() GomegaMatcher {
// 	return WithTransform(func(err *errors.StatusError) metav1.StatusReason { return err.ErrStatus.Reason }, Equal("NotFound"))
// }

func printSubscriptions(ctx context.Context) error {
	// print subscription details
	subscriptionList := eventingv1alpha1.SubscriptionList{}
	if err := k8sClient.List(ctx, &subscriptionList); err != nil {
		logf.Log.V(1).Info("error while getting subscription list", "error", err)
		return err
	}
	logf.Log.V(1).Info("subscriptions", "subscriptions", subscriptionList)
	return nil
}

func printNamespaces(namespaceName string, ctx context.Context) error {
	namespace := v1.Namespace{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: namespaceName}, &namespace); err != nil && !errors.IsNotFound(err) {
		logf.Log.V(1).Info("error while getting namespace", "error", err)
		return err
	}
	logf.Log.V(1).Info("namespace", "namespace", namespace)
	return nil
}

const (
	urlAuth         = "/auth"
	urlMessagingApi = "/messaging"
)

// bebMock implements a mock for BEB
type bebMock struct {
	requests  []http.Request
	bebConfig *config.Config
}

func (m *bebMock) start() string {
	// implementation based on https://pages.github.tools.sap/KernelServices/APIDefinitions/?urls.primaryName=Business%20Event%20Bus%20-%20CloudEvents
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logf.Log.WithName("beb mock")
		payload, err := ioutil.ReadAll(r.Body)
		Expect(err).ShouldNot(HaveOccurred())

		log.V(1).Info("received request",
			"uri", r.RequestURI,
			"method", r.Method,
			"payload", payload,
		)
		config.GetDefaultConfig()

		m.requests = append(m.requests, *r)

		w.Header().Set("Content-Type", "application/json")

		// oauth2 request
		if r.Method == "POST" && r.RequestURI == urlAuth {
			bebAuthResponseSuccess(w)
			return
		}
		// messaging API request
		if strings.HasPrefix(r.RequestURI, urlMessagingApi) {
			switch r.Method {
			case http.MethodDelete:
				bebDeleteResponseSuccess(w)
			case http.MethodPost:
				bebCreateSuccess(w)
			case http.MethodGet:
				switch r.RequestURI {
				case m.bebConfig.ListURL:
					bebListSuccess(w)
				// get on a single subscription
				default:
					parsedUrl, err := url.Parse(r.RequestURI)
					Expect(err).ShouldNot(HaveOccurred())
					subscriptionName := parsedUrl.Path
					bebGetSuccess(w, subscriptionName)
				}
				return
			default:
				w.WriteHeader(http.StatusNotImplemented)
			}
			return
		}
	}))
	uri := ts.URL

	return uri
}

func bebAuthResponseSuccess(w http.ResponseWriter) {
	token := oauth2.Token{
		AccessToken:  "some-token",
		TokenType:    "",
		RefreshToken: "",
	}
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(token)
	Expect(err).ShouldNot(HaveOccurred())
}

func bebCreateSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
	response := bebtypes.CreateResponse{
		Response: bebtypes.Response{
			StatusCode: http.StatusAccepted,
			Message:    "",
		},
		Href: "",
	}
	err := json.NewEncoder(w).Encode(response)
	Expect(err).ShouldNot(HaveOccurred())
}

func bebGetSuccess(w http.ResponseWriter, name string) {
	w.WriteHeader(http.StatusOK)
	s := bebtypes.Subscription{
		Name:               name,
		SubscriptionStatus: bebtypes.SubscriptionStatusActive,
	}
	err := json.NewEncoder(w).Encode(s)
	Expect(err).ShouldNot(HaveOccurred())
}

func bebListSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
	response := bebtypes.Response{
		StatusCode: http.StatusOK,
		Message:    "",
	}
	err := json.NewEncoder(w).Encode(response)
	Expect(err).ShouldNot(HaveOccurred())
}

func bebDeleteResponseSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
