package controllers

import (
	"context"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

const (
	timeout               = time.Second * 10
	interval              = time.Millisecond * 250
	subscriptionName      = "test-subs-1"
	subscriptionNamespace = "test-subs-1"
	subscriptionID        = "test-subs-1"
)

var _ = ginkgo.Describe("Subscription", func() {
	// TODO: test required fields are provided  but with wrong values => basically testing the CRD schema
	// TODO: test required fields are provided => basically testing the CRD schema
	ginkgo.Context("When creating a valid BEB Subscription", func() {
		ginkgo.It("Should reconcile the Subscription", func() {
			ctx := context.Background()
			givenSubscription := fixtureSubscription()
			ensureSubscriptionCreated(givenSubscription, ctx)

			ginkgo.By("Setting a finalizer")
			subscriptionLookupKey := types.NamespacedName{Name: subscriptionName, Namespace: subscriptionNamespace}

			getSubscription(subscriptionLookupKey, ctx).Should(
				haveName(subscriptionName),
				haveFinalizer(finalizerName),
			)

			ginkgo.By("Creating a BEB Subscription")
			// TODO(nachtmaar): check that an HTTP call against BEB was done
		})
	})
})

func fixtureSubscription() *eventingv1alpha1.Subscription {
	// Create a valid subscription
	return &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      subscriptionName,
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
		namespace := v1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: subscription.Namespace,
			},
		}
		gomega.Expect(k8sClient.Create(ctx, &namespace)).Should(gomega.BeNil())
	}
	gomega.Expect(k8sClient.Create(ctx, subscription)).Should(gomega.BeNil())
}

// TODO: move matchers  to extra file ?
func haveName(name string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventingv1alpha1.Subscription) string { return s.Name }, gomega.Equal(name))
}

func haveFinalizer(finalizer string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventingv1alpha1.Subscription) []string { return s.ObjectMeta.Finalizers }, gomega.ContainElement(finalizer))
}
