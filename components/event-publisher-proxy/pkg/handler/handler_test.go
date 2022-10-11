package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	eclogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype/eventtypetest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/legacytest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/metricstest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender/beb"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func Test_extractCloudEventFromRequest(t *testing.T) {
	type args struct {
		request *http.Request
	}
	type wants struct {
		event              *cev2event.Event
		errorAssertionFunc assert.ErrorAssertionFunc
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "Valid event",
			args: args{
				request: CreateValidStructuredRequest(t),
			},
			wants: wants{
				event:              CreateCloudEvent(t),
				errorAssertionFunc: assert.NoError,
			},
		},
		{
			name: "Invalid event",
			args: args{
				request: CreateInvalidStructuredRequest(t),
			},
			wants: wants{
				event:              nil,
				errorAssertionFunc: assert.Error,
			},
		},
		{
			name: "Entirely broken Request",
			args: args{
				request: CreateBrokenRequest(t),
			},
			wants: wants{
				event:              nil,
				errorAssertionFunc: assert.Error,
			},
		},
		{
			name: "Valid event",
			args: args{
				request: CreateValidBinaryRequest(t),
			},
			wants: wants{
				event:              CreateCloudEvent(t),
				errorAssertionFunc: assert.NoError,
			},
		},
		{
			name: "Invalid event",
			args: args{
				request: CreateInvalidBinaryRequest(t),
			},
			wants: wants{
				event:              nil,
				errorAssertionFunc: assert.Error,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEvent, err := extractCloudEventFromRequest(tt.args.request)
			if !tt.wants.errorAssertionFunc(t, err, fmt.Sprintf("extractCloudEventFromRequest(%v)", tt.args.request)) {
				return
			}
			assert.Equalf(t, tt.wants.event, gotEvent, "extractCloudEventFromRequest(%v)", tt.args.request)
		})
	}
}

func Test_writeResponse(t *testing.T) {
	type args struct {
		statusCode int
		respBody   []byte
	}
	tests := []struct {
		name          string
		args          args
		assertionFunc assert.ErrorAssertionFunc
	}{
		{
			name: "Response and body",
			args: args{
				statusCode: 200,
				respBody:   []byte("foo"),
			},
			assertionFunc: assert.NoError,
		},
		{
			name: "Response and no body",
			args: args{
				statusCode: 200,
				respBody:   nil,
			},
			assertionFunc: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			writer := httptest.NewRecorder()

			// when
			err := writeResponse(writer, tt.args.statusCode, tt.args.respBody)

			// then
			tt.assertionFunc(t, err, fmt.Sprintf("writeResponse(%v, %v)", tt.args.statusCode, tt.args.respBody))
			assert.Equal(t, tt.args.statusCode, writer.Result().StatusCode)
			body, err := io.ReadAll(writer.Result().Body)
			assert.NoError(t, err)
			if tt.args.respBody != nil {
				assert.Equal(t, tt.args.respBody, body)
			} else {
				assert.Equal(t, []byte(""), body)
			}
		})
	}
}

