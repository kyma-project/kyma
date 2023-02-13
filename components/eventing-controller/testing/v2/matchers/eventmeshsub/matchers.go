package eventmeshsub

import (
	eventMeshtypes "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/types"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
)

func HaveWebhookAuth(webhookAuth eventMeshtypes.WebhookAuth) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventMeshtypes.Subscription) eventMeshtypes.WebhookAuth {
		return *s.WebhookAuth
	}, gomega.Equal(webhookAuth))
}

func HaveEvents(events eventMeshtypes.Events) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventMeshtypes.Subscription) eventMeshtypes.Events { return s.Events },
		gomega.Equal(events))
}

func HaveQoS(qos eventMeshtypes.Qos) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventMeshtypes.Subscription) eventMeshtypes.Qos { return s.Qos },
		gomega.Equal(qos))
}

func HaveExemptHandshake(exemptHandshake bool) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventMeshtypes.Subscription) bool { return s.ExemptHandshake },
		gomega.Equal(exemptHandshake))
}

func HaveWebhookURL(webhookURL string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventMeshtypes.Subscription) string { return s.WebhookURL },
		gomega.Equal(webhookURL))
}

func HaveStatusPaused() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventMeshtypes.Subscription) eventMeshtypes.SubscriptionStatus {
		return s.SubscriptionStatus
	}, gomega.Equal(eventMeshtypes.SubscriptionStatusPaused))
}

func HaveStatusActive() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventMeshtypes.Subscription) eventMeshtypes.SubscriptionStatus {
		return s.SubscriptionStatus
	}, gomega.Equal(eventMeshtypes.SubscriptionStatusActive))
}

func HaveContentMode(contentMode string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventMeshtypes.Subscription) string { return s.ContentMode },
		gomega.Equal(contentMode))
}

func HaveNonEmptyLastFailedDeliveryReason() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(s *eventMeshtypes.Subscription) string {
		return s.LastFailedDeliveryReason
	}, gomega.Not(gomega.BeEmpty()))
}
