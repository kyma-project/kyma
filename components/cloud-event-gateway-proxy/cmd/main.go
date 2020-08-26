package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"go.opencensus.io/stats/view"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/oauth"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/signals"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

const (
	// Constants for the underlying HTTP Client transport. These would enable better connection reuse.
	// TODO: figure out the magic numbers
	defaultMaxIdleConnections        = 1000
	defaultMaxIdleConnectionsPerHost = 1000
	//defaultTTL                       int32 = 255
	//defaultMetricsPort                     = 9090
	component             = "cloud_event_gateway_proxy"
	emsPublishingEndpoint = "/events"
)

func main() {
	flag.Parse()

	ctx := signals.NewContext()

	// Report stats on Go memory usage every 30 seconds.
	msp := metrics.NewMemStatsAll()
	msp.Start(ctx, 30*time.Second)
	if err := view.Register(msp.DefaultViews()...); err != nil {
		log.Fatalf("failed to exports go memstats view: %v", err)
	}

	var env gateway.EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("failed to process env var: %v", err)
	}

	publishURL := fmt.Sprintf("%s%s", env.EmsCEURL, emsPublishingEndpoint)

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	logger.Info("Starting the Cloud Event Gateway Proxy")

	oAuth2Config := oauth.Config(env)

	httpClient := oAuth2Config.Client(ctx)
	connectionArgs := sender.ConnectionArgs{
		MaxIdleConns:        defaultMaxIdleConnections,
		MaxIdleConnsPerHost: defaultMaxIdleConnectionsPerHost,
	}
	messageSender, err := sender.NewHttpMessageSender(&connectionArgs, publishURL, httpClient)
	if err != nil {
		logger.Fatal("Unable to create message sender", zap.Error(err))
	}

	h := &handler.Handler{
		Receiver: receiver.NewHttpMessageReceiver(env.Port),
		Sender:   messageSender,
		//Reporter:     reporter,
		Logger: logger,
	}

	// Start blocks forever.
	if err = h.Start(ctx); err != nil {
		logger.Error("Start() returned an error", zap.Error(err))
	}

	logger.Info("Exiting...")
}