func TestHandler_publishCloudEvents(t *testing.T) {
	type fields struct {
		Sender           sender.GenericSender
		collector        metrics.PublishingMetricsCollector
		eventTypeCleaner eventtype.Cleaner
	}
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "Publish structured Cloudevent",
			fields: fields{
				Sender: &GenericSenderStub{
					Err: nil,
					Result: beb.HTTPPublishResult{
						Status: 204,
						Body:   []byte(""),
					},
				},
				collector:        metricstest.PublishingMetricsCollectorStub{},
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateValidStructuredRequest(t),
			},
			wantStatus: 204,
		},
		{
			name: "Publish binary Cloudevent",
			fields: fields{
				Sender: &GenericSenderStub{
					Err: nil,
					Result: beb.HTTPPublishResult{
						Status: 204,
						Body:   []byte(""),
					},
				},
				collector:        metricstest.PublishingMetricsCollectorStub{},
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateValidBinaryRequest(t),
			},
			wantStatus: 204,
		},
		{
			name: "Publish invalid structured CloudEvent",
			fields: fields{
				Sender:           &GenericSenderStub{},
				collector:        metricstest.PublishingMetricsCollectorStub{},
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateInvalidStructuredRequest(t),
			},
			wantStatus: 400,
		},
		{
			name: "Publish invalid binary CloudEvent",
			fields: fields{
				Sender:           &GenericSenderStub{},
				collector:        metricstest.PublishingMetricsCollectorStub{},
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateInvalidBinaryRequest(t),
			},
			wantStatus: 400,
		},
		{
			name: "Publish binary CloudEvent but cannot clean",
			fields: fields{
				Sender:    &GenericSenderStub{},
				collector: metricstest.PublishingMetricsCollectorStub{},
				eventTypeCleaner: &eventtypetest.CleanerStub{
					CleanType: "",
					Error:     fmt.Errorf("I cannot clean"),
				},
			},
			args: args{
				request: CreateValidBinaryRequest(t),
			},
			wantStatus: 400,
		},
		{
			name: "Publish binary CloudEvent but cannot send",
			fields: fields{
				Sender: &GenericSenderStub{
					Err: fmt.Errorf("I cannot send"),
				},
				collector:        metricstest.PublishingMetricsCollectorStub{},
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateValidBinaryRequest(t),
			},
			wantStatus: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			logger, err := eclogger.New("text", "debug")
			assert.NoError(t, err)

			h := &Handler{
				Sender:           tt.fields.Sender,
				Logger:           logger,
				collector:        tt.fields.collector,
				eventTypeCleaner: tt.fields.eventTypeCleaner,
			}
			writer := httptest.NewRecorder()

			// when
			h.publishCloudEvents(writer, tt.args.request)

			// then
			assert.Equal(t, tt.wantStatus, writer.Result().StatusCode)
			body, err := io.ReadAll(writer.Result().Body)
			assert.NoError(t, err)
			if tt.wantBody != nil {
				assert.Equal(t, tt.wantBody, body)
			}
		})
	}
}

func TestHandler_publishLegacyEventsAsCE(t *testing.T) {
	type fields struct {
		Sender            sender.GenericSender
		LegacyTransformer legacy.RequestToCETransformer
		collector         metrics.PublishingMetricsCollector
		eventTypeCleaner  eventtype.Cleaner
	}
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
		wantBody   []byte
	}{
		{
			name: "Send valid legacy event",
			fields: fields{
				Sender: &GenericSenderStub{
					Result: beb.HTTPPublishResult{
						Status: 204,
					},
				},
				LegacyTransformer: legacy.NewTransformer("namespace", "im.a.prefix", NewApplicationListerOrDie(context.Background(), "testapp")),
				collector:         metricstest.PublishingMetricsCollectorStub{},
				eventTypeCleaner:  eventtypetest.CleanerStub{},
			},
			args: args{
				request: legacytest.ValidLegacyRequestOrDie(t, "v1", "testapp", "object.created"),
			},
			wantStatus: 200,
			wantBody:   nil,
		},
		{
			name: "Send valid legacy event but cannot send to backend",
			fields: fields{
				Sender: &GenericSenderStub{
					Err: fmt.Errorf("i cannot send"),
				},
				LegacyTransformer: legacy.NewTransformer("namespace", "im.a.prefix", NewApplicationListerOrDie(context.Background(), "testapp")),
				collector:         metricstest.PublishingMetricsCollectorStub{},
				eventTypeCleaner:  eventtypetest.CleanerStub{},
			},
			args: args{
				request: legacytest.ValidLegacyRequestOrDie(t, "v1", "testapp", "object.created"),
			},
			wantStatus: 500,
			wantBody:   nil,
		},
		{
			name: "Send invalid legacy event",
			fields: fields{
				Sender: &GenericSenderStub{
					Result: beb.HTTPPublishResult{
						Status: 204,
					},
				},
				LegacyTransformer: legacy.NewTransformer("namespace", "im.a.prefix", NewApplicationListerOrDie(context.Background(), "testapp")),
				collector:         metricstest.PublishingMetricsCollectorStub{},
				eventTypeCleaner:  eventtypetest.CleanerStub{},
			},
			args: args{
				request: legacytest.InvalidLegacyRequestOrDie(t, "v1", "testapp", "object.created"),
			},
			wantStatus: 400,
			wantBody:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			logger, err := eclogger.New("text", "debug")
			assert.NoError(t, err)

			h := &Handler{
				Sender:            tt.fields.Sender,
				Logger:            logger,
				LegacyTransformer: tt.fields.LegacyTransformer,
				collector:         tt.fields.collector,
				eventTypeCleaner:  tt.fields.eventTypeCleaner,
			}
			writer := httptest.NewRecorder()

			// when
			h.publishLegacyEventsAsCE(writer, tt.args.request)

			// then
			assert.Equal(t, tt.wantStatus, writer.Result().StatusCode)
			body, err := io.ReadAll(writer.Result().Body)
			assert.NoError(t, err)
			if tt.wantBody != nil {
				assert.Equal(t, tt.wantBody, body)
			}
		})
	}
}

