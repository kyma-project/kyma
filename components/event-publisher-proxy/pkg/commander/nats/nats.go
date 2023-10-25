package nats

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/builder"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	pkgnats "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender/jetstream"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/signals"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // TODO: remove as this is only required in a dev setup
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
}

// NewCommander creates the Commander for publisher to NATS.
func NewCommander(opts *options.Options, metricsCollector *metrics.Collector, logger *logger.Logger) *Commander {
	return &Commander{
		envCfg:           new(env.NATSConfig),
		logger:           logger,
		metricsCollector: metricsCollector,
		opts:             opts,
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
		pkgnats.WithName("Kyma Publisher"),
	)
	if err != nil {
		return xerrors.Errorf("failed to connect to backend server for %s : %v", natsCommanderName, err)
	}
	defer connection.Close()

	// configure the message sender
	messageSender := jetstream.NewSender(ctx, connection, c.envCfg, c.opts, c.logger)

	// cluster config
	k8sConfig := config.GetConfigOrDie()

	// setup application lister
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	applicationLister := application.NewLister(ctx, dynamicClient)

	// configure legacyTransformer
	legacyTransformer := legacy.NewTransformer(
		c.envCfg.ToConfig().EventMeshNamespace,
		c.envCfg.ToConfig().EventTypePrefix,
		applicationLister,
	)

	// configure Subscription Lister
	subDynamicSharedInfFactory := subscribed.GenerateSubscriptionInfFactory(k8sConfig)
	subLister := subDynamicSharedInfFactory.ForResource(subscribed.GVR).Lister()
	subscribedProcessor := &subscribed.Processor{
		SubscriptionLister: &subLister,
		Prefix:             c.envCfg.ToConfig().EventTypePrefix,
		Namespace:          c.envCfg.ToConfig().EventMeshNamespace,
		Logger:             c.logger,
	}

	// sync informer cache or die
	c.namedLogger().Info("Waiting for informers caches to sync")
	informers.WaitForCacheSyncOrDie(ctx, subDynamicSharedInfFactory, c.logger)
	c.namedLogger().Info("Informers are synced successfully")

	// configure event type cleaner
	eventTypeCleanerV1 := eventtype.NewCleaner(c.envCfg.EventTypePrefix, applicationLister, c.logger)

	// configure event type cleaner for subscription CRD v1alpha2
	eventTypeCleaner := cleaner.NewJetStreamCleaner(c.logger)

	// configure cloud event builder for subscription CRD v1alpha2
	ceBuilder := builder.NewGenericBuilder(env.JetStreamSubjectPrefix, eventTypeCleaner,
		applicationLister, c.logger)

	// start handler which blocks until it receives a shutdown signal
	h := handler.New(
		messageReceiver,
		messageSender,
		messageSender,
		c.envCfg.RequestTimeout,
		legacyTransformer,
		c.opts,
		subscribedProcessor,
		c.logger,
		c.metricsCollector,
		eventTypeCleanerV1,
		ceBuilder,
		c.envCfg.EventTypePrefix,
		env.JetStreamBackend,
	)
	if err := h.Start(ctx); err != nil {
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
	return c.logger.WithContext().Named(natsCommanderName).With("backend", natsBackend)
}
