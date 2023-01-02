package nats

import (
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"

	backendnats "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/nats/jetstream"
)

func BeValidSubscription() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber backendnats.Subscriber) bool {
		return subscriber.IsValid()
	}, gomega.BeTrue())
}

func BeJetStreamSubscriptionWithSubject(subject string, natsConfig backendnats.Config) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber backendnats.Subscriber) bool {
		info, err := subscriber.ConsumerInfo()
		if err != nil {
			return false
		}
		js := jetstream.JetStream{
			Config: natsConfig,
		}
		return info.Config.FilterSubject == js.GetJetStreamSubject(subject)
	}, gomega.BeTrue())
}

func BeExistingSubscription() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber backendnats.Subscriber) bool {
		return subscriber != nil
	}, gomega.BeTrue())
}
