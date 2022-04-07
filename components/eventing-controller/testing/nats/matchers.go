package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
)

func BeValidSubscription() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscription *nats.Subscription) bool {
		return subscription.IsValid()
	}, gomega.BeTrue())
}

func BeSubscriptionWithSubject(subject string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscription *nats.Subscription) bool {
		return subscription.Subject == subject
	}, gomega.BeTrue())
}

func BeJetStreamSubscriptionWithSubject(subject string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscription *nats.Subscription) bool {
		info, err := subscription.ConsumerInfo()
		if err != nil {
			return false
		}
		return info.Config.FilterSubject == subject
	}, gomega.BeTrue())
}

func BeExistingSubscription() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscription *nats.Subscription) bool {
		return subscription != nil
	}, gomega.BeTrue())
}
