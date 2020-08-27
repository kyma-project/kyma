package handler

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/cloudevents/sdk-go/v2/binding"
	cev2client "github.com/cloudevents/sdk-go/v2/client"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"

	cloudevents "github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/cloudevents"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/ems"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/health"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/sender"
)

const (
	// noDuration signals that the dispatch step hasn't started
	noDuration = -1
)

var (
	additionalHeaders = http.Header{
		"qos":    []string{string(ems.QosAtLeastOnce)},
		"Accept": []string{"application/json"},
	}
)

type Handler struct {
	// Receiver receives incoming HTTP requests
	Receiver *receiver.HttpMessageReceiver
	// Sender sends requests to the broker
	Sender *sender.HttpMessageSender
	// Defaults sets default values to incoming events
	Defaulter cev2client.EventDefaulter

	Logger *zap.Logger
}

func NewHandler(receiver *receiver.HttpMessageReceiver, sender *sender.HttpMessageSender, logger *zap.Logger) *Handler {
	return &Handler{Receiver: receiver, Sender: sender, Logger: logger}
}

func (h *Handler) Start(ctx context.Context) error {
	return h.Receiver.StartListen(ctx, health.WithLivenessCheck(health.WithReadinessCheck(h)))
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// validate request method
	if request.Method != http.MethodPost {
		h.Logger.Warn("Unexpected request method", zap.String("method", request.Method))
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// validate request URI
	if request.RequestURI != "/publish" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	ctx := request.Context()

	message := cehttp.NewMessageFromHttpRequest(request)
	defer func() { _ = message.Finish(nil) }()

	event, err := binding.ToEvent(ctx, message)
	if err != nil {
		h.Logger.Warn("Failed to extract event from request", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := event.Validate(); err != nil {
		h.Logger.Warn("Request is invalid as per CE spec", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	statusCode, dispatchTime, respBody := h.receive(ctx, event)
	h.Logger.Info("Event dispatched", zap.Duration("duration", dispatchTime))

	writer.WriteHeader(statusCode)

	if _, err := writer.Write(respBody); err != nil {
		h.Logger.Error("Failed to write response", zap.Error(err))
	}
}

func (h *Handler) receive(ctx context.Context, event *cev2event.Event) (int, time.Duration, []byte) {
	if h.Defaulter != nil {
		newEvent := h.Defaulter(ctx, *event)
		event = &newEvent
	}

	h.Logger.Info("Event received", zap.Any("id", event.ID()))

	return h.send(ctx, event)
}

func (h *Handler) send(ctx context.Context, event *cev2event.Event) (int, time.Duration, []byte) {
	request, err := h.Sender.NewCloudEventRequestWithTarget(ctx, h.Sender.Target)
	if err != nil {
		h.Logger.Error("Failed to prepare a cloudevent request with target", zap.Error(err))
		return http.StatusInternalServerError, noDuration, []byte{}
	}

	message := binding.ToMessage(event)
	defer func() { _ = message.Finish(nil) }()

	err = cloudevents.WriteRequestWithHeaders(ctx, message, request, additionalHeaders)
	if err != nil {
		h.Logger.Error("Failed to add additional headers to the request", zap.Error(err))
		return http.StatusInternalServerError, noDuration, []byte{}
	}

	resp, dispatchTime, err := h.sendAndRecordDispatchTime(request)
	if err != nil {
		h.Logger.Error("Failed to send event and record dispatch time", zap.Error(err))
		return http.StatusInternalServerError, dispatchTime, []byte{}
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.Logger.Error("Failed to read response body", zap.Error(err))
		return http.StatusInternalServerError, dispatchTime, []byte{}
	}

	return resp.StatusCode, dispatchTime, body
}

func (h *Handler) sendAndRecordDispatchTime(request *http.Request) (*http.Response, time.Duration, error) {
	start := time.Now()
	resp, err := h.Sender.Send(request)
	dispatchTime := time.Since(start)
	return resp, dispatchTime, err
}
