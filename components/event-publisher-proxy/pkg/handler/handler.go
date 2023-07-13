package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/api"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"github.com/cloudevents/sdk-go/v2/binding"
	cev2client "github.com/cloudevents/sdk-go/v2/client"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/builder"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/tracing"
)

// EventingHandler is responsible for receiving HTTP requests and dispatching them to the Backend.
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
	// builds the cloud event according to Subscription v1alpha2 specifications
	ceBuilder          builder.CloudEventBuilder
	router             *mux.Router
	activeBackend      env.ActiveBackend
	OldEventTypePrefix string
}

// NewHandler returns a new HTTP Handler instance.
func NewHandler(receiver *receiver.HTTPMessageReceiver, sender sender.GenericSender, healthChecker health.Checker,
	requestTimeout time.Duration, legacyTransformer legacy.RequestToCETransformer, opts *options.Options,
	subscribedProcessor *subscribed.Processor, logger *logger.Logger, collector metrics.PublishingMetricsCollector,
	eventTypeCleaner eventtype.Cleaner, ceBuilder builder.CloudEventBuilder, oldEventTypePrefix string,
	activeBackend env.ActiveBackend) *Handler {
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
		ceBuilder:           ceBuilder,
		OldEventTypePrefix:  oldEventTypePrefix,
		activeBackend:       activeBackend,
	}
}

// setupMux configures the request router for all required endpoints.
func (h *Handler) setupMux() {
	router := mux.NewRouter()
	router.Use(h.collector.MetricsMiddleware())
	router.HandleFunc(PublishEndpoint, h.maxBytes(h.publishCloudEvents)).Methods(http.MethodPost)
	router.HandleFunc(LegacyEndpointPattern, h.maxBytes(h.publishLegacyEventsAsCE)).Methods(http.MethodPost)
	router.HandleFunc(
		SubscribedEndpointPattern,
		h.maxBytes(h.SubscribedProcessor.ExtractEventsFromSubscriptions)).Methods(http.MethodGet)
	router.HandleFunc(health.ReadinessURI, h.maxBytes(h.HealthChecker.ReadinessCheck))
	router.HandleFunc(health.LivenessURI, h.maxBytes(h.HealthChecker.LivenessCheck))
	h.router = router
}

// Start starts the Handler with the given context.
func (h *Handler) Start(ctx context.Context) error {
	h.setupMux()
	return h.Receiver.StartListen(ctx, h.router, h.Logger)
}

// maxBytes installs a MaxBytesReader onto the request, so that incoming request that is larger than a given size
// will cause an error.
func (h *Handler) maxBytes(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, h.Options.MaxRequestSize)
		f(w, r)
	}
}

// handleSendEventAndRecordMetricsLegacy handles the publishing of metrics.
// It writes to the user request if any error occurs.
// Otherwise, returns the result.
func (h *Handler) handleSendEventAndRecordMetricsLegacy(
	writer http.ResponseWriter, request *http.Request, event *cev2event.Event) error {
	err := h.sendEventAndRecordMetrics(request.Context(), event, h.Sender.URL(), request.Header)
	if err != nil {
		h.namedLogger().Error(err)
		httpStatus := http.StatusInternalServerError
		var pubErr sender.PublishError
		if ok := errors.As(err, &pubErr); ok {
			httpStatus = pubErr.Code()
		}
		h.LegacyTransformer.WriteCEResponseAsLegacyResponse(writer, httpStatus, event, err.Error())
		return err
	}
	return nil
}

// handlePublishLegacyEvent handles the publishing of events for Subscription v1alpha2 CRD.
// It writes to the user request if any error occurs.
// Otherwise, return the published event.
func (h *Handler) handlePublishLegacyEvent(writer http.ResponseWriter, publishData *api.PublishRequestData, request *http.Request) (*cev2event.Event, error) {
	ceEvent, err := h.LegacyTransformer.TransformPublishRequestToCloudEvent(publishData)
	if err != nil {
		legacy.WriteJSONResponse(writer, legacy.ErrorResponse(http.StatusInternalServerError, err))
		return nil, nil
	}

	// build a new cloud event instance as per specifications per backend
	event, err := h.ceBuilder.Build(*ceEvent)
	if err != nil {
		legacy.WriteJSONResponse(writer, legacy.ErrorResponseBadRequest(err.Error()))
		return nil, err
	}

	err = h.handleSendEventAndRecordMetricsLegacy(writer, request, event)
	if err != nil {
		return nil, err
	}

	return event, err
}

