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
	ceBuilder     builder.CloudEventBuilder
	router        *mux.Router
	activeBackend string
}

const (
	originalTypeHeaderName = "ce-original-type"
)

// NewHandler returns a new HTTP Handler instance.
func NewHandler(receiver *receiver.HTTPMessageReceiver, sender sender.GenericSender, healthChecker health.Checker, requestTimeout time.Duration,
	legacyTransformer legacy.RequestToCETransformer, opts *options.Options, subscribedProcessor *subscribed.Processor,
	logger *logger.Logger, collector metrics.PublishingMetricsCollector, eventTypeCleaner eventtype.Cleaner,
	ceBuilder builder.CloudEventBuilder, activeBackend string) *Handler {

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
		activeBackend:       activeBackend,
	}
}

// setupMux configures the request router for all required endpoints.
func (h *Handler) setupMux() {
	router := mux.NewRouter()
	router.Use(h.collector.MetricsMiddleware())
	router.HandleFunc(PublishEndpoint, h.maxBytes(h.publishCloudEvents)).Methods(http.MethodPost)
	router.HandleFunc(LegacyEndpointPattern, h.maxBytes(h.publishLegacyEventsAsCE)).Methods(http.MethodPost)
	if h.Options.EnableNewCRDVersion {
		router.HandleFunc(
			SubscribedEndpointPattern,
			h.maxBytes(h.SubscribedProcessor.ExtractEventsFromSubscriptions)).Methods(http.MethodGet)
	} else {
		router.HandleFunc(
			SubscribedEndpointPattern,
			h.maxBytes(h.SubscribedProcessor.ExtractEventsFromSubscriptionsV1alpha1)).Methods(http.MethodGet)
	}
	router.HandleFunc(health.ReadinessURI, h.maxBytes(h.HealthChecker.ReadinessCheck))
	router.HandleFunc(health.LivenessURI, h.maxBytes(h.HealthChecker.LivenessCheck))
	h.router = router
}

// Start starts the Handler with the given context.
func (h *Handler) Start(ctx context.Context) error {
	h.setupMux()
	return h.Receiver.StartListen(ctx, h.router, h.Logger)
}

// maxBytes installs a MaxBytesReader onto the request, so that incoming request larger than a given size will cause an error.
func (h *Handler) maxBytes(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, h.Options.MaxRequestSize)
		f(w, r)
	}
}

func (h *Handler) handlePublishLegacyEventV1alpha2(writer http.ResponseWriter, publishData *api.PublishRequestData, request *http.Request) (sender.PublishResult, *cev2event.Event) {
	ceEvent, errResp, _ := h.LegacyTransformer.ExtractCEFromLegacyPublishRequestData(publishData)
	if errResp != nil {
		legacy.WriteJSONResponse(writer, errResp)
		return nil, nil
	}

	// set original type header
	ceEvent.SetExtension(originalTypeHeaderName, ceEvent.Type())

	// build a new cloud event instance as per specifications per backend
	event, err := h.ceBuilder.Build(*ceEvent)
	if err != nil {
		legacy.WriteJSONResponse(writer, legacy.ErrorResponseBadRequest(err.Error()))
		return nil, nil
	}

	result, err := h.sendEventAndRecordMetrics(request.Context(), event, h.Sender.URL(), request.Header)
	if err != nil {
		h.namedLogger().With().Error(err)
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, sender.ErrInsufficientStorage) {
			httpStatus = http.StatusInsufficientStorage
		} else if errors.Is(err, sender.ErrBackendTargetNotFound) {
			httpStatus = http.StatusBadGateway
		}
		h.LegacyTransformer.TransformsCEResponseToLegacyResponse(writer, httpStatus, event, err.Error())
		return nil, nil
	}
	h.namedLogger().With().Debug(result)

	return result, event
}

