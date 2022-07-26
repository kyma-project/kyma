package beb

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"go.uber.org/zap"

	"github.com/cloudevents/sdk-go/v2/binding"
	cev2client "github.com/cloudevents/sdk-go/v2/client"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cev2http "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/internal"
	cloudevents "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/ems"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"
)

const (
	// noDuration signals that the dispatch step has not started yet.
	noDuration = -1

	bebCommanderName      = "beb-commander"
	destinationServiceBEB = "beb"
)

var (
	// additionalHeaders are the required headers by EMS for publish requests.
	// Any alteration or removal of those headers might cause publish requests to fail.
	additionalHeaders = http.Header{
		"qos":    []string{string(ems.QosAtLeastOnce)},
		"Accept": []string{internal.ContentTypeApplicationJSON},
	}
)

// Handler is responsible for receiving HTTP requests and dispatching them to the EMS gateway.
// It also assures that the messages received are compliant with the Cloud Events spec.
type Handler struct {
	// Receiver receives incoming HTTP requests
	Receiver *receiver.HTTPMessageReceiver
	// Sender sends requests to the broker
	Sender *sender.BebMessageSender
	// Defaulter sets default values to incoming events
	Defaulter cev2client.EventDefaulter
	// LegacyTransformer handles transformations needed to handle legacy events
	LegacyTransformer *legacy.Transformer
	// RequestTimeout timeout for outgoing requests
	RequestTimeout time.Duration
	//SubscribedProcessor processes requests for /:app/v1/events/subscribed endpoint
	SubscribedProcessor *subscribed.Processor
	// Logger default logger
	Logger *logger.Logger
	// Options configures HTTP server
	Options *options.Options
	// collector collects metrics
	collector *metrics.Collector
	// eventTypeCleaner cleans the cloud event type
	eventTypeCleaner eventtype.Cleaner
}

// NewHandler returns a new HTTP Handler instance.
func NewHandler(receiver *receiver.HTTPMessageReceiver, sender *sender.BebMessageSender, requestTimeout time.Duration,
	legacyTransformer *legacy.Transformer, opts *options.Options, subscribedProcessor *subscribed.Processor,
	logger *logger.Logger, collector *metrics.Collector, eventTypeCleaner eventtype.Cleaner) *Handler {
	return &Handler{
		Receiver:            receiver,
		Sender:              sender,
		RequestTimeout:      requestTimeout,
		LegacyTransformer:   legacyTransformer,
		SubscribedProcessor: subscribedProcessor,
		Logger:              logger,
		Options:             opts,
		collector:           collector,
		eventTypeCleaner:    eventTypeCleaner,
	}
}

// Start starts the Handler with the given context.
func (h *Handler) Start(ctx context.Context) error {
	healthChecker := health.NewChecker()
	return h.Receiver.StartListen(ctx, healthChecker.Check(h), h.Logger)
}

// ServeHTTP serves an HTTP request and returns an HTTP response.
// It ensures that the incoming request is a valid Cloud Event, then dispatches it
// to the EMS gateway and writes back the HTTP response.
func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// validate request method
	if request.Method != http.MethodPost && request.Method != http.MethodGet {
		h.namedLogger().Warnf("Unexpected request method: %s", request.Method)
		h.writeResponse(writer, http.StatusMethodNotAllowed, nil)
		return
	}
	// Limit server from reading a huge payload
	request.Body = http.MaxBytesReader(writer, request.Body, h.Options.MaxRequestSize)
	uri := request.RequestURI

	// Process /publish endpoint
	// Gets a CE and sends it to BEB
	if handler.IsARequestWithCE(uri) {
		h.publishCloudEvents(writer, request)
		return
	}

	// Process /:application/v1/events
	// Publishes a legacy event as CE v1.0 to BEB
	if handler.IsARequestWithLegacyEvent(uri) {
		h.publishLegacyEventsAsCE(writer, request)
		return
	}

	// Process /:application/v1/events/subscribed
	// Fetches the list of subscriptions available for the given application
	if handler.IsARequestForSubscriptions(uri) {
		h.SubscribedProcessor.ExtractEventsFromSubscriptions(writer, request)
		return
	}

	h.writeResponse(writer, http.StatusNotFound, nil)
}

func (h *Handler) publishLegacyEventsAsCE(writer http.ResponseWriter, request *http.Request) {
	event, eventTypeOriginal := h.LegacyTransformer.TransformLegacyRequestsToCE(writer, request)
	if event == nil {
		h.namedLogger().Debug("Failed to transform legacy event to CloudEvent, event is nil")
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), h.RequestTimeout)
	defer cancel()
	h.receive(ctx, event)
	statusCode, dispatchTime, respBody := h.send(ctx, event)

	// Change response as per old error codes
	h.LegacyTransformer.TransformsCEResponseToLegacyResponse(writer, statusCode, event, string(respBody))

	h.namedLogger().With(
		"id", event.ID(),
		"source", event.Source(),
		"before", eventTypeOriginal,
		"after", event.Type(),
		"statusCode", statusCode,
		"duration", dispatchTime,
		"responseBody", string(respBody),
	).Info("Event dispatched")
}

