package beb

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/beb"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/signals"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Commander implements the Commander interface.
type Commander struct {
	cancel           context.CancelFunc
	envCfg           *env.BebConfig
	logger           *logrus.Logger
	metricsCollector *metrics.Collector
	opts             *options.Options
}

// NewCommander creates the Commander for publisher to BEB.
func NewCommander(opts *options.Options, metricsCollector *metrics.Collector, logger *logrus.Logger) *Commander {
	return &Commander{
		metricsCollector: metricsCollector,
		logger:           logger,
		envCfg:           new(env.BebConfig),
		opts:             opts,
	}
}

// Init implements the Commander interface and initializes the publisher to BEB.
func (c *Commander) Init() error {
	if err := envconfig.Process("", c.envCfg); err != nil {
		c.logger.Errorf("Read BEB configuration failed with error: %s", err)
		return err
	}
	return nil
}

// Start implements the Commander interface and starts the publisher.
func (c *Commander) Start() error {
	c.logger.Infof("Starting Event Publisher to BEB, envCfg: %v; opts: %#v", c.envCfg.String(), c.opts)

	// configure message receiver
	messageReceiver := receiver.NewHttpMessageReceiver(c.envCfg.Port)

	// assure uniqueness
	var ctx context.Context
	ctx, c.cancel = context.WithCancel(signals.NewContext())

	// configure auth client
	client := oauth.NewClient(ctx, c.envCfg)
	defer client.CloseIdleConnections()

	// configure message sender
	messageSender := sender.NewBebMessageSender(c.envCfg.EmsPublishURL, client)

	// cluster config
	k8sConfig := config.GetConfigOrDie()

	// setup application lister
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	applicationLister := application.NewLister(ctx, dynamicClient)

	// configure legacyTransformer
	legacyTransformer := legacy.NewTransformer(
		c.envCfg.BEBNamespace,
		c.envCfg.EventTypePrefix,
		applicationLister,
	)

	// Configure Subscription Lister
	subDynamicSharedInfFactory := subscribed.GenerateSubscriptionInfFactory(k8sConfig)
	subLister := subDynamicSharedInfFactory.ForResource(subscribed.GVR).Lister()
	subscribedProcessor := &subscribed.Processor{
		SubscriptionLister: &subLister,
		Config:             c.envCfg,
		Logger:             c.logger,
	}
	// Sync informer cache or die
	c.logger.Info("Waiting for informers caches to sync")
	informers.WaitForCacheSyncOrDie(ctx, subDynamicSharedInfFactory)
	c.logger.Info("Informers are synced successfully")

	// start handler which blocks until it receives a shutdown signal
	if err := beb.NewHandler(messageReceiver, messageSender, c.envCfg.RequestTimeout, legacyTransformer, c.opts,
		subscribedProcessor, c.logger, c.metricsCollector).Start(ctx); err != nil {
		c.logger.Errorf("Start handler failed with error: %s", err)
		return err
	}
	c.logger.Info("Shutdown the Event Publisher Proxy")
	return nil
}

// Stop implements the Commander interface and stops the publisher.
func (c *Commander) Stop() error {
	c.cancel()
	return nil
}
