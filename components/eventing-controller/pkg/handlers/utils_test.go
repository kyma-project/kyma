package handlers

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	bebtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/api/events/types"
)

func Test_getHash(t *testing.T) {
	g := NewGomegaWithT(t)

	bebSubscription := bebtypes.Subscription{}
	hash, err := getHash(&bebSubscription)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(hash).To(BeNumerically(">", 0))
}

func Test_getInternalView4Ev2(t *testing.T) {
	g := NewGomegaWithT(t)

	// given
	subscription := &eventingv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Subscription",
			APIVersion: "eventing.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace",
		},
		// TODO: validate all fields from here in the controller
		Spec: eventingv1alpha1.SubscriptionSpec{
			Id:       "id",
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

	// then
	bebSubscription, err := getInternalView4Ev2(subscription)

	// when
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(bebSubscription.Name).To(BeEquivalentTo(subscription.Name))
	g.Expect(bebSubscription.Events).To(BeEquivalentTo(bebtypes.Events{
		{
			Source: "/default/kyma/myinstance",
			Type:   "kyma.ev2.poc.event1.v1",
		},
	}))
	g.Expect(bebSubscription)

	g.Expect(bebSubscription.ExemptHandshake).To(BeEquivalentTo(subscription.Spec.ProtocolSettings.ExemptHandshake))
	g.Expect(bebSubscription.WebhookUrl).To(BeEquivalentTo(subscription.Spec.Sink))
	g.Expect(bebSubscription.Qos).To(BeEquivalentTo(bebtypes.QosAtLeastOnce))
	// TODO: test all other attributes as well
}
