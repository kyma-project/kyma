package main

import (
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/http"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/signals"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
)

func main() {
	logger := logrus.New()
	opts := options.ParseArgs()
	cfg := new(env.Config)
	if err := envconfig.Process("", cfg); err != nil {
		logger.Fatalf("Start handler failed with error: %s", err)
	}
	// configure message receiver
	messageReceiver := receiver.NewHttpMessageReceiver(cfg.Port)

	// configure auth client
	ctx := signals.NewContext()
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	// metrics server
	metricsServer := metrics.NewServer(logger)
	defer metricsServer.Stop()
	if err := metricsServer.Start(opts.MetricsAddress); err != nil {
		logger.Infof("Metrics server failed to start with error: %v", err)
	}

	// metrics collector
	metricsCollector := metrics.NewCollector()
	prometheus.MustRegister(metricsCollector)

	// configure message sender
	messageSender := sender.NewHttpMessageSender(cfg.EmsPublishURL, client)

	// cluster config
	k8sConfig := config.GetConfigOrDie()

	// setup application lister
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	applicationLister := application.NewLister(ctx, dynamicClient)

	// configure legacyTransformer
	legacyTransformer := legacy.NewTransformer(
		cfg.BEBNamespace,
		cfg.EventTypePrefix,
		applicationLister,
	)

	// Configure Subscription Lister
	subDynamicSharedInfFactory := subscribed.GenerateSubscriptionInfFactory(k8sConfig)
	subLister := subDynamicSharedInfFactory.ForResource(subscribed.GVR).Lister()
	subscribedProcessor := &subscribed.Processor{
		SubscriptionLister: &subLister,
		Config:             cfg,
		Logger:             logger,
	}
	// Sync informer cache or die
	logger.Info("Waiting for informers caches to sync")
	informers.WaitForCacheSyncOrDie(ctx, subDynamicSharedInfFactory)
	logger.Info("Informers are synced successfully")

	// start handler which blocks until it receives a shutdown signal
	if err := http.NewHandler(messageReceiver, messageSender, cfg.RequestTimeout, legacyTransformer, opts,
		subscribedProcessor, logger, metricsCollector).Start(ctx); err != nil {
		logger.Fatalf("Start handler failed with error: %s", err)
	}
	logger.Info("Shutdown the Event Publisher Proxy")
}
