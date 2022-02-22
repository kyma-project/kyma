package handlers

import (
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

var _ MessagingBackend = &JetStream{}

type JetStream struct {
	config env.NatsConfig
	logger *logger.Logger
}

func NewJetStream(config env.NatsConfig, logger *logger.Logger) *JetStream {
	return &JetStream{
		logger: logger,
		config: config,
	}
}

func (js *JetStream) Initialize(_ env.Config) error {
	// TODO: implement me
	return nil
}

func (js *JetStream) SyncSubscription(_ *eventingv1alpha1.Subscription, _ ...interface{}) (bool, error) {
	panic("implement me")
}

func (js *JetStream) DeleteSubscription(_ *eventingv1alpha1.Subscription) error {
	panic("implement me")
}
