package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/signals"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	logger := logrus.New()
	logger.Info("Event Publisher NATS started")
	opts := options.ParseArgs()
	cfgNats := new(env.NatsConfig)
	if err := envconfig.Process("", cfgNats); err != nil {
		logger.Fatalf("Read NATS configuration failed with error: %s", err)
	}

	// configure message receiver
	messageReceiver := receiver.NewHttpMessageReceiver(cfgNats.Port)

	// assure uniqueness
	ctx := signals.NewContext()

	// configure message sender
	messageSenderToNats := sender.NewNatsMessageSender(ctx, cfgNats.NatsPublishURL, logger)

	// configure legacyTransformer
	legacyTransformer := legacy.NewTransformer(
		cfgNats.ToConfig().BEBNamespace,
		cfgNats.ToConfig().EventTypePrefix,
	)

	// Configure Subscription Lister
	k8sConfig := config.GetConfigOrDie()
	subDynamicSharedInfFactory := subscribed.GenerateSubscriptionInfFactory(k8sConfig)
	subLister := subDynamicSharedInfFactory.ForResource(subscribed.GVR).Lister()
	subscribedProcessor := &subscribed.Processor{
		SubscriptionLister: &subLister,
		Config:             cfgNats.ToConfig(),
		Logger:             logger,
	}
	// Sync informer cache or die
	logger.Info("Waiting for informers caches to sync")
	subscribed.WaitForCacheSyncOrDie(ctx, subDynamicSharedInfFactory)
	logger.Info("Informers are synced successfully")

	// start handler which blocks until it receives a shutdown signal
	if err := handler.NewNatsHandler(messageReceiver, messageSenderToNats, cfgNats.RequestTimeout, legacyTransformer, opts, subscribedProcessor, logger).Start(ctx); err != nil {
		logger.Fatalf("Start handler failed with error: %s", err)
	}

	logger.Info("Event Publisher NATS shutdown")
}
