package handlers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/opts"
	trc "github.com/kyma-project/kyma/components/event-bus/internal/trace"
	"github.com/nats-io/go-nats-streaming"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	// push request headers to endpoint
	headerSourceID         = "KYMA-Source-ID"
	headerEventType        = "KYMA-Event-Type"
	headerEventTypeVersion = "KYMA-Event-Type-Version"
	headerEventID          = "KYMA-Event-ID"
	headerEventTime        = "KYMA-Event-Time"
)

type MessageHandlerFactory struct {
	tr     *http.Transport
	tracer *trc.Tracer
}

func NewMessageHandlerFactory(opts *opts.Options, tracer *trc.Tracer) *MessageHandlerFactory {
	tr := initHTTPTransport(opts)
	return &MessageHandlerFactory{
		tr:     tr,
		tracer: tracer,
	}
}

func initHTTPTransport(opts *opts.Options) *http.Transport {
	return &http.Transport{
		MaxIdleConns:       opts.MaxIdleConns,
		IdleConnTimeout:    opts.IdleConnTimeout,
		DisableCompression: true,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: opts.TLSSkipVerify},
	}
}

func (mhf *MessageHandlerFactory) NewMsgHandler(sub *v1alpha1.Subscription, opts *opts.Options, requestProvider common.RequestProvider) func(msg *stan.Msg) {

	client := &http.Client{
		Transport: mhf.tr,
		Timeout:   time.Duration(sub.PushRequestTimeoutMS) * time.Millisecond,
	}

	msgHandler := func(msg *stan.Msg) {
		var payload []byte
		cloudEvent, err := convertToCloudEvent(&msg.Data)
		if err == nil { // message is cloud event compliant, send the data field as the payload if possible
			var marshallError error
			payload, marshallError = json.Marshal(cloudEvent.Data)
			if marshallError != nil { // data is not valid JSON type value, send the message as it is
				payload = msg.Data
			}
		} else { // message is not cloud event compliant, send the message as it is
			payload = msg.Data
		}

		req, err := requestProvider(http.MethodPost, sub.Endpoint, bytes.NewBuffer(payload))
		if err != nil {
			panic(fmt.Sprintf("push HTTP request creation failed: %v", err))
		}
		req.Header.Add("Content-Type", "application/json")

		subNameSupplier := func() string { return sub.Name }
		addOptionalHeader(req, sub.IncludeSubscriptionNameHeader, opts.SubscriptionNameHeader, subNameSupplier)

		var pushSpan *opentracing.Span
		if cloudEvent != nil {
			if traceContext := getTraceContext(cloudEvent); traceContext != nil && (*mhf.tracer).Started() {
				pushSpan = mhf.createPushSpan(traceContext, cloudEvent, sub)
				defer trc.FinishSpan(pushSpan)

				addTraceContext(pushSpan, req)
			}
			annotateCloudEventHeaders(req, cloudEvent)
		}

		resp, err := client.Do(req)

		if resp != nil {
			trc.TagSpanWithHttpStatusCode(pushSpan, uint16(resp.StatusCode))
		}

		if err != nil { // failed to send push request
			// TODO introduce a metric, counter of message send failure, and update it here
			log.Printf("MsgHandler :: Error send push request failed: %v\n", err)

			trc.TagSpanAsError(pushSpan, "failed to send push request", err.Error())

			// just return from this delivery attempt, message delivery will be retried by NATS Streaming
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 { // subscriber failed to receive or process the message, returned non-2xx status code
			// TODO introduce a metric, counter of delivery failure per response code, and update it here

			trc.TagSpanAsError(pushSpan, "subscriber failed to receive or process the message, returned non-2xx status code", "")

			// just return from this delivery attempt, message delivery will be retried by NATS Streaming
			return
		}

		if err := msg.Ack(); err != nil { // failed to ACK message to NATS Streaming
			// TODO introduce a metric, counter of ACK to NATS Streaming failures, and update it here
			log.Printf("MsgHandler :: Error ACK failed: %v\n", err)

			trc.TagSpanAsError(pushSpan, "failed to ACK message to NATS Streaming", err.Error())

			// just return from this delivery attempt, message delivery will be retried by NATS Streaming
			return
		}

		// TODO introduce a metric, counter of successful deliveries, and update it here
	}

	return msgHandler
}
func annotateCloudEventHeaders(req *http.Request, cloudEvent *api.CloudEvent) {
	// add cloud event properties as request headers except the data field
	req.Header.Add(headerSourceID, cloudEvent.SourceID)
	req.Header.Add(headerEventType, cloudEvent.EventType)
	req.Header.Add(headerEventTypeVersion, cloudEvent.EventTypeVersion)
	req.Header.Add(headerEventID, cloudEvent.EventID)
	req.Header.Add(headerEventTime, cloudEvent.EventTime)
}

func (mhf *MessageHandlerFactory) createPushSpan(traceContext *api.TraceContext, cloudEvent *api.CloudEvent, sub *v1alpha1.Subscription) *opentracing.Span {
	spanContext := trc.ReadTraceContext(traceContext)
	pushSpan := trc.StartSpan(spanContext, &(*mhf.tracer).Options().OperationName, ext.SpanKindConsumer)
	addSpanTagsForCloudEventAndSubscription(pushSpan, cloudEvent, sub)
	return pushSpan
}

func addSpanTagsForCloudEventAndSubscription(pushSpan *opentracing.Span, cloudEvent *api.CloudEvent, sub *v1alpha1.Subscription) {
	tags := trc.CreateTraceTagsFromCloudEvent(cloudEvent)
	tags[trc.SubscriptionName] = sub.Name
	tags[trc.SubscriptionEnvironment] = sub.Namespace
	trc.SetSpanTags(pushSpan, &tags)
}

func addTraceContext(span *opentracing.Span, req *http.Request) error {
	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	return opentracing.GlobalTracer().Inject((*span).Context(), opentracing.HTTPHeaders, carrier)
}

func addOptionalHeader(request *http.Request, includeHeader bool, key string, valueSupplier func() string) {
	if includeHeader {
		value := valueSupplier()
		request.Header.Add(key, value)
	}
}

func convertToCloudEvent(payload *[]byte) (*api.CloudEvent, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload is null")
	}

	var cloudEvent *api.CloudEvent
	err := json.Unmarshal(*payload, &cloudEvent)

	if err != nil {
		return nil, err
	}

	if err := api.ValidatePublish(&cloudEvent.PublishRequest, api.GetDefaultEventOptions()); err != nil {
		return nil, fmt.Errorf("payload is not valid: %v", string(*payload))
	}

	return cloudEvent, nil
}

func getTraceContext(cloudEvent *api.CloudEvent) *api.TraceContext {
	if cloudEvent == nil {
		return nil
	}

	if traceContextWrapper, present := cloudEvent.Extensions[api.FieldTraceContext]; present {
		if traceContextMap, ok := traceContextWrapper.(map[string]interface{}); ok {
			traceContext := api.TraceContext{}
			for key, value := range traceContextMap {
				traceContext[key] = fmt.Sprint(value)
			}
			return &traceContext
		}
	}

	return nil
}
