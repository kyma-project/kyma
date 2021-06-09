package handlers

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	controllertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func Test_SyncBebSubscription(t *testing.T) {
	g := NewGomegaWithT(t)

	// given
	beb := Beb{
		Log: ctrl.Log,
	}
	clientId := "client-id"
	clientSecret := "client-secret"
	tokenEndpoint := "token-endpoint"
	envConfig := env.Config{
		ClientID:      clientId,
		ClientSecret:  clientSecret,
		TokenEndpoint: tokenEndpoint,
	}
	err := os.Setenv("CLIENT_ID", clientId)
	g.Expect(err).ShouldNot(HaveOccurred())
	err = os.Setenv("CLIENT_SECRET", clientSecret)
	g.Expect(err).ShouldNot(HaveOccurred())
	err = os.Setenv("TOKEN_ENDPOINT", tokenEndpoint)
	g.Expect(err).ShouldNot(HaveOccurred())

	beb.Initialize(envConfig)

	// when
	subscription := fixtureValidSubscription("some-name", "some-namespace")
	subscription.Status.Emshash = 0
	subscription.Status.Ev2hash = 0

	apiRule := controllertesting.NewAPIRule(subscription, controllertesting.WithPath)
	controllertesting.WithService("foo-host", "foo-svc", apiRule)

	// then
	changed, err := beb.SyncSubscription(subscription, nil, apiRule)
	g.Expect(err).To(Not(BeNil()))
	g.Expect(changed).To(BeFalse())
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
			ID:       "id",
			Protocol: "BEB",
			ProtocolSettings: &eventingv1alpha1.ProtocolSettings{
				ContentMode: func() *string {
					cm := eventingv1alpha1.ProtocolSettingsContentModeBinary
					return &cm
				}(),
				ExemptHandshake: func() *bool {
					eh := true
					return &eh
				}(),
				Qos: func() *string {
					qos := "AT-LEAST_ONCE"
					return &qos
				}(),
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
							Value:    controllertesting.EventSource,
						},
						EventType: &eventingv1alpha1.Filter{
							Type:     "exact",
							Property: "type",
							Value:    controllertesting.EventTypeNotClean,
						},
					},
				},
			},
		},
	}
}
