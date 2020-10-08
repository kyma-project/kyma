package controllers

import (
	"context"
	"time"

	"github.com/onsi/ginkgo"
	ginkgotable "github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

const (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250
	// subscriptionName      = "test-subs-1"
	// TODO(nachtmaar) switch back to custom namespace? there is a namespace deletion problem :/ or otherwise cleanup in each test respectively
	subscriptionNamespace = "default"
	subscriptionID        = "test-subs-1"

	namespaceCleanupTimeout  = time.Second * 30
	namespaceCleanupInterval = time.Second * 1
)

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

var _ = ginkgo.Describe("Subscription", func() {
	ginkgo.BeforeEach(func() {
		// Clean up all resources in test namespace
		// If a custom namespace is used in a test, the test is responsible for cleaning resources up
		ctx := context.Background()

		testNamespace := fixtureNamespace(subscriptionNamespace)
		if testNamespace.Name != "default" {
			gomega.Eventually(func() error {
				// logf.Log.V(1).Info("deleting namespace", "namespace", testNamespace)
				deletePolicy := metav1.DeletePropagationForeground
				// gracePeriod := int64(0)
				if err := k8sClient.Delete(ctx, testNamespace, &client.DeleteOptions{
					PropagationPolicy: &deletePolicy,
					// GracePeriodSeconds: &gracePeriod,
				}); err != nil && !errors.IsNotFound(err) {

					// TODO: is there an easier way to log in debug mode ?
					// logf.Log.V(1).Info("error while deleting namespace: ", "error", err)
					if err := printSubscriptions(ctx); err != nil {
						return err
					}

					if err := printNamespaces(testNamespace.Name, ctx); err != nil {
						return err
					}

					return err
				}

				return nil
			}, namespaceCleanupTimeout, namespaceCleanupInterval).Should(gomega.Or(
				gomega.BeNil(),
			))
		}

		_ = printSubscriptions(ctx)
		_ = printNamespaces(testNamespace.Name, ctx)
	})

	// TODO: test required fields are provided  but with wrong values => basically testing the CRD schema
	// TODO: test required fields are provided => basically testing the CRD schema
	ginkgo.Context("When creating a valid Subscription", func() {
		ginkgo.It("Should reconcile the Subscription", func() {
			subscriptionName := "test-valid-subscription-1"
			ctx := context.Background()
			givenSubscription := fixtureValidSubscription(subscriptionName)
			ensureSubscriptionCreated(givenSubscription, ctx)
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: subscriptionNamespace}

			ginkgo.By("Setting a finalizer")
			getSubscription(subscriptionLookupKey, ctx).Should(
				gomega.Not(gomega.BeNil()),
				haveName(subscriptionName),
				haveFinalizer(finalizerName),
			)

			ginkgo.By("Creating a BEB Subscription")
			// TODO(nachtmaar): check that an HTTP call against BEB was done

			ginkgo.By("Emitting some k8s events")
			// TODO(nachtmaar):
		})
	})

	ginkgo.Context("When deleting a valid Subscription", func() {
		ginkgo.It("Should reconcile the Subscription", func() {
			subscriptionName := "test-delete-valid-subscription-1"
			ctx := context.Background()
			givenSubscription := fixtureValidSubscription(subscriptionName)
			ensureSubscriptionCreated(givenSubscription, ctx)
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: subscriptionNamespace}

			// ensure subscription is given
			getSubscription(subscriptionLookupKey, ctx).Should(
				gomega.Not(gomega.BeNil()),
				haveName(subscriptionName),
				haveFinalizer(finalizerName),
			)

			ginkgo.By("Deleting the BEB Subscription")
			// TODO(nachtmaar): check that an HTTP call against BEB was done

			ginkgo.By("Removing the finalizer")
			getSubscription(subscriptionLookupKey, ctx).Should(
				gomega.Not(gomega.BeNil()),
				haveName(subscriptionName),
				gomega.Not(haveFinalizer(finalizerName)),
			)

			ginkgo.By("Emitting some k8s events")
			// TODO(nachtmaar):
		})
	})

	ginkgotable.DescribeTable("Schema tests",
		func(subscription *eventingv1alpha1.Subscription) {
			ctx := context.Background()
			ginkgo.By("Letting the APIServer reject the custom resource")
			ensureSubscriptionCreationFails(subscription, ctx)
		},
		ginkgotable.Entry("filter missing",
			func() *eventingv1alpha1.Subscription {
				subscription := fixtureValidSubscription("schema-filter-missing")
				subscription.Spec.Filter = nil
				return subscription
			}()),
		ginkgotable.Entry("protocolsettings missing",
			func() *eventingv1alpha1.Subscription {
				subscription := fixtureValidSubscription("schema-filter-missing")
				subscription.Spec.ProtocolSettings = nil
				return subscription
			}()),
		ginkgotable.Entry("protocolsettings.webhookauth missing",
			func() *eventingv1alpha1.Subscription {
				subscription := fixtureValidSubscription("schema-filter-missing")
				subscription.Spec.ProtocolSettings.WebhookAuth = nil
				return subscription
			}()),
		// TODO: find a way to set values to nil or remove in raw format, currently not testable with this test impl.
		// ginkgotable.Entry("protocol empty",
		// 	func() *eventingv1alpha1.Subscription {
		// 		subscription := fixtureValidSubscription("schema-filter-missing")
		// 		subscription.Spec.Protocol = ""
		// 		return subscription
		// 	}()),
	)

})

