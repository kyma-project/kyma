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
	Port              int    `envconfig:"HTTP_PORT" required:"true" default:"8080"`
}

func (e *envConfig) getApplicationSource() string {
	return e.ApplicationSource
}

type httpAdapter struct {
	ceClient       cloudevents.Client
	statsReporter  source.StatsReporter
	envConfig      *envConfig
	adapterContext context.Context
	logger         *zap.SugaredLogger
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
		logger:         logging.FromContext(ctx),
	}
}

// Start is the entrypoint for the adapter and is called by sharedmain coming from pkg/adapter
func (h *httpAdapter) Start(stopCh <-chan struct{}) error {
	logger := h.logger

	logger.Info("starting adapter")

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithPort(h.envConfig.Port),
		cloudevents.WithPath(path),
	)
	if err != nil {
		return fmt.Errorf("failed to create transport, %v", err)
	}
	decodeTransport, err := cloudevents.NewHTTPTransport()
	if err != nil {
		return fmt.Errorf("failed to create transport, %v", err)
	}
	c, err := cloudevents.NewClient(t, cloudevents.WithConverterFn(getConverterFunc(decodeTransport, h.envConfig.ApplicationSource)))
	if err != nil {
		return fmt.Errorf("failed to create client, %v", err)
	}

	log.Printf("will listen on :%d%s\n", h.envConfig.Port, path)
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(context.Background(), h.gotEvent))

	<-stopCh
	logger.Info("stopping adapter")
	return nil
}

// getConverterFunc returns a function which enriches the event with the application source
// the application source is forwarded to this adapter from the http controller which get's it forwarded from the application-operator
//func getConverterFunc(t *cloudeventshttp.Transport, applicationSource string) func(context.Context, transport.Message, error) (*cloudevents.Event, error) {
//	return func(ctx context.Context, m transport.Message, err error) (*cloudevents.Event, error) {
//		if msg, ok := m.(*cloudeventshttp.Message); ok {
//			event, err := t.MessageToEvent(ctx, msg)
//			if err != nil {
//				return nil, err
//			}
//			event.SetSource(applicationSource)
//			return event, nil
//		}
//		return nil, fmt.Errorf("expected message type to be http.Message but got %T", m)
//	}
//}

func (h *httpAdapter) gotEvent(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
	fmt.Printf("Got Event Context: %+v\n", event.Context)
	fmt.Printf("Got Transport Context: %+v\n", cloudevents.HTTPTransportContextFrom(ctx))

	fmt.Printf("----------------------------\n")


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
	rctx, revt, err := h.ceClient.Send(context.TODO(), event)
	rtctx := cloudevents.HTTPTransportContextFrom(rctx)
	if err != nil {
		h.logger.Error("failed to send cloudevent to sink", zap.Error(err), zap.Any("sink", h.envConfig.GetSinkURI()))
		resp.Error(http.StatusBadGateway, "unable to forward event to sink")
		return nil
	}

	// report the event
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