func (h *Handler) handlePublishLegacyEventV1alpha1(writer http.ResponseWriter, publishData *api.PublishRequestData, request *http.Request) (sender.PublishResult, *cev2event.Event) {
	event, _ := h.LegacyTransformer.TransformLegacyRequestsToCE(writer, publishData)
	if event == nil {
		h.namedLogger().Error("Failed to transform legacy event to CloudEvent, event is nil")
		return nil, nil
	}

	result, err := h.sendEventAndRecordMetrics(request.Context(), event, h.Sender.URL(), request.Header)
	if err != nil {
		h.namedLogger().With().Error(err)
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, sender.ErrInsufficientStorage) {
			httpStatus = http.StatusInsufficientStorage
		} else if errors.Is(err, sender.ErrBackendTargetNotFound) {
			httpStatus = http.StatusBadGateway
		}
		h.LegacyTransformer.TransformsCEResponseToLegacyResponse(writer, httpStatus, event, err.Error())
		return nil, nil
	}
	h.namedLogger().With().Debug(result)

	return result, event
}

// publishLegacyEventsAsCE converts an incoming request in legacy event format to a cloudevent and dispatches it using
// the configured GenericSender.
func (h *Handler) publishLegacyEventsAsCE(writer http.ResponseWriter, request *http.Request) {
	var publishedEvent *cev2event.Event
	var successResult sender.PublishResult

	// extract publish data from request
	publishRequestData, errResp, _ := h.LegacyTransformer.ExtractPublishRequestData(request)
	if errResp != nil {
		legacy.WriteJSONResponse(writer, errResp)
		return
	}

	// publish event for Subscription v1alpha2
	if h.Options.EnableNewCRDVersion {
		successResult, publishedEvent = h.handlePublishLegacyEventV1alpha2(writer, publishRequestData, request)
		// if publishedEvent is nil, then it means that the publishing failed
		// and the response is already returned to the user
		if publishedEvent == nil {
			return
		}
	}

	// publish event for Subscription v1alpha1
	// In case of EnableNewCRDVersion is true and the active backend is JetStream
	// then we will publish event on both possible subjects
	// i.e. with prefix (`sap.kyma.custom`) and without prefix
	// this behaviour will be depreciated when we remove support for JetStream with Subscription `exact` typeMatching
	if !h.Options.EnableNewCRDVersion || h.activeBackend == env.JetStreamBackend {
		successResult, publishedEvent = h.handlePublishLegacyEventV1alpha1(writer, publishRequestData, request)
		// if publishedEvent is nil, then it means that the publishing failed
		// and the response is already returned to the user
		if publishedEvent == nil {
			return
		}
	}

	// return success response to user
	// change response as per old error codes
	h.LegacyTransformer.TransformsCEResponseToLegacyResponse(writer, successResult.HTTPStatus(), publishedEvent, string(successResult.ResponseBody()))
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
	event.SetExtension(originalTypeHeaderName, eventTypeOriginal)

	if h.Options.EnableNewCRDVersion && !strings.HasPrefix(eventTypeOriginal, env.OldEventTypePrefix) {
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

	result, err := h.sendEventAndRecordMetrics(ctx, event, h.Sender.URL(), request.Header)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, sender.ErrInsufficientStorage) {
			httpStatus = http.StatusInsufficientStorage
		}
		writer.WriteHeader(httpStatus)
		h.namedLogger().With().Error(err)
		return
	}
	h.namedLogger().With().Debug(result)

	err = writeResponse(writer, result.HTTPStatus(), result.ResponseBody())
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
func (h *Handler) sendEventAndRecordMetrics(ctx context.Context, event *cev2event.Event, host string, header http.Header) (sender.PublishResult, error) {
	ctx, cancel := context.WithTimeout(ctx, h.RequestTimeout)
	defer cancel()
	h.applyDefaults(ctx, event)
	tracing.AddTracingContextToCEExtensions(header, event)
	start := time.Now()
	result, err := h.Sender.Send(ctx, event)
	duration := time.Since(start)
	if err != nil {
		h.collector.RecordBackendError()
		return nil, err
	}
	h.collector.RecordEventType(event.Type(), event.Source(), result.HTTPStatus())
	h.collector.RecordBackendLatency(duration, result.HTTPStatus(), host)
	h.collector.RecordBackendRequests(result.HTTPStatus(), host)
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