func (h *Handler) publishCloudEvents(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), h.RequestTimeout)
	defer cancel()

	message := cev2http.NewMessageFromHttpRequest(request)
	defer func() { _ = message.Finish(nil) }()

	event, err := binding.ToEvent(ctx, message)
	if err != nil {
		h.namedLogger().Errorw("Failed to extract event from request", "error", err)
		h.writeResponse(writer, http.StatusBadRequest, []byte(err.Error()))
		return
	}

	eventTypeOriginal := event.Type()
	eventTypeClean, err := h.eventTypeCleaner.Clean(eventTypeOriginal)
	if err != nil {
		h.writeResponse(writer, http.StatusBadRequest, []byte(err.Error()))
		return
	}
	event.SetType(eventTypeClean)

	if err := event.Validate(); err != nil {
		h.namedLogger().Errorw("Request doesn't correspond to CloudEvent spec", "error", err)
		h.writeResponse(writer, http.StatusBadRequest, []byte(err.Error()))
		return
	}

	if request.Header.Get(cev2http.ContentType) == cev2event.ApplicationCloudEventsJSON {
		ctx = binding.WithForceStructured(ctx)
	} else {
		ctx = binding.WithForceBinary(ctx)
	}

	h.receive(ctx, event)
	statusCode, dispatchTime, respBody := h.send(ctx, event)
	h.writeResponse(writer, statusCode, respBody)

	h.namedLogger().With(
		"id", event.ID(),
		"source", event.Source(),
		"before", eventTypeOriginal,
		"after", eventTypeClean,
		"statusCode", statusCode,
		"duration", dispatchTime,
		"responseBody", string(respBody),
	).Info("Event dispatched")
}

// writeResponse writes the HTTP response given the status code and response body.
func (h *Handler) writeResponse(writer http.ResponseWriter, statusCode int, respBody []byte) {
	writer.WriteHeader(statusCode)

	if respBody == nil {
		return
	}
	if _, err := writer.Write(respBody); err != nil {
		h.namedLogger().Errorw("Failed to write response body", "error", err)
	}
}

// receive applies the default values (if any) to the given Cloud Event.
func (h *Handler) receive(ctx context.Context, event *cev2event.Event) {
	if h.Defaulter != nil {
		newEvent := h.Defaulter(ctx, *event)
		event = &newEvent
	}

	h.namedLogger().Infof("CloudEvent received id:[%s]", event.ID())
}

// send dispatches the given Cloud Event to the EMS gateway and returns the response details and dispatch time.
func (h *Handler) send(ctx context.Context, event *cev2event.Event) (int, time.Duration, []byte) {
	request, err := h.Sender.NewRequestWithTarget(ctx, h.Sender.Target)
	if err != nil {
		h.namedLogger().Errorw("Failed to prepare a CloudEvent request", "error", err)
		return http.StatusInternalServerError, noDuration, []byte{}
	}

	message := binding.ToMessage(event)
	defer func() { _ = message.Finish(nil) }()

	err = cloudevents.WriteRequestWithHeaders(ctx, message, request, additionalHeaders)
	if err != nil {
		h.namedLogger().Errorw("Failed to add additional headers to the request", "error", err)
		return http.StatusInternalServerError, noDuration, []byte{}
	}

	resp, dispatchTime, err := h.sendAndRecordDispatchTime(request)
	if err != nil {
		h.namedLogger().Errorw("Failed to send CloudEvent and record dispatch time", "error", err)
		h.collector.RecordError()
		return http.StatusInternalServerError, dispatchTime, []byte{}
	}
	defer func() { _ = resp.Body.Close() }()
	h.collector.RecordLatency(dispatchTime, resp.StatusCode, destinationServiceBEB)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.namedLogger().Errorw("Failed to read response body", "error", err)
		return http.StatusInternalServerError, dispatchTime, []byte{}
	}

	return resp.StatusCode, dispatchTime, body
}

// sendAndRecordDispatchTime sends a CloudEvent and records the time taken while sending.
func (h *Handler) sendAndRecordDispatchTime(request *http.Request) (*http.Response, time.Duration, error) {
	start := time.Now()
	resp, err := h.Sender.Send(request)
	dispatchTime := time.Since(start)
	return resp, dispatchTime, err
}

func (h *Handler) namedLogger() *zap.SugaredLogger {
	return h.Logger.WithContext().Named(bebCommanderName)
}
