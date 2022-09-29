//nolint:gosec
package utils

import (
	"fmt"
	"testing"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
	. "github.com/onsi/gomega"
)

func TestGetInternalView4Ev2(t *testing.T) {
	defaultProtocolSettings := &eventingv1alpha1.ProtocolSettings{
		ContentMode: func() *string {
			cm := types.ContentModeBinary
			return &cm
		}(),
		ExemptHandshake: func() *bool {
			eh := true
			return &eh
		}(),
		Qos: utils.StringPtr(string(types.QosAtLeastOnce)),
	}

	defaultWebhookAuth := &types.WebhookAuth{
		Type:         types.AuthTypeClientCredentials,
		GrantType:    types.GrantTypeClientCredentials,
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
		TokenURL:     "tokenURL",
	}
	defaultNameMapper := NewBEBSubscriptionNameMapper("my-shoot", 50)

	bebSubEvents := types.Events{types.Event{
		Source: eventingtesting.EventSource,
		Type:   eventingtesting.OrderCreatedEventType,
	}}

	defaultNamespace := "defaultNS"
	svcName := "foo-svc"
	host := "foo-host"
	scheme := "https"
	expectedWebhookURL := fmt.Sprintf("%s://%s", scheme, host)
	g := NewGomegaWithT(t)

	t.Run("subscription with protocol settings where defaults are overridden", func(t *testing.T) {
		// given
		subscription := eventingtesting.NewSubscription("name", "namespace",
			eventingtesting.WithOrderCreatedFilter(),
			eventingtesting.WithValidSink("ns", svcName),
		)

		subscription.Spec.ProtocolSettings = eventingtesting.NewProtocolSettings(
			eventingtesting.WithBinaryContentMode(),
			eventingtesting.WithExemptHandshake(),
			eventingtesting.WithAtLeastOnceQOS(),
			eventingtesting.WithDefaultWebhookAuth(),
		)

		// Values should be overridden by the given values in subscription
		expectedWebhookAuth := &types.WebhookAuth{
			Type:         types.AuthTypeClientCredentials,
			GrantType:    types.GrantTypeClientCredentials,
			ClientID:     subscription.Spec.ProtocolSettings.WebhookAuth.ClientID,
			ClientSecret: subscription.Spec.ProtocolSettings.WebhookAuth.ClientSecret,
			TokenURL:     subscription.Spec.ProtocolSettings.WebhookAuth.TokenURL,
		}
		expectedBEBSubscription := eventingtesting.NewBEBSubscription(
			defaultNameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace),
			*subscription.Spec.ProtocolSettings.ContentMode,
			expectedWebhookURL,
			bebSubEvents,
			expectedWebhookAuth,
		)

		apiRule := eventingtesting.NewAPIRule(subscription,
			eventingtesting.WithPath(),
			eventingtesting.WithService(svcName, host),
		)

		// then
		gotBEBSubscription, err := GetInternalView4Ev2(subscription, apiRule, defaultWebhookAuth, defaultProtocolSettings, "", defaultNameMapper)

		// when
		g.Expect(err).To(BeNil())
		g.Expect(*expectedBEBSubscription).To(Equal(*gotBEBSubscription))
	})

	t.Run("subscription with default setting", func(t *testing.T) {
		// given
		subscription := eventingtesting.NewSubscription("name", "namespace",
			eventingtesting.WithOrderCreatedFilter(),
			eventingtesting.WithValidSink("ns", svcName),
		)

		expectedBEBSubWithDefault := eventingtesting.NewBEBSubscription(
			defaultNameMapper.MapSubscriptionName(subscription.Name, subscription.Namespace),
			*defaultProtocolSettings.ContentMode,
			expectedWebhookURL,
			bebSubEvents,
			defaultWebhookAuth, // WebhookAuth should retain defaults
		)

		apiRule := eventingtesting.NewAPIRule(subscription,
			eventingtesting.WithPath(),
			eventingtesting.WithService(svcName, host),
		)

		// then
		gotBEBSubscription, err := GetInternalView4Ev2(subscription, apiRule, defaultWebhookAuth, defaultProtocolSettings, defaultNamespace, defaultNameMapper)

		// when
		g.Expect(err).To(BeNil())
		g.Expect(*expectedBEBSubWithDefault).To(Equal(*gotBEBSubscription))
	})

	t.Run("subscription with custom webhookauth config followed by a subscription with default webhookauth config should not alter the default config", func(t *testing.T) {
		// given
		subWithGivenWebhookAuth := eventingtesting.NewSubscription("name", "namespace",
			eventingtesting.WithOrderCreatedFilter(),
			eventingtesting.WithValidSink("ns", svcName),
		)

		subWithGivenWebhookAuth.Spec.ProtocolSettings = eventingtesting.NewProtocolSettings(
			eventingtesting.WithBinaryContentMode(),
			eventingtesting.WithExemptHandshake(),
			eventingtesting.WithAtLeastOnceQOS(),
			eventingtesting.WithDefaultWebhookAuth(),
		)
		expectedWebhookAuth := types.WebhookAuth{
			Type:         types.AuthTypeClientCredentials,
			GrantType:    types.GrantTypeClientCredentials,
			ClientID:     subWithGivenWebhookAuth.Spec.ProtocolSettings.WebhookAuth.ClientID,
			ClientSecret: subWithGivenWebhookAuth.Spec.ProtocolSettings.WebhookAuth.ClientSecret,
			TokenURL:     subWithGivenWebhookAuth.Spec.ProtocolSettings.WebhookAuth.TokenURL,
		}

		expectedBEBSubWithWebhookAuth := eventingtesting.NewBEBSubscription(
			defaultNameMapper.MapSubscriptionName(subWithGivenWebhookAuth.Name, subWithGivenWebhookAuth.Namespace),
			*subWithGivenWebhookAuth.Spec.ProtocolSettings.ContentMode,
			expectedWebhookURL,
			bebSubEvents,
			&expectedWebhookAuth, // WebhookAuth should retain the supplied config
		)

		apiRule := eventingtesting.NewAPIRule(subWithGivenWebhookAuth,
			eventingtesting.WithPath(),
			eventingtesting.WithService(svcName, host),
		)

		// then
		gotBEBSubscription, err := GetInternalView4Ev2(subWithGivenWebhookAuth, apiRule, defaultWebhookAuth, defaultProtocolSettings, defaultNamespace, defaultNameMapper)

		// when
		g.Expect(err).To(BeNil())
		g.Expect(*expectedBEBSubWithWebhookAuth).To(Equal(*gotBEBSubscription))

		// Use another subscription without webhookAuthConfig
		// given
		subscriptionWithoutWebhookAuth := eventingtesting.NewSubscription("name", "namespace",
			eventingtesting.WithOrderCreatedFilter(),
			eventingtesting.WithValidSink("ns", svcName),
		)

		expectedBEBSubWithDefault := eventingtesting.NewBEBSubscription(
			defaultNameMapper.MapSubscriptionName(subscriptionWithoutWebhookAuth.Name, subscriptionWithoutWebhookAuth.Namespace),
			*subWithGivenWebhookAuth.Spec.ProtocolSettings.ContentMode,
			expectedWebhookURL,
			bebSubEvents,
			defaultWebhookAuth, // WebhookAuth should retain defaults
		)

		apiRule = eventingtesting.NewAPIRule(subscriptionWithoutWebhookAuth,
			eventingtesting.WithPath(),
			eventingtesting.WithService(svcName, host),
		)

		// then
		gotBEBSubWithDefaultCfg, err := GetInternalView4Ev2(subscriptionWithoutWebhookAuth, apiRule, defaultWebhookAuth, defaultProtocolSettings, defaultNamespace, defaultNameMapper)

		// when
		g.Expect(err).To(BeNil())
		g.Expect(*expectedBEBSubWithDefault).To(Equal(*gotBEBSubWithDefaultCfg))
		g.Expect(*expectedBEBSubWithDefault.WebhookAuth).To(Equal(*gotBEBSubWithDefaultCfg.WebhookAuth))

	})
}
