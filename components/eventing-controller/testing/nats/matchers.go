package nats

import (
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/jetstream"
)

func BeValidSubscription() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber nats.Subscriber) bool {
		return subscriber.IsValid()
	}, gomega.BeTrue())
}

func BeSubscriptionWithSubject(subject string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber nats.Subscriber) bool {
		return subscriber.SubscriptionSubject() == subject
	}, gomega.BeTrue())
}

func BeJetStreamSubscriptionWithSubject(subject string) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber nats.Subscriber) bool {
		info, err := subscriber.ConsumerInfo()
		if err != nil {
			return false
		}
		js := jetstream.JetStream{}
		return info.Config.FilterSubject == js.GetJetStreamSubject(subject)
	}, gomega.BeTrue())
}

func BeExistingSubscription() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber nats.Subscriber) bool {
		return subscriber != nil
	}, gomega.BeTrue())
}
