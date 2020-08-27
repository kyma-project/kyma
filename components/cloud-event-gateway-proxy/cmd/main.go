package main

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/signals"
)

const (
	// Constants for the underlying HTTP Client transport. These would enable better connection reuse.
	defaultMaxIdleConns        = 1000
	defaultMaxIdleConnsPerHost = 1000

	// emsPublishEndpoint the endpoint used to publish cloudevents
	emsPublishEndpoint = "/events"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger with error: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	var env gateway.EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		logger.Fatal("Start handler failed", zap.Error(err))
	}

	logger.Info("Start the Cloudevent Gateway Proxy")

	ctx := signals.NewContext()
	authCfg := oauth.Config(env)
	httpClient := authCfg.Client(ctx)
	publishURL := fmt.Sprintf("%s%s", env.EmsCEURL, emsPublishEndpoint)
	connectionArgs := sender.ConnectionArgs{MaxIdleConns: defaultMaxIdleConns, MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost}
	messageSender, err := sender.NewHttpMessageSender(&connectionArgs, publishURL, httpClient)
	if err != nil {
		logger.Fatal("Unable to create message sender", zap.Error(err))
	}

	// start handler which blocks until it receives a shutdown signal
	h := handler.NewHandler(receiver.NewHttpMessageReceiver(env.Port), messageSender, logger)
	if err = h.Start(ctx); err != nil {
		logger.Fatal("Start handler failed", zap.Error(err))
	}

	logger.Info("Shutdown the Cloudevent Gateway Proxy")
}
