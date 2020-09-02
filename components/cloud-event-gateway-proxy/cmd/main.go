package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/signals"
)

const (
	// emsPublishEndpoint the endpoint used to publish CloudEvents to EMS
	emsPublishEndpoint = "/events"
)

func main() {
	logger := logrus.New()

	env := &gateway.EnvConfig{}
	if err := envconfig.Process("", env); err != nil {
		logger.Fatalf("Start handler failed with error: %s", err)
	}

	logger.Info("Start the Cloudevent Gateway Proxy")

	// configure message receiver
	messageReceiver := receiver.NewHttpMessageReceiver(env.Port)

	// configure auth client
	ctx := signals.NewContext()
	client := oauth.NewClient(ctx, env)
	defer client.CloseIdleConnections()

	// configure message sender
	publishURL := fmt.Sprintf("%s%s", env.EmsCEURL, emsPublishEndpoint)
	messageSender, err := sender.NewHttpMessageSender(publishURL, client)
	if err != nil {
		logger.Fatalf("Unable to create message sender with error: %s", err)
	}

	// start handler which blocks until it receives a shutdown signal
	if err := handler.NewHandler(messageReceiver, messageSender, logger).Start(ctx); err != nil {
		logger.Fatalf("Start handler failed with error: %s", err)
	}

	logger.Info("Shutdown the Cloudevent Gateway Proxy")
}
