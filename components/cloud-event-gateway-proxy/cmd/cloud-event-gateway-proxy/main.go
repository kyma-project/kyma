package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/env"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/signals"
)

func main() {
	logger := logrus.New()

	cfg := new(env.Config)
	if err := envconfig.Process("", cfg); err != nil {
		logger.Fatalf("Start handler failed with error: %s", err)
	}

	logger.Info("Start the Cloudevent Gateway Proxy")

	// configure message receiver
	messageReceiver := receiver.NewHttpMessageReceiver(cfg.Port)

	// configure auth client
	ctx := signals.NewContext()
	client := oauth.NewClient(ctx, cfg)
	defer client.CloseIdleConnections()

	// configure message sender
	messageSender := sender.NewHttpMessageSender(cfg.EmsPublishURL, client)

	// start handler which blocks until it receives a shutdown signal
	if err := handler.NewHandler(messageReceiver, messageSender, cfg.RequestTimeout, logger).Start(ctx); err != nil {
		logger.Fatalf("Start handler failed with error: %s", err)
	}

	logger.Info("Shutdown the Cloudevent Gateway Proxy")
}
