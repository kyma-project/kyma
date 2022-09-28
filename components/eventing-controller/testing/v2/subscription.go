package v2

import (
	"fmt"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
)

type SubscriptionOpt func(subscription *v1alpha2.Subscription)

func WithMaxInFlight(maxInFlight string) SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		sub.Spec.Config = map[string]string{
			v1alpha2.MaxInFlightMessages: fmt.Sprint(maxInFlight),
		}
	}
}

func WithSource(source string) SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		sub.Spec.Source = source
	}
}

func WithTypes(types []string) SubscriptionOpt {
	return func(sub *v1alpha2.Subscription) {
		sub.Spec.Types = types
	}
}

func WithWebhookAuthForBEB() SubscriptionOpt {
	return func(s *v1alpha2.Subscription) {
		s.Spec.Config = map[string]string{
			v1alpha2.Protocol:                        "BEB",
			v1alpha2.ProtocolSettingsContentMode:     v1alpha1.ProtocolSettingsContentModeBinary,
			v1alpha2.ProtocolSettingsExemptHandshake: "true",
			v1alpha2.ProtocolSettingsQos:             "true",
			v1alpha2.WebhookAuthType:                 "oauth2",
			v1alpha2.WebhookAuthGrantType:            "client_credentials",
			v1alpha2.WebhookAuthClientID:             "xxx",
			v1alpha2.WebhookAuthClientSecret:         "xxx",
			v1alpha2.WebhookAuthTokenURL:             "https://oauth2.xxx.com/oauth2/token",
			v1alpha2.WebhookAuthScope:                "guid-identifier,root",
		}
	}
}

func WithProtocolBEB() SubscriptionOpt {
	return func(s *v1alpha2.Subscription) {
		if s.Spec.Config == nil {
			s.Spec.Config = map[string]string{}
		}
		s.Spec.Config[v1alpha2.Protocol] = "BEB"
	}
}
