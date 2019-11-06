package http

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/components/event-sources/apis/sources"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/source"
	"log"
	"net/http"
)

type envConfig struct {
	adapter.EnvConfig
	ApplicationSource string `envconfig:"APPLICATION_SOURCE" required:"true"`

	// PORT as required by knative serving runtime contract
	Port              int    `envconfig:"PORT" required:"true" default:"8080"`
}

func (e *envConfig) getApplicationSource() string {
	return e.ApplicationSource
}

type httpAdapter struct {
	ceClient       cloudevents.Client
	statsReporter  source.StatsReporter
	envConfig      *envConfig
	adapterContext context.Context
	logger         *zap.Logger
}

const resourceGroup = "http." + sources.GroupName
const path = "/"

func NewEnvConfig() adapter.EnvConfigAccessor {
	return &envConfig{}
}

func NewAdapter(ctx context.Context, processed adapter.EnvConfigAccessor, ceClient cloudevents.Client, reporter source.StatsReporter) adapter.Adapter {

	envConfig, ok := processed.(*envConfig)
	if !ok {
		panic(fmt.Sprintf("cannot create adapter, expecting a *envconfig, but got a %T", processed))
	}

	return &httpAdapter{
		adapterContext: ctx,
		ceClient:       ceClient,
		statsReporter:  reporter,
		envConfig:      envConfig,
		logger:         logging.FromContext(ctx).Desugar(),
	}
}

// Start is the entrypoint for the adapter and is called by sharedmain coming from pkg/adapter
func (h *httpAdapter) Start(stopCh <-chan struct{}) error {
	logger := h.logger

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithPort(h.envConfig.Port),
		cloudevents.WithPath(path),
	)
	if err != nil {
		return fmt.Errorf("failed to create transport, %v", err)
	}
	c, err := cloudevents.NewClient(t)
	if err != nil {
		return fmt.Errorf("failed to create client, %v", err)
	}

	log.Printf("will listen on :%d%s\n", h.envConfig.Port, path)
	fmt.Printf("client: %v", c)
	if err := c.StartReceiver(h.adapterContext, h.serveHTTP); err != nil {
		log.Fatalf("failed to start receiver: %v", err)
	}

	<-stopCh
	logger.Info("stopping adapter")
	return nil
}

// isSupportedCloudEvent determines if an incoming cloud event is accepted
func isSupportedCloudEvent(event cloudevents.Event) bool {
	eventVersion := event.Context.GetSpecVersion()
	return eventVersion != cloudevents.VersionV01 && eventVersion != cloudevents.VersionV02 && eventVersion != cloudevents.VersionV03
}

func (h *httpAdapter) serveHTTP(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
	logger := h.logger
	logger.Debug("got event", zap.Any("event_context", event.Context))

	if !isSupportedCloudEvent(event) {
		resp.Error(http.StatusBadRequest, "unsupported cloudevents version")
		return nil
	}

	// enrich the event with the application source
	// the application source is forwarded to this adapter from the http controller which get's it forwarded from the application-operator
	event.SetSource(h.envConfig.ApplicationSource)

	reportArgs := &source.ReportArgs{
		Namespace:     h.envConfig.GetNamespace(),
		EventSource:   event.Source(),
		EventType:     event.Type(),
		Name:          "http_adapter",
		ResourceGroup: resourceGroup,
	}

	// TODO(nachtmaar): forward event to resp.RespondWith ??
	logger.Debug("sending event", zap.Any("sink", h.envConfig.GetSinkURI()))
	rctx, revt, err := h.ceClient.Send(ctx, event)
	rtctx := cloudevents.HTTPTransportContextFrom(rctx)
	if err != nil {
		h.logger.Error("failed to send cloudevent to sink", zap.Error(err), zap.Any("sink", h.envConfig.GetSinkURI()))
		resp.Error(http.StatusBadGateway, "unable to forward event to sink")
		return nil
	}

	// report a sent event
	if err := h.statsReporter.ReportEventCount(reportArgs, rtctx.StatusCode); err != nil {
		h.logger.Warn("cannot report event count", zap.Error(err))
	}

	if rtctx.StatusCode/100 == 2 {
		resp.RespondWith(http.StatusOK, revt)
		return nil
	}

	if rtctx.StatusCode/100 == 4 {
		resp.RespondWith(rtctx.StatusCode, revt)
		return nil
	}

	resp.RespondWith(http.StatusInternalServerError, revt)
	return nil
}
