package generic

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"github.com/cloudevents/sdk-go/v2/binding"
	cev2client "github.com/cloudevents/sdk-go/v2/client"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/tracing"
)

// EventingHandler is responsible for receiving HTTP requests and dispatching them to the EMS gateway.
// It also assures that the messages received are compliant with the Cloud Events spec.
type EventingHandler interface {
	Start(ctx context.Context) error
}
type Handler struct {
	Name string
	// Receiver receives incoming HTTP requests
	Receiver *receiver.HTTPMessageReceiver
	// Sender sends requests to the broker
	Sender        sender.GenericSender
	HealthChecker health.Checker
	// Defaulter sets default values to incoming events
	Defaulter cev2client.EventDefaulter
	// LegacyTransformer handles transformations needed to handle legacy events
	LegacyTransformer legacy.RequestToCETransformer
	// RequestTimeout timeout for outgoing requests
	RequestTimeout time.Duration
	// SubscribedProcessor processes requests for /:app/v1/events/subscribed endpoint
	SubscribedProcessor *subscribed.Processor
	// Logger default logger
	Logger *logger.Logger
	// Options configures HTTP server
	Options *options.Options
	// collector collects metrics
	collector metrics.PublishingMetricsCollector
	// eventTypeCleaner cleans the cloud event type
	eventTypeCleaner eventtype.Cleaner
	router           *mux.Router
}

// NewHandler returns a new HTTP Handler instance.
func NewHandler(receiver *receiver.HTTPMessageReceiver, sender sender.GenericSender, healthChecker health.Checker, requestTimeout time.Duration,
	legacyTransformer legacy.RequestToCETransformer, opts *options.Options, subscribedProcessor *subscribed.Processor,
	logger *logger.Logger, collector metrics.PublishingMetricsCollector, eventTypeCleaner eventtype.Cleaner) *Handler {
	return &Handler{
		Receiver:            receiver,
		Sender:              sender,
		HealthChecker:       healthChecker,
		RequestTimeout:      requestTimeout,
		LegacyTransformer:   legacyTransformer,
		SubscribedProcessor: subscribedProcessor,
		Logger:              logger,
		Options:             opts,
		collector:           collector,
		eventTypeCleaner:    eventTypeCleaner,
	}
}

func (h *Handler) setupMux() {
	router := mux.NewRouter()
	router.HandleFunc(handler.PublishEndpoint, h.maxBytes(h.publishCloudEvents)).Methods(http.MethodPost)
	router.HandleFunc(handler.LegacyEndpointPattern, h.maxBytes(h.publishLegacyEventsAsCE)).Methods(http.MethodPost)
	router.HandleFunc(handler.SubscribedEndpointPattern, h.maxBytes(h.SubscribedProcessor.ExtractEventsFromSubscriptions)).Methods(http.MethodGet)
	router.HandleFunc(health.ReadinessURI, h.maxBytes(h.HealthChecker.ReadinessCheck))
	router.HandleFunc(health.LivenessURI, h.maxBytes(h.HealthChecker.LivenessCheck))
	h.router = router
}

// Start starts the Handler with the given context.
func (h *Handler) Start(ctx context.Context) error {
	h.setupMux()
	return h.Receiver.StartListen(ctx, h.router, h.Logger)
}

func (h *Handler) maxBytes(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, h.Options.MaxRequestSize)
		f(w, r)
	}
}

func (h *Handler) publishLegacyEventsAsCE(writer http.ResponseWriter, request *http.Request) {
	event, _ := h.LegacyTransformer.TransformLegacyRequestsToCE(writer, request)
	if event == nil {
		h.namedLogger().Debug("Failed to transform legacy event to CloudEvent, event is nil")
		return
	}
	ctx := request.Context()

	result, err := h.sendEventAndRecordMetrics(ctx, event, request.URL.Host, request.Header)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		h.namedLogger().With().Error(err)
		return
	}

	// Change response as per old error codes
	h.LegacyTransformer.TransformsCEResponseToLegacyResponse(writer, result.HTTPStatus(), event, string(result.ResponseBody()))
}

func (h *Handler) publishCloudEvents(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	event, err := extractCloudEventFromRequest(request)
	if err != nil {
		h.namedLogger().With().Error(err)
		e := writeResponse(writer, http.StatusBadRequest, []byte(err.Error()))
		if e != nil {
			h.namedLogger().Error(e)
		}
		return
	}

	eventTypeOriginal := event.Type()
	eventTypeClean, err := h.eventTypeCleaner.Clean(eventTypeOriginal)
	if err != nil {
		h.namedLogger().Error(err)
		e := writeResponse(writer, http.StatusBadRequest, []byte(err.Error()))
		if e != nil {
			h.namedLogger().Error(e)
		}
		return
	}
	event.SetType(eventTypeClean)

	result, err := h.sendEventAndRecordMetrics(ctx, event, request.URL.Host, request.Header)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		h.namedLogger().With().Error(err)
		return
	}

	err = writeResponse(writer, result.HTTPStatus(), result.ResponseBody())
	if err != nil {
		h.namedLogger().With().Error(err)
	}
}

func extractCloudEventFromRequest(request *http.Request) (*cev2event.Event, error) {
	message := cev2http.NewMessageFromHttpRequest(request)
	defer func() { _ = message.Finish(nil) }()

	event, err := binding.ToEvent(context.Background(), message)
	if err != nil {
		return nil, err
	}

	err = event.Validate()
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (h *Handler) sendEventAndRecordMetrics(ctx context.Context, event *cev2event.Event, host string, header http.Header) (sender.PublishResult, error) {
	ctx, cancel := context.WithTimeout(ctx, h.RequestTimeout)
	defer cancel()
	h.applyDefaults(ctx, event)
	tracing.AddTracingContextToCEExtensions(header, event)
	start := time.Now()
	result, err := h.Sender.Send(ctx, event)
	duration := time.Since(start)
	if err != nil {
		h.collector.RecordError()
		return nil, err
	}
	h.collector.RecordEventType(event.Type(), event.Source(), result.HTTPStatus())
	h.collector.RecordLatency(duration, result.HTTPStatus(), host)
	h.collector.RecordRequests(result.HTTPStatus(), host)
	return result, nil
}

// writeResponse writes the HTTP response given the status code and response body.
func writeResponse(writer http.ResponseWriter, statusCode int, respBody []byte) error {
	writer.WriteHeader(statusCode)

	if respBody == nil {
		return nil
	}
	_, err := writer.Write(respBody)
	return err
}

// applyDefaults applies the default values (if any) to the given Cloud Event.
func (h *Handler) applyDefaults(ctx context.Context, event *cev2event.Event) {
	if h.Defaulter != nil {
		newEvent := h.Defaulter(ctx, *event)
		*event = newEvent
	}
}

func (h *Handler) namedLogger() *zap.SugaredLogger {
	return h.Logger.WithContext().Named(h.Name)
}