// fixtureValidSubscription returns a valid subscription
func fixtureValidSubscription(name string) *eventingv1alpha1.Subscription {
	return &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: subscriptionNamespace,
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
func getSubscription(lookupKey types.NamespacedName, ctx context.Context) gomega.AsyncAssertion {
	return gomega.Eventually(func() *eventingv1alpha1.Subscription {
		wantSubscription := &eventingv1alpha1.Subscription{}
		if err := k8sClient.Get(ctx, lookupKey, wantSubscription); err != nil {
			return nil
		}
		return wantSubscription
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
		// 		gomega.Expect(false)
		// 	}
		// }
		if namespace.Name != "default" {
			gomega.Expect(k8sClient.Create(ctx, namespace)).Should(gomega.Or(
				// ignore if namespaces is already created
				// isK8sAlreadyExistsError(),
				gomega.BeNil(),
			))
		}
	}
	gomega.Expect(k8sClient.Create(ctx, subscription)).Should(gomega.BeNil())
}

// ensureSubscriptionCreationFails creates a Subscription in the k8s cluster and ensures that it is reject because of invalid schema
func ensureSubscriptionCreationFails(subscription *eventingv1alpha1.Subscription, ctx context.Context) {
	if subscription.Namespace != "default " {
		namespace := fixtureNamespace(subscription.Namespace)
		if namespace.Name != "default" {
			gomega.Expect(k8sClient.Create(ctx, namespace)).Should(gomega.Or(
				// ignore if namespaces is already created
				// isK8sAlreadyExistsError(),
				gomega.BeNil(),
			))
		}
	}
	gomega.Expect(k8sClient.Create(ctx, subscription)).Should(
		gomega.And(
			// prevent nil-pointer stacktrace
			gomega.Not(gomega.BeNil()),
			isK8sUnprocessableEntity(),
		),
	)
	// gomega.Expect(getK8sError(k8sClient.Create(ctx, subscription)).Status()).Should(isK8sUnprocessableEntity())
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
func haveName(name string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventingv1alpha1.Subscription) string { return s.Name }, gomega.Equal(name))
}

func haveFinalizer(finalizer string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventingv1alpha1.Subscription) []string { return s.ObjectMeta.Finalizers }, gomega.ContainElement(finalizer))
}

func isK8sAlreadyExistsError() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(err *errors.StatusError) metav1.StatusReason { return err.ErrStatus.Reason }, gomega.Equal("AlreadyExists"))
}

func isK8sUnprocessableEntity() gomegatypes.GomegaMatcher {
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
	return gomega.WithTransform(func(err *errors.StatusError) metav1.StatusReason { return err.ErrStatus.Reason }, gomega.Equal(metav1.StatusReasonInvalid))
}

// func isK8sKnotFoundError() gomegatypes.GomegaMatcher {
// 	return gomega.WithTransform(func(err *errors.StatusError) metav1.StatusReason { return err.ErrStatus.Reason }, gomega.Equal("NotFound"))
// }
