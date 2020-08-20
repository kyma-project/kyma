package handler

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/ems"

	"github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/events"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/health"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/receiver"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/sender"

	cev2client "github.com/cloudevents/sdk-go/v2/client"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	"go.uber.org/zap"
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

	// Metrics server
	// MetricsServer *metrics.Server
	//// Reporter reports stats of status code and dispatch time
	//Reporter StatsReporter
	Logger *zap.Logger
}

func (h *Handler) Start(ctx context.Context) error {
	return h.Receiver.StartListen(ctx, health.WithLivenessCheck(health.WithReadinessCheck(h)))
}

//// Blocking
//func (h *Handler) StartMetricsServer(ctx context.Context) error {
//	var err error
//	if h.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", recv.metricsPort)); err != nil {
//		return err
//	}
//
//	recv.server = &nethttp.Server{
//		Addr:    recv.listener.Addr().String(),
//		Handler: recv.handler,
//	}
//
//	errChan := make(chan error, 1)
//	go func() {
//		errChan <- recv.server.Serve(recv.listener)
//	}()
//
//	// wait for the server to return or ctx.Done().
//	select {
//	case <-ctx.Done():
//		ctx, cancel := context.WithTimeout(context.Background(), getShutdownTimeout(ctx))
//		defer cancel()
//		err := recv.server.Shutdown(ctx)
//		<-errChan // Wait for server goroutine to exit
//		return err
//	case err := <-errChan:
//		return err
//	}
//}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// validate request method
	if request.Method != http.MethodPost {
		h.Logger.Warn("unexpected request method", zap.String("method", request.Method))
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
		h.Logger.Warn("failed to extract event from request", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	err = event.Validate()
	if err != nil {
		h.Logger.Warn("request is valid as per CE spec", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	statusCode, dispatchTime, respBody := h.receive(ctx, request.Header, event)
	if dispatchTime > noDuration {
		//	_ = h.Reporter.ReportEventDispatchTime(reporterArgs, statusCode, dispatchTime)
	}
	//_ = h.Reporter.ReportEventCount(reporterArgs, statusCode)

	writer.WriteHeader(statusCode)

	_, err = writer.Write(respBody)
	if err != nil {
		h.Logger.Error("failed to write response", zap.Error(err))
	}
}

func (h *Handler) receive(ctx context.Context, headers http.Header, event *cev2event.Event) (int, time.Duration, []byte) {

	if h.Defaulter != nil {
		newEvent := h.Defaulter(ctx, *event)
		event = &newEvent
	}

	h.Logger.Info("Event received", zap.Any("", *event))
	return h.send(ctx, headers, event)
}

func (h *Handler) send(ctx context.Context, headers http.Header, event *cev2event.Event) (int, time.Duration, []byte) {

	request, err := h.Sender.NewCloudEventRequestWithTarget(ctx, h.Sender.Target)
	if err != nil {
		h.Logger.Error("failed in NewCloudEventRequestWithTarget", zap.Error(err))
		return http.StatusInternalServerError, noDuration, []byte{}
	}

	// Should we do this???
	message := binding.ToMessage(event)
	defer func() { _ = message.Finish(nil) }()

	err = events.WriteHttpRequestWithAdditionalHeaders(ctx, message, request, additionalHeaders)
	if err != nil {
		h.Logger.Error("failed to add additional headers to the req", zap.Error(err))
		return http.StatusInternalServerError, noDuration, []byte{}
	}
	resp, dispatchTime, err := h.sendAndRecordDispatchTime(request)
	if err != nil {
		h.Logger.Error("failed in sendAndRecordDispatchTime", zap.Error(err))
		return http.StatusInternalServerError, dispatchTime, []byte{}
	}
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.Logger.Error("failed to read response body", zap.Error(err))
		return http.StatusInternalServerError, dispatchTime, []byte{}
	}
	return resp.StatusCode, dispatchTime, respBody
}

func (h *Handler) sendAndRecordDispatchTime(request *http.Request) (*http.Response, time.Duration, error) {
	start := time.Now()
	resp, err := h.Sender.Send(request)
	dispatchTime := time.Since(start)
	return resp, dispatchTime, err
}
