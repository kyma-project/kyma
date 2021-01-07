package handlers

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func Test_getHash(t *testing.T) {
	g := NewGomegaWithT(t)

	bebSubscription := types.Subscription{}
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
		Spec: eventingv1alpha1.SubscriptionSpec{
			ID:       "id",
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
			Sink: "https://foo-host",
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

	apiRule := reconcilertesting.NewAPIRule(subscription, reconcilertesting.WithPath)
	reconcilertesting.WithService("foo-host", "foo-svc", apiRule)

	defaultWebhookAuth := &types.WebhookAuth{}

	// then
	bebSubscription, err := getInternalView4Ev2(subscription, apiRule, defaultWebhookAuth)

	// when
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(bebSubscription.Name).To(BeEquivalentTo(subscription.Name))
	g.Expect(bebSubscription.Events).To(BeEquivalentTo(types.Events{
		{
			Source: "/default/kyma/myinstance",
			Type:   "kyma.ev2.poc.event1.v1",
		},
	}))
	g.Expect(bebSubscription)

	g.Expect(bebSubscription.ExemptHandshake).To(BeEquivalentTo(subscription.Spec.ProtocolSettings.ExemptHandshake))
	g.Expect(bebSubscription.WebhookUrl).To(BeEquivalentTo(subscription.Spec.Sink))
	g.Expect(bebSubscription.Qos).To(BeEquivalentTo(types.QosAtLeastOnce))
	g.Expect(bebSubscription.WebhookAuth.ClientID).To(BeEquivalentTo(subscription.Spec.ProtocolSettings.WebhookAuth.ClientId))
	g.Expect(bebSubscription.WebhookAuth.ClientSecret).To(BeEquivalentTo(subscription.Spec.ProtocolSettings.WebhookAuth.ClientSecret))
	g.Expect(bebSubscription.WebhookAuth.GrantType).To(BeEquivalentTo(types.GrantTypeClientCredentials))
	g.Expect(bebSubscription.WebhookAuth.Type).To(BeEquivalentTo(types.AuthTypeClientCredentials))
	g.Expect(bebSubscription.WebhookAuth.TokenURL).To(BeEquivalentTo(subscription.Spec.ProtocolSettings.WebhookAuth.TokenUrl))
}

func Test_getInternalView4Ems(t *testing.T) {
	g := NewGomegaWithT(t)

	// given
	emsSubscription := &types.Subscription{
		Name:            "ev2subs1",
		ContentMode:     types.ContentModeStructured,
		ExemptHandshake: true,
		Qos:             types.QosAtLeastOnce,
		WebhookUrl:      "https://webhook.xxx.com",

		Events: []types.Event{
			{
				Source: "/default/kyma/myinstance",
				Type:   "kyma.ev2.poc.event1.v1",
			},
		},
	}

	// then
	bebSubscription, err := getInternalView4Ems(emsSubscription)

	// when
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(bebSubscription.Name).To(BeEquivalentTo(emsSubscription.Name))
	g.Expect(bebSubscription.ContentMode).To(BeEquivalentTo(emsSubscription.ContentMode))
	g.Expect(bebSubscription.ExemptHandshake).To(BeEquivalentTo(emsSubscription.ExemptHandshake))
	g.Expect(bebSubscription.Qos).To(BeEquivalentTo(types.QosAtLeastOnce))
	g.Expect(bebSubscription.WebhookUrl).To(BeEquivalentTo(emsSubscription.WebhookUrl))

	g.Expect(bebSubscription.Events).To(BeEquivalentTo(types.Events{
		{
			Source: "/default/kyma/myinstance",
			Type:   "kyma.ev2.poc.event1.v1",
		},
	}))
	g.Expect(bebSubscription)
}

func TestGetRandSuffix(t *testing.T) {
	totalExecutions := 10
	lengthOfRandomSuffix := 6
	results := make(map[string]bool)
	for i := 0; i < totalExecutions; i++ {
		result := GetRandString(lengthOfRandomSuffix)
		if _, ok := results[result]; ok {
			t.Fatalf("generated string already exists: %s", result)
		}
		results[result] = true
	}
}
