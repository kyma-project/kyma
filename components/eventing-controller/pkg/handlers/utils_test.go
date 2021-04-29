package handlers

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"

	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func TestGetHash(t *testing.T) {
	g := NewGomegaWithT(t)

	bebSubscription := types.Subscription{}
	hash, err := getHash(&bebSubscription)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(hash).To(BeNumerically(">", 0))
}

func TestGetInternalView4Ev2(t *testing.T) {
	defaultProtocolSettings := &eventingv1alpha1.ProtocolSettings{
		ContentMode: func() *string {
			cm := "test"
			return &cm
		}(),
		ExemptHandshake: func() *bool {
			eh := true
			return &eh
		}(),
		Qos: func() *string {
			qos := "AT_LEAST_ONCE"
			return &qos
		}(),
	}

	defaultWebhookAuth := &types.WebhookAuth{
		Type:         types.AuthTypeClientCredentials,
		GrantType:    types.GrantTypeClientCredentials,
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
		TokenURL:     "tokenURL",
	}
	defaultNamespace := "defaultNS"
	svcName := "foo-svc"
	host := "foo-host"
	scheme := "https"
	expectedWebhookURL := fmt.Sprintf("%s://%s", scheme, host)
	g := NewGomegaWithT(t)

	t.Run("subscription with protocolsettings where defaults are overriden", func(t *testing.T) {
		// given
		subscription := reconcilertesting.NewSubscription("name", "namespace", eventingtesting.WithEventTypeFilter)
		eventingtesting.WithValidSink("ns", svcName, subscription)

		givenProtocolSettings := &eventingv1alpha1.ProtocolSettings{
			ContentMode: func() *string {
				contentMode := eventingv1alpha1.ProtocolSettingsContentModeBinary
				return &contentMode
			}(),
			ExemptHandshake: func() *bool {
				exemptHandshake := true
				return &exemptHandshake
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
		}
		subscription.Spec.ProtocolSettings = givenProtocolSettings

		// Values should be overriden by the given values in subscription
		expectedBEBSubscription := types.Subscription{
			Name:            subscription.Name,
			ContentMode:     *givenProtocolSettings.ContentMode,
			Qos:             types.QosAtLeastOnce,
			ExemptHandshake: *givenProtocolSettings.ExemptHandshake,
			Events: types.Events{
				{
					Source: reconcilertesting.EventSource,
					Type:   reconcilertesting.EventType,
				},
			},
			WebhookUrl: expectedWebhookURL,
			WebhookAuth: &types.WebhookAuth{
				Type:         types.AuthTypeClientCredentials,
				GrantType:    types.GrantTypeClientCredentials,
				ClientID:     subscription.Spec.ProtocolSettings.WebhookAuth.ClientId,
				ClientSecret: subscription.Spec.ProtocolSettings.WebhookAuth.ClientSecret,
				TokenURL:     subscription.Spec.ProtocolSettings.WebhookAuth.TokenUrl,
			},
		}

		apiRule := reconcilertesting.NewAPIRule(subscription, reconcilertesting.WithPath)
		reconcilertesting.WithService(host, svcName, apiRule)

		// then
		gotBEBSubscription, err := getInternalView4Ev2(subscription, apiRule, defaultWebhookAuth, defaultProtocolSettings, "")

		// when
		g.Expect(err).To(BeNil())
		g.Expect(expectedBEBSubscription).To(Equal(*gotBEBSubscription))
	})

	t.Run("subscription with default setting", func(t *testing.T) {
		// given
		subscription := reconcilertesting.NewSubscription("name", "namespace", eventingtesting.WithEmptySourceEventType)
		eventingtesting.WithValidSink("ns", svcName, subscription)

		// Values should retain defaults
		expectedBEBSubscription := types.Subscription{
			Name: subscription.Name,
			Events: types.Events{
				{
					Source: defaultNamespace,
					Type:   reconcilertesting.EventType,
				},
			},
			WebhookUrl:      expectedWebhookURL,
			WebhookAuth:     defaultWebhookAuth,
			Qos:             types.QosAtLeastOnce,
			ExemptHandshake: *defaultProtocolSettings.ExemptHandshake,
			ContentMode:     *defaultProtocolSettings.ContentMode,
		}

		apiRule := reconcilertesting.NewAPIRule(subscription, reconcilertesting.WithPath)
		reconcilertesting.WithService(host, svcName, apiRule)

		// then
		gotBEBSubscription, err := getInternalView4Ev2(subscription, apiRule, defaultWebhookAuth, defaultProtocolSettings, defaultNamespace)

		// when
		g.Expect(err).To(BeNil())
		g.Expect(expectedBEBSubscription).To(Equal(*gotBEBSubscription))
	})
}

func TestGetInternalView4Ems(t *testing.T) {
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
				Source: reconcilertesting.EventSource,
				Type:   reconcilertesting.EventTypeNotClean,
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
			Source: reconcilertesting.EventSource,
			Type:   reconcilertesting.EventTypeNotClean,
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
