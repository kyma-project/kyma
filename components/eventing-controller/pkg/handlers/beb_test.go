package handlers

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"

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
	nameMapper := NewBebSubscriptionNameMapper("shoot001", MaxBEBSubscriptionNameLength)
	beb := NewBEB(nameMapper, ctrl.Log)
	// start BEB Mock
	bebMock := startBebMock()
	envConf := env.Config{
		BebApiUrl:                bebMock.MessagingURL,
		ClientID:                 "client-id",
		ClientSecret:             "client-secret",
		TokenEndpoint:            bebMock.TokenURL,
		WebhookActivationTimeout: 0,
		WebhookTokenEndpoint:     "webhook-token-endpoint",
		Domain:                   "domain.com",
		EventTypePrefix:          controllertesting.EventTypePrefix,
		BEBNamespace:             "/default/ns",
		Qos:                      "AT_LEAST_ONCE",
	}

	err := beb.Initialize(envConf)
	g.Expect(err).To(BeNil())

	// when
	subscription := fixtureValidSubscription("some-name", "some-namespace")
	subscription.Status.Emshash = 0
	subscription.Status.Ev2hash = 0

	apiRule := controllertesting.NewAPIRule(subscription, controllertesting.WithPath)
	controllertesting.WithService("foo-host", "foo-svc", apiRule)

	// then
	changed, err := beb.SyncSubscription(subscription, &Cleaner{}, apiRule)
	g.Expect(err).To(BeNil())
	g.Expect(changed).To(BeTrue())
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

func startBebMock() *controllertesting.BebMock {
	bebConfig := &config.Config{}
	beb := controllertesting.NewBebMock(bebConfig)
	bebURI := beb.Start()
	tokenURL := fmt.Sprintf("%s%s", bebURI, controllertesting.TokenURLPath)
	messagingURL := fmt.Sprintf("%s%s", bebURI, controllertesting.MessagingURLPath)
	beb.TokenURL = tokenURL
	beb.MessagingURL = messagingURL
	bebConfig = config.GetDefaultConfig(messagingURL)
	beb.BebConfig = bebConfig
	return beb
}

type Cleaner struct {
}

func (c *Cleaner) Clean(eventType string) (string, error) {
	// Cleaning is not needed in this test
	return eventType, nil
}