// handlePublishLegacyEventV1alpha1 handles the publishing of events for Subscription v1alpha1 CRD.
// It writes to the user request if any error occurs.
// Otherwise, return the published event.
func (h *Handler) handlePublishLegacyEventV1alpha1(writer http.ResponseWriter, publishData *api.PublishRequestData, request *http.Request) (*cev2event.Event, error) {
	event, _ := h.LegacyTransformer.WriteLegacyRequestsToCE(writer, publishData)
	if event == nil {
		h.namedLogger().Error("Failed to transform legacy event to CloudEvent, event is nil")
		return nil, nil
	}

	err := h.handleSendEventAndRecordMetricsLegacy(writer, request, event)
	if err != nil {
		return nil, err
	}

	return event, err
}

// publishLegacyEventsAsCE converts an incoming request in legacy event format to a cloudevent and dispatches it using
// the configured GenericSender.
func (h *Handler) publishLegacyEventsAsCE(writer http.ResponseWriter, request *http.Request) {
	// extract publish data from request
	publishRequestData, errResp, _ := h.LegacyTransformer.ExtractPublishRequestData(request)
	if errResp != nil {
		legacy.WriteJSONResponse(writer, errResp)
		return
	}

	// publish event for Subscription
	publishedEvent, err := h.handlePublishLegacyEvent(writer, publishRequestData, request)
	// if publishedEvent is nil, then it means that the publishing failed
	// and the response is already returned to the user
	if err != nil {
		return
	}

	// publish event for Subscription v1alpha1
	// In case: the active backend is JetStream
	// then we will publish event on both possible subjects
	// i.e. with prefix (`sap.kyma.custom`) and without prefix
	// this behaviour will be deprecated when we remove support for JetStream with Subscription `exact` typeMatching
	if h.activeBackend == env.JetStreamBackend {
		publishedEvent, err = h.handlePublishLegacyEventV1alpha1(writer, publishRequestData, request)
		// if publishedEvent is nil, then it means that the publishing failed
		// and the response is already returned to the user
		if err != nil {
			return
		}
	}

	// return success response to user
	// change response as per old error codes
	h.LegacyTransformer.WriteCEResponseAsLegacyResponse(writer, http.StatusNoContent, publishedEvent, "")
}

// publishCloudEvents validates an incoming cloudevent and dispatches it using
// the configured GenericSender.
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

	//nolint:nestif // it will be improved when v1alpha1 is deprecated.
	if !strings.HasPrefix(eventTypeOriginal, h.OldEventTypePrefix) {
		// build a new cloud event instance as per specifications per backend
		event, err = h.ceBuilder.Build(*event)
		if err != nil {
			e := writeResponse(writer, http.StatusBadRequest, []byte(err.Error()))
			if e != nil {
				h.namedLogger().Error(e)
			}
			return
		}
	} else {
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
	}

	err = h.sendEventAndRecordMetrics(ctx, event, h.Sender.URL(), request.Header)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		var pubErr sender.PublishError
		if ok := errors.As(err, &pubErr); ok {
			httpStatus = pubErr.Code()
		}
		writer.WriteHeader(httpStatus)
		h.namedLogger().With().Error(err)
		return
	}
	err = writeResponse(writer, http.StatusNoContent, []byte(""))
	if err != nil {
		h.namedLogger().With().Error(err)
	}
}

// extractCloudEventFromRequest converts an incoming CloudEvent request to an Event.
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

// sendEventAndRecordMetrics dispatches an Event and records metrics based on dispatch success.
func (h *Handler) sendEventAndRecordMetrics(ctx context.Context, event *cev2event.Event, host string, header http.Header) error {
	ctx, cancel := context.WithTimeout(ctx, h.RequestTimeout)
	defer cancel()
	h.applyDefaults(ctx, event)
	tracing.AddTracingContextToCEExtensions(header, event)
	start := time.Now()
	err := h.Sender.Send(ctx, event)
	duration := time.Since(start)
	if err != nil {
		var pubErr sender.PublishError
		code := 500
		if ok := errors.As(err, &pubErr); ok {
			code = pubErr.Code()
		}
		h.collector.RecordBackendLatency(duration, code, host)
		h.collector.RecordBackendRequests(code, host)
		h.collector.RecordBackendError()
		return err
	}
	originalEventType := event.Type()
	originalTypeHeader, ok := event.Extensions()[builder.OriginalTypeHeaderName]
	if !ok {
		h.namedLogger().With().Debugw("event header doesn't exist", "header",
			builder.OriginalTypeHeaderName)
	} else {
		originalEventType, ok = originalTypeHeader.(string)
		if !ok {
			h.namedLogger().With().Warnw("failed to convert event original event type extension value to string",
				builder.OriginalTypeHeaderName, originalTypeHeader)
			originalEventType = event.Type()
		}
	}
	h.collector.RecordEventType(originalEventType, event.Source(), http.StatusNoContent)
	h.collector.RecordBackendLatency(duration, http.StatusNoContent, host)
	h.collector.RecordBackendRequests(http.StatusNoContent, host)
	return nil
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
