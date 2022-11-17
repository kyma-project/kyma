package reconcilertesting

import (
	"fmt"

	eventingv1alpha2 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/jetstreamv2"
	"github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
)

func BeValidSubscription() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber jetstreamv2.Subscriber) bool {
		return subscriber.IsValid()
	}, gomega.BeTrue())
}

func BeNatsSubWithMaxPending(expectedMaxAckPending int) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber jetstreamv2.Subscriber) (int, error) {
		info, err := subscriber.ConsumerInfo()
		if err != nil {
			return -1, err
		}
		return info.Config.MaxAckPending, nil
	}, gomega.Equal(expectedMaxAckPending))
}

func BeJetStreamSubscriptionWithSubject(source, subject string,
	typeMatching eventingv1alpha2.TypeMatching) gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber jetstreamv2.Subscriber) (bool, error) {
		info, err := subscriber.ConsumerInfo()
		if err != nil {
			return false, err
		}
		js := jetstreamv2.JetStream{}
		result := info.Config.FilterSubject == js.GetJetStreamSubject(source, subject, typeMatching)
		if !result {
			return false, fmt.Errorf(
				"BeJetStreamSubscriptionWithSubject expected %v to be equal to %v",
				info.Config.FilterSubject,
				js.GetJetStreamSubject(source, subject, typeMatching),
			)
		}
		return true, nil
	}, gomega.BeTrue())
}

func BeExistingSubscription() gomegatypes.GomegaMatcher {
	return gomega.WithTransform(func(subscriber jetstreamv2.Subscriber) bool {
		return subscriber != nil
	}, gomega.BeTrue())
}