func TestHandler_maxBytes(t *testing.T) {
	type fields struct {
		maxBytes int
	}
	tests := []struct {
		name       string
		fields     fields
		wantStatus int
	}{
		{
			name: "request small enough",
			fields: fields{
				maxBytes: 10000,
			},
			wantStatus: 200,
		},
		{
			name: "request too large",
			fields: fields{
				maxBytes: 1,
			},
			wantStatus: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			h := &Handler{
				Options: &options.Options{
					MaxRequestSize: int64(tt.fields.maxBytes),
				},
			}
			writer := httptest.NewRecorder()
			var mberr *http.MaxBytesError
			f := func(writer http.ResponseWriter, r *http.Request) {
				_, err := io.ReadAll(r.Body)
				if errors.As(err, &mberr) {
					writer.WriteHeader(http.StatusBadRequest)
				}
				writer.WriteHeader(http.StatusOK)
			}

			// when
			h.maxBytes(f)(writer, &http.Request{
				Method: http.MethodPost,
				Body:   io.NopCloser(strings.NewReader(strings.Repeat("#", 5))),
			})

			// then
			assert.Equal(t, tt.wantStatus, writer.Result().StatusCode)
		})
	}
}

func TestHandler_sendEventAndRecordMetrics(t *testing.T) {
	type fields struct {
		Sender    sender.GenericSender
		Defaulter client.EventDefaulter
		collector metrics.PublishingMetricsCollector
	}
	type args struct {
		ctx    context.Context
		host   string
		event  *cev2event.Event
		header http.Header
	}
	type wants struct {
		result          sender.PublishResult
		assertionFunc   assert.ErrorAssertionFunc
		metricErrors    int
		metricTotal     int
		metricLatency   int
		metricPublished int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{
			name: "No Error",
			fields: fields{
				Sender: &GenericSenderStub{
					Err:           nil,
					SleepDuration: 0,
					Result: beb.HTTPPublishResult{
						Status: 204,
						Body:   nil,
					},
				},
				Defaulter: nil,
				collector: metrics.NewCollector(),
			},
			args: args{
				ctx:   context.Background(),
				host:  "foo",
				event: &cev2event.Event{},
			},
			wants: wants{
				result: beb.HTTPPublishResult{
					Status: 204,
					Body:   nil,
				},
				assertionFunc:   assert.NoError,
				metricErrors:    0,
				metricTotal:     1,
				metricLatency:   1,
				metricPublished: 1,
			},
		},
		{
			name: "Sending not successful, error returned",
			fields: fields{
				Sender: &GenericSenderStub{
					Err:           errors.New("i failed"),
					SleepDuration: 5,
				},
				Defaulter: nil,
				collector: metrics.NewCollector(),
			},
			args: args{
				ctx:   context.Background(),
				host:  "foo",
				event: &cev2event.Event{},
			},
			wants: wants{
				result:          nil,
				assertionFunc:   assert.Error,
				metricErrors:    1,
				metricTotal:     0,
				metricLatency:   0,
				metricPublished: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			h := &Handler{
				Sender:    tt.fields.Sender,
				Defaulter: tt.fields.Defaulter,
				collector: tt.fields.collector,
			}

			// when
			got, err := h.sendEventAndRecordMetrics(tt.args.ctx, tt.args.event, tt.args.host, tt.args.header)

			// then
			if !tt.wants.assertionFunc(t, err, fmt.Sprintf("sendEventAndRecordMetrics(%v, %v, %v)", tt.args.ctx, tt.args.host, tt.args.event)) {
				return
			}
			assert.Equalf(t, tt.wants.result, got, "sendEventAndRecordMetrics(%v, %v, %v)", tt.args.ctx, tt.args.host, tt.args.event)
			metricstest.EnsureMetricErrors(t, h.collector, tt.wants.metricErrors)
			metricstest.EnsureMetricTotalRequests(t, h.collector, tt.wants.metricTotal)
			metricstest.EnsureMetricLatency(t, h.collector, tt.wants.metricLatency)
			metricstest.EnsureMetricEventTypePublished(t, h.collector, tt.wants.metricPublished)
		})
	}
}

