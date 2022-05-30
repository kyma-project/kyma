package nats

import (
	"context"

	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // TODO: remove as this is only required in a dev setup
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/signals"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
)

// Commander implements the Commander interface.
type Commander struct {
	cancel           context.CancelFunc
	metricsCollector *metrics.Collector
	logger           *logrus.Logger
	envCfg           *env.NatsConfig
	opts             *options.Options
	jetstreamMode    bool
}

// NewCommander creates the Commander for publisher to NATS.
func NewCommander(opts *options.Options, metricsCollector *metrics.Collector, logger *logrus.Logger, jetstreamMode bool) *Commander {
	return &Commander{
		envCfg:           new(env.NatsConfig),
		logger:           logger,
		metricsCollector: metricsCollector,
		opts:             opts,
		jetstreamMode:    jetstreamMode,
	}
}

// Init implements the Commander interface and initializes the publisher to NATS.
func (c *Commander) Init() error {
	if err := envconfig.Process("", c.envCfg); err != nil {
		c.logger.Errorf("Read %s configuration failed with error: %s", c.getBackendName(), err)
		return err
	}
	return nil
}

// Start implements the Commander interface and starts the publisher.
func (c *Commander) Start() error {
	c.logger.Infof("Starting Event Publisher to %s, envCfg: %v; opts: %#v", c.getBackendName(), c.envCfg.String(), c.opts)

	// assure uniqueness
	var ctx context.Context
	ctx, c.cancel = context.WithCancel(signals.NewContext())

	// configure message receiver
	messageReceiver := receiver.NewHTTPMessageReceiver(c.envCfg.Port)

	// connect to nats
	connection, err := pkgnats.Connect(c.envCfg.URL,
		pkgnats.WithRetryOnFailedConnect(c.envCfg.RetryOnFailedConnect),
		pkgnats.WithMaxReconnects(c.envCfg.MaxReconnects),
		pkgnats.WithReconnectWait(c.envCfg.ReconnectWait),
	)
	if err != nil {
		c.logger.Errorf("Failed to connect to %s server with error: %s", c.getBackendName(), err)
		return err
	}
	defer connection.Close()

	// configure the message sender
	var messageSenderToNats sender.GenericSender
	if c.jetstreamMode {
		messageSenderToNats = sender.NewJetstreamMessageSender(ctx, connection, c.envCfg, c.logger)
	} else {
		messageSenderToNats = sender.NewNatsMessageSender(ctx, connection, c.logger)
	}

	// cluster config
	k8sConfig := config.GetConfigOrDie()

	// setup application lister
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	applicationLister := application.NewLister(ctx, dynamicClient)

	// configure legacyTransformer
	legacyTransformer := legacy.NewTransformer(
		c.envCfg.ToConfig().BEBNamespace,
		c.envCfg.ToConfig().EventTypePrefix,
		applicationLister,
	)

	// configure Subscription Lister
	subDynamicSharedInfFactory := subscribed.GenerateSubscriptionInfFactory(k8sConfig)
	subLister := subDynamicSharedInfFactory.ForResource(subscribed.GVR).Lister()
	subscribedProcessor := &subscribed.Processor{
		SubscriptionLister: &subLister,
		Config:             c.envCfg.ToConfig(),
		Logger:             c.logger,
	}

	// sync informer cache or die
	c.logger.Info("Waiting for informers caches to sync")
	informers.WaitForCacheSyncOrDie(ctx, subDynamicSharedInfFactory)
	c.logger.Info("Informers are synced successfully")

	// configure event type cleaner
	eventTypeCleaner := eventtype.NewCleaner(c.envCfg.EventTypePrefix, applicationLister, c.logger)

	// start handler which blocks until it receives a shutdown signal
	if err := nats.NewHandler(messageReceiver, &messageSenderToNats, c.envCfg.RequestTimeout, legacyTransformer, c.opts,
		subscribedProcessor, c.logger, c.metricsCollector, eventTypeCleaner).Start(ctx); err != nil {
		c.logger.Errorf("Start handler failed with error: %s", err)
		return err
	}

	c.logger.Infof("Event Publisher %s shutdown", c.getBackendName())

	return nil
}

// Stop implements the Commander interface and stops the publisher.
func (c *Commander) Stop() error {
	c.cancel()
	return nil
}

func (c *Commander) getBackendName() string {
	if c.jetstreamMode {
		return "NATS in Jetstream mode"
	}
	return "NATS"
}
