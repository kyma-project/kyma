package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/health"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy-events"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/subscribed"

	"github.com/sirupsen/logrus"

	"github.com/cloudevents/sdk-go/v2/binding"
	cev2client "github.com/cloudevents/sdk-go/v2/client"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
)


// Handler is responsible for receiving HTTP requests and dispatching them to the EMS gateway.
// It also assures that the messages received are compliant with the Cloud Events spec.
type NatsHandler struct {
	// Receiver receives incoming HTTP requests
	Receiver *receiver.HttpMessageReceiver
	// Sender sends requests to the broker
	Sender *sender.NatsMessageSender
	// Defaulter sets default values to incoming events
	Defaulter cev2client.EventDefaulter
	// LegacyTransformer handles transformations needed to handle legacy events
	LegacyTransformer *legacy.Transformer
	// RequestTimeout timeout for outgoing requests
	RequestTimeout time.Duration
	//SubscribedProcessor processes requests for /:app/v1/events/subscribed endpoint
	SubscribedProcessor *subscribed.Processor
	// Logger default logger
	Logger *logrus.Logger
	// Options configures HTTP server
	Options *options.Options
}

// NewHandler returns a new Handler instance for the Event Publisher Proxy.
func NewNatsHandler(receiver *receiver.HttpMessageReceiver, sender *sender.NatsMessageSender, requestTimeout time.Duration, legacyTransformer *legacy.Transformer, opts *options.Options, subscribedProcessor *subscribed.Processor, logger *logrus.Logger) *NatsHandler {
	return &NatsHandler{
		Receiver:            receiver,
		Sender:              sender,
		RequestTimeout:      requestTimeout,
		LegacyTransformer:   legacyTransformer,
		SubscribedProcessor: subscribedProcessor,
		Logger:              logger,
		Options:             opts,
	}
}

// Start starts the Handler with the given context.
func (h *NatsHandler) Start(ctx context.Context) error {
	return h.Receiver.StartListen(ctx, health.CheckHealth(h))
}

// ServeHTTP serves an HTTP request and returns back an HTTP response.
// It ensures that the incoming request is a valid Cloud Event, then dispatches it
// to the EMS gateway and writes back the HTTP response.
func (h *NatsHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// validate request method
	if request.Method != http.MethodPost && request.Method != http.MethodGet {
		h.Logger.Warnf("Unexpected request method: %s", request.Method)
		h.writeResponse(writer, http.StatusMethodNotAllowed, nil)
		return
	}
	// Limit server from reading a huge payload
	request.Body = http.MaxBytesReader(writer, request.Body, h.Options.MaxRequestSize)
	uri := request.RequestURI

	// Process /publish endpoint
	// Gets a CE and sends it to BEB
	if isARequestWithCE(uri) {
		h.publishCloudEvents(writer, request)
		return
	}

	// Process /:application/v1/events
	// Publishes a legacy event as CE v1.0 to BEB
	if isARequestWithLegacyEvent(uri) {
		h.publishLegacyEventsAsCE(writer, request)
		return
	}

	// Process /:application/v1/events/subscribed
	// Fetches the list of subscriptions available for the given application
	if isARequestForSubscriptions(uri) {
		h.SubscribedProcessor.ExtractEventsFromSubscriptions(writer, request)
		return
	}

	h.writeResponse(writer, http.StatusNotFound, nil)
	return
}

func (h *NatsHandler) publishLegacyEventsAsCE(writer http.ResponseWriter, request *http.Request) {
	event := h.LegacyTransformer.TransformsLegacyRequestsToCE(writer, request)
	if event == nil {
		h.Logger.Debug("failed to transform legacy event to CE, event is nil")
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), h.RequestTimeout)
	defer cancel()
	h.receive(ctx, event)
	statusCode, dispatchTime, respBody := h.send(ctx, event)
	// Change response as per old error codes
	h.LegacyTransformer.TransformsCEResponseToLegacyResponse(writer, statusCode, event, string(respBody))
	h.Logger.Infof("Event dispatched id:[%s] statusCode:[%d] duration:[%s] responseBody:[%s]", event.ID(), statusCode, dispatchTime, respBody)
}

func (h *NatsHandler) publishCloudEvents(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), h.RequestTimeout)
	defer cancel()

	message := cehttp.NewMessageFromHttpRequest(request)
	defer func() { _ = message.Finish(nil) }()

	event, err := binding.ToEvent(ctx, message)
	if err != nil {
		h.Logger.Warnf("Failed to extract event from request with error: %s", err)
		h.writeResponse(writer, http.StatusBadRequest, []byte(err.Error()))
		return
	}

	if err := event.Validate(); err != nil {
		h.Logger.Warnf("Request is invalid as per CE spec with error: %s", err)
		h.writeResponse(writer, http.StatusBadRequest, []byte(err.Error()))
		return
	}

	h.receive(ctx, event)
	statusCode, dispatchTime, respBody := h.send(ctx, event)
	h.writeResponse(writer, statusCode, respBody)

	h.Logger.Infof("Event dispatched id:[%s] statusCode:[%d] duration:[%s] responseBody:[%s]", event.ID(), statusCode, dispatchTime, respBody)
}

// writeResponse writes the HTTP response given the status code and response body.
func (h *NatsHandler) writeResponse(writer http.ResponseWriter, statusCode int, respBody []byte) {
	writer.WriteHeader(statusCode)

	if respBody == nil {
		return
	}
	if _, err := writer.Write(respBody); err != nil {
		h.Logger.Errorf("Failed to write response body with error: %s", err)
	}
}

// receive applies the default values (if any) to the given Cloud Event.
func (h *NatsHandler) receive(ctx context.Context, event *cev2event.Event) {
	if h.Defaulter != nil {
		newEvent := h.Defaulter(ctx, *event)
		event = &newEvent
	}

	h.Logger.Infof("Event received id:[%s]", event.ID())
}

// send dispatches the given Cloud Event to NATS and returns the response details and dispatch time.
func (h *NatsHandler) send(ctx context.Context, event *cev2event.Event) (int, time.Duration, []byte) {
	start := time.Now()
	resp, err := h.Sender.Send(ctx, event)
	dispatchTime := time.Since(start)
	if err != nil {
		return resp, dispatchTime, []byte(err.Error())
	}
	return resp, dispatchTime, []byte{}
}