func TestHandler_sendEventAndRecordMetrics_TracingAndDefaults(t *testing.T) {
	// given
	stub := &GenericSenderStub{
		Err:           nil,
		SleepDuration: 0,
		Result:        beb.HTTPPublishResult{Status: http.StatusInternalServerError},
	}

	h := &Handler{
		Sender:    stub,
		Defaulter: nil,
		collector: metrics.NewCollector(),
	}
	header := http.Header{}
	headers := []string{"traceparent", "X-B3-TraceId", "X-B3-ParentSpanId", "X-B3-SpanId", "X-B3-Sampled", "X-B3-Flags"}

	for _, v := range headers {
		header.Add(v, v)
	}
	expectedExtensions := map[string]interface{}{
		"traceparent":    "traceparent",
		"b3traceid":      "X-B3-TraceId",
		"b3parentspanid": "X-B3-ParentSpanId",
		"b3spanid":       "X-B3-SpanId",
		"b3sampled":      "X-B3-Sampled",
		"b3flags":        "X-B3-Flags",
	}
	// when
	_, err := h.sendEventAndRecordMetrics(context.Background(), CreateCloudEvent(t), "", header)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expectedExtensions, stub.ReceivedEvent.Context.GetExtensions())
}

func CreateCloudEvent(t *testing.T) *cev2event.Event {
	builder := testingutils.NewCloudEventBuilder(
		testingutils.WithCloudEventType(testingutils.CloudEventType),
	)
	payload, _ := builder.BuildStructured()
	newEvent := cloudevents.NewEvent()
	err := json.Unmarshal([]byte(payload), &newEvent)
	assert.NoError(t, err)
	newEvent.SetType(testingutils.CloudEventType)
	err = newEvent.SetData("", map[string]interface{}{"foo": "bar"})
	assert.NoError(t, err)

	return &newEvent
}

// CreateValidStructuredRequest creates a structured cloudevent as http request.
func CreateValidStructuredRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", strings.NewReader("{\"specversion\":\"1.0\",\"type\":\"prefix.testapp1023.order.created.v1\",\"source\":\"/default/sap.kyma/id\",\"id\":\"8945ec08-256b-11eb-9928-acde48001122\",\"data\":{\"foo\":\"bar\"}}"))
	req.Header.Add("Content-Type", "application/cloudevents+json")
	return req
}

// CreateBrokenRequest creates a structured cloudevent request that cannot be parsed.
func CreateBrokenRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", strings.NewReader("I AM JUST A BROKEN REQUEST"))
	req.Header.Add("Content-Type", "application/cloudevents+json")
	return req
}

// CreateInvalidStructuredRequest creates an invalid structured cloudevent as http request. The `type` is missing.
func CreateInvalidStructuredRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", strings.NewReader("{\"specversion\":\"1.0\",\"source\":\"/default/sap.kyma/id\",\"id\":\"8945ec08-256b-11eb-9928-acde48001122\",\"data\":{\"foo\":\"bar\"}}"))
	req.Header.Add("Content-Type", "application/cloudevents+json")
	return req
}

// CreateValidBinaryRequest creates a valid binary cloudevent as http request.
func CreateValidBinaryRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", strings.NewReader("{\"foo\":\"bar\"}"))
	req.Header.Add("Ce-Specversion", "1.0")
	req.Header.Add("Ce-Type", "prefix.testapp1023.order.created.v1")
	req.Header.Add("Ce-Source", "/default/sap.kyma/id")
	req.Header.Add("Ce-ID", "8945ec08-256b-11eb-9928-acde48001122")
	return req
}

// CreateInvalidBinaryRequest creates an invalid binary cloudevent as http request. The `type` is missing.
func CreateInvalidBinaryRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", strings.NewReader("{\"foo\":\"bar\"}"))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Ce-Specversion", "1.0")
	req.Header.Add("Ce-Source", "/default/sap.kyma/id")
	req.Header.Add("Ce-ID", "8945ec08-256b-11eb-9928-acde48001122")
	return req
}

type GenericSenderStub struct {
	Err           error
	SleepDuration time.Duration
	Result        sender.PublishResult
	ReceivedEvent *cev2event.Event
}

func (g *GenericSenderStub) Send(_ context.Context, event *cev2event.Event) (sender.PublishResult, error) {
	g.ReceivedEvent = event
	time.Sleep(g.SleepDuration)
	return g.Result, g.Err
}

func (g *GenericSenderStub) URL() string {
	return ""
}
