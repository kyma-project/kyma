package nats

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/xerrors"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // TODO: remove as this is only required in a dev setup
	"sigs.k8s.io/controller-runtime/pkg/client/config"

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

const (
	natsBackend       = "nats"
	natsCommanderName = natsBackend + "-commander"
)

// Commander implements the Commander interface.
type Commander struct {
	cancel           context.CancelFunc
	metricsCollector *metrics.Collector
	logger           *logger.Logger
	envCfg           *env.NATSConfig
	opts             *options.Options
	jetstreamMode    bool
}

// NewCommander creates the Commander for publisher to NATS.
func NewCommander(opts *options.Options, metricsCollector *metrics.Collector, logger *logger.Logger, jetstreamMode bool) *Commander {
	return &Commander{
		envCfg:           new(env.NATSConfig),
		logger:           logger,
		metricsCollector: metricsCollector,
		opts:             opts,
		jetstreamMode:    jetstreamMode,
	}
}

// Init implements the Commander interface and initializes the publisher to NATS.
func (c *Commander) Init() error {
	if err := envconfig.Process("", c.envCfg); err != nil {
		return xerrors.Errorf("failed to read configuration for %s : %v", natsCommanderName, err)
	}
	return nil
}

// Start implements the Commander interface and starts the publisher.
func (c *Commander) Start() error {
	c.namedLogger().Infow("Starting Event Publisher", "configuration", c.envCfg.String(), "startup arguments", c.opts)

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
		return xerrors.Errorf("failed to connect to backend server for %s : %v", natsCommanderName, err)
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
	c.namedLogger().Info("Waiting for informers caches to sync")
	informers.WaitForCacheSyncOrDie(ctx, subDynamicSharedInfFactory, c.logger)
	c.namedLogger().Info("Informers are synced successfully")

	// configure event type cleaner
	eventTypeCleaner := eventtype.NewCleaner(c.envCfg.EventTypePrefix, applicationLister, c.logger)

	// start handler which blocks until it receives a shutdown signal
	if err := nats.NewHandler(messageReceiver, &messageSenderToNats, c.envCfg.RequestTimeout, legacyTransformer, c.opts,
		subscribedProcessor, c.logger, c.metricsCollector, eventTypeCleaner).Start(ctx); err != nil {
		return xerrors.Errorf("failed to start handler for %s : %v", natsCommanderName, err)
	}

	c.namedLogger().Infof("Event Publisher was shut down")

	return nil
}

// Stop implements the Commander interface and stops the publisher.
func (c *Commander) Stop() error {
	c.cancel()
	return nil
}

func (c *Commander) namedLogger() *zap.SugaredLogger {
	return c.logger.WithContext().Named(natsCommanderName).With("backend", natsBackend, "jestream mode", c.jetstreamMode)
}
