//nolint:lll // this test uses many long lines directly from prometheus output
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

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/builder"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender/common"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	cev2event "github.com/cloudevents/sdk-go/v2/event"
	eclogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype/eventtypetest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/histogram/mocks"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/metricstest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
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
		name     string
		args     args
		wantType string
		wants    wants
	}{
		{
			name: "Valid event",
			args: args{
				request: CreateValidStructuredRequestV1Alpha1(t),
			},
			wantType: fmt.Sprintf("sap.kyma.custom.%s", testingutils.CloudEventType),
			wants: wants{
				event:              CreateCloudEvent(t),
				errorAssertionFunc: assert.NoError,
			},
		},
		{
			name: "Invalid event",
			args: args{
				request: CreateInvalidStructuredRequestV1Alpha1(t),
			},
			wants: wants{
				event:              nil,
				errorAssertionFunc: assert.Error,
			},
		},
		{
			name: "Entirely broken Request",
			args: args{
				request: CreateBrokenRequestV1Alpha1(t),
			},
			wants: wants{
				event:              nil,
				errorAssertionFunc: assert.Error,
			},
		},
		{
			name: "Valid event",
			args: args{
				request: CreateValidBinaryRequestV1Alpha1(t),
			},
			wantType: fmt.Sprintf("sap.kyma.custom.%s", testingutils.CloudEventType),
			wants: wants{
				event:              CreateCloudEvent(t),
				errorAssertionFunc: assert.NoError,
			},
		},
		{
			name: "Invalid event",
			args: args{
				request: CreateInvalidBinaryRequestV1Alpha1(t),
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
			if tt.wantType != "" {
				tt.wants.event.SetType(tt.wantType)
			}
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

func TestHandler_publishCloudEvents_v1alpha1(t *testing.T) {
	type fields struct {
		Sender           sender.GenericSender
		collector        metrics.PublishingMetricsCollector
		eventTypeCleaner eventtype.Cleaner
	}
	type args struct {
		request *http.Request
	}

	const bucketsFunc = "Buckets"
	latency := new(mocks.BucketsProvider)
	latency.On(bucketsFunc).Return(nil)
	latency.Test(t)

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantStatus int
		wantBody   []byte
		wantTEF    string
	}{
		{
			name: "Publish structured Cloudevent",
			fields: fields{
				Sender: &GenericSenderStub{
					Err:        nil,
					BackendURL: "FOO",
				},
				collector:        metrics.NewCollector(latency),
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateValidStructuredRequestV1Alpha1(t),
			},
			wantStatus: 204,
			wantTEF: metricstest.MakeTEFBackendDuration(204, "FOO") +
				metricstest.MakeTEFBackendRequests(204, "FOO") +
				metricstest.MakeTEFEventTypePublished(204, "/default/sap.kyma/id", ""),
		},
		{
			name: "Publish binary Cloudevent",
			fields: fields{
				Sender: &GenericSenderStub{
					Err:        nil,
					BackendURL: "FOO",
				},
				collector:        metrics.NewCollector(latency),
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateValidBinaryRequestV1Alpha1(t),
			},
			wantStatus: 204,
			wantTEF: metricstest.MakeTEFBackendDuration(204, "FOO") +
				metricstest.MakeTEFBackendRequests(204, "FOO") +
				metricstest.MakeTEFEventTypePublished(204, "/default/sap.kyma/id", ""),
		},
		{
			name: "Publish invalid structured CloudEvent",
			fields: fields{
				Sender:           &GenericSenderStub{},
				collector:        metrics.NewCollector(latency),
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateInvalidStructuredRequestV1Alpha1(t),
			},
			wantStatus: 400,
			wantBody:   []byte("type: MUST be a non-empty string\n"),
		},
		{
			name: "Publish invalid binary CloudEvent",
			fields: fields{
				Sender:           &GenericSenderStub{},
				collector:        metrics.NewCollector(latency),
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateInvalidBinaryRequestV1Alpha1(t),
			},
			wantStatus: 400,
		},
		{
			name: "Publish binary CloudEvent but cannot clean",
			fields: fields{
				Sender:    &GenericSenderStub{},
				collector: metrics.NewCollector(latency),
				eventTypeCleaner: &eventtypetest.CleanerStub{
					CleanType: "",
					Error:     fmt.Errorf("I cannot clean"),
				},
			},
			args: args{
				request: CreateValidBinaryRequestV1Alpha1(t),
			},
			wantStatus: 400,
			wantBody:   []byte("I cannot clean"),
			wantTEF:    "", // client error will not be recorded as EPP internal error. So no metric will be updated.
		},
		{
			name: "Publish binary CloudEvent but cannot send",
			fields: fields{
				Sender: &GenericSenderStub{
					Err: common.BackendPublishError{},
				},
				collector:        metrics.NewCollector(latency),
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateValidBinaryRequestV1Alpha1(t),
			},
			wantStatus: 500,
			wantTEF: metricstest.MakeTEFBackendDuration(500, "") +
				metricstest.MakeTEFBackendRequests(500, "") +
				metricstest.MakeTEFBackendErrors(),
		},
		{
			name: "Publish binary CloudEvent but backend is full",
			fields: fields{
				Sender: &GenericSenderStub{
					Err: common.ErrInsufficientStorage,
				},
				collector:        metrics.NewCollector(latency),
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateValidBinaryRequestV1Alpha1(t),
			},
			wantStatus: http.StatusInsufficientStorage,
			wantTEF: metricstest.MakeTEFBackendDuration(507, "") +
				metricstest.MakeTEFBackendRequests(507, "") +
				metricstest.MakeTEFBackendErrors(),
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
				Options:          &options.Options{},
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

			metricstest.EnsureMetricMatchesTextExpositionFormat(t, h.collector, tt.wantTEF)
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
		result                  sender.PublishError
		assertionFunc           assert.ErrorAssertionFunc
		metricErrors            int
		metricTotal             int
		metricLatency           int
		metricPublished         int
		metricLatencyTEF        string
		metricPublishedTotalTEF string
	}

	const bucketsFunc = "Buckets"
	latency := new(mocks.BucketsProvider)
	latency.On(bucketsFunc).Return(nil)
	latency.Test(t)
	latencyMetricTEF := `
					# HELP eventing_epp_backend_duration_milliseconds The duration of sending events to the messaging server in milliseconds
					# TYPE eventing_epp_backend_duration_milliseconds histogram
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="0.005"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="0.01"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="0.025"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="0.05"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="0.1"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="0.25"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="0.5"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="1"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="2.5"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="5"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="10"} 1
					eventing_epp_backend_duration_milliseconds_bucket{code="204",destination_service="foo",le="+Inf"} 1
					eventing_epp_backend_duration_milliseconds_sum{code="204",destination_service="foo"} 0
					eventing_epp_backend_duration_milliseconds_count{code="204",destination_service="foo"} 1
					`

	ceEvent := CreateCloudEvent(t)
	ceEventWithOriginalEventType := ceEvent.Clone()
	ceEventWithOriginalEventType.SetExtension(builder.OriginalTypeHeaderName, testingutils.CloudEventNameAndVersion)

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
				},
				Defaulter: nil,
				collector: metrics.NewCollector(latency),
			},
			args: args{
				ctx:   context.Background(),
				host:  "foo",
				event: ceEvent,
			},
			wants: wants{
				assertionFunc:    assert.NoError,
				metricErrors:     0,
				metricTotal:      1,
				metricLatency:    1,
				metricPublished:  1,
				metricLatencyTEF: latencyMetricTEF,
				metricPublishedTotalTEF: `
					# HELP eventing_epp_event_type_published_total The total number of events published for a given eventTypeLabel
					# TYPE eventing_epp_event_type_published_total counter
					eventing_epp_event_type_published_total{code="204",event_source="/default/sap.kyma/id",event_type="prefix.testapp1023.order.created.v1"} 1
					`,
			},
		},
		{
			name: "No Error - set original event type top published metric",
			fields: fields{
				Sender: &GenericSenderStub{
					Err:           nil,
					SleepDuration: 0,
				},
				Defaulter: nil,
				collector: metrics.NewCollector(latency),
			},
			args: args{
				ctx:   context.Background(),
				host:  "foo",
				event: &ceEventWithOriginalEventType,
			},
			wants: wants{
				assertionFunc:    assert.NoError,
				metricErrors:     0,
				metricTotal:      1,
				metricLatency:    1,
				metricPublished:  1,
				metricLatencyTEF: latencyMetricTEF,
				metricPublishedTotalTEF: `
					# HELP eventing_epp_event_type_published_total The total number of events published for a given eventTypeLabel
					# TYPE eventing_epp_event_type_published_total counter
					eventing_epp_event_type_published_total{code="204",event_source="/default/sap.kyma/id",event_type="order.created.v1"} 1
					`,
			},
		},
		{
			name: "Sending not successful, error returned",
			fields: fields{
				Sender: &GenericSenderStub{
					Err:           common.BackendPublishError{},
					SleepDuration: 5,
				},
				Defaulter: nil,
				collector: metrics.NewCollector(latency),
			},
			args: args{
				ctx:   context.Background(),
				host:  "foo",
				event: &cev2event.Event{},
			},
			wants: wants{
				result:           nil,
				assertionFunc:    assert.Error,
				metricErrors:     1,
				metricTotal:      1,
				metricLatency:    1,
				metricPublished:  0,
				metricLatencyTEF: metricstest.MakeTEFBackendDuration(500, "foo"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			logger, _ := eclogger.New("text", "debug")
			h := &Handler{
				Sender:    tt.fields.Sender,
				Defaulter: tt.fields.Defaulter,
				collector: tt.fields.collector,
				Logger:    logger,
			}

			// when
			err := h.sendEventAndRecordMetrics(tt.args.ctx, tt.args.event, tt.args.host, tt.args.header)

			// then
			if !tt.wants.assertionFunc(t, err, fmt.Sprintf("sendEventAndRecordMetrics(%v, %v, %v)", tt.args.ctx, tt.args.host, tt.args.event)) {
				return
			}
			metricstest.EnsureMetricErrors(t, h.collector, tt.wants.metricErrors)
			metricstest.EnsureMetricTotalRequests(t, h.collector, tt.wants.metricTotal)
			metricstest.EnsureMetricLatency(t, h.collector, tt.wants.metricLatency)
			metricstest.EnsureMetricEventTypePublished(t, h.collector, tt.wants.metricPublished)
			metricstest.EnsureMetricMatchesTextExpositionFormat(t, h.collector, tt.wants.metricLatencyTEF, "eventing_epp_backend_duration_milliseconds")
			metricstest.EnsureMetricMatchesTextExpositionFormat(t, h.collector, tt.wants.metricPublishedTotalTEF, "eventing_epp_event_type_published_total")
		})
	}
}

func TestHandler_sendEventAndRecordMetrics_TracingAndDefaults(t *testing.T) {
	// given
	stub := &GenericSenderStub{
		SleepDuration: 0,
		Err:           common.BackendPublishError{HTTPCode: http.StatusInternalServerError},
	}

	const bucketsFunc = "Buckets"
	latency := new(mocks.BucketsProvider)
	latency.On(bucketsFunc).Return(nil)
	latency.Test(t)
	logger, _ := eclogger.New("text", "debug")
	h := &Handler{
		Sender:    stub,
		Defaulter: nil,
		collector: metrics.NewCollector(latency),
		Logger:    logger,
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
	err := h.sendEventAndRecordMetrics(context.Background(), CreateCloudEvent(t), "", header)

	// then
	assert.Error(t, err)
	assert.Equal(t, expectedExtensions, stub.ReceivedEvent.Context.GetExtensions())
}

func CreateCloudEvent(t *testing.T) *cev2event.Event {
	builder := testingutils.NewCloudEventBuilder(
		testingutils.WithCloudEventType(testingutils.CloudEventTypeWithPrefix),
	)
	payload, _ := builder.BuildStructured()
	newEvent := cloudevents.NewEvent()
	err := json.Unmarshal([]byte(payload), &newEvent)
	assert.NoError(t, err)
	newEvent.SetType(testingutils.CloudEventTypeWithPrefix)
	err = newEvent.SetData("", map[string]interface{}{"foo": "bar"})
	assert.NoError(t, err)

	return &newEvent
}

// CreateValidStructuredRequestV1Alpha1 creates a structured cloudevent as http request.
func CreateValidStructuredRequestV1Alpha1(t *testing.T) *http.Request {
	t.Helper()
	s := `{
			"specversion":"1.0",
			"type":"sap.kyma.custom.testapp1023.order.created.v1",
			"source":"/default/sap.kyma/id",
			"id":"8945ec08-256b-11eb-9928-acde48001122",
			"data":{"foo":"bar"}
			}`
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", strings.NewReader(s))
	req.Header.Add("Content-Type", "application/cloudevents+json")
	return req
}

// CreateBrokenRequestV1Alpha1 creates a structured cloudevent request that cannot be parsed.
func CreateBrokenRequestV1Alpha1(t *testing.T) *http.Request {
	t.Helper()
	reader := strings.NewReader("I AM JUST A BROKEN REQUEST")
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", reader)
	req.Header.Add("Content-Type", "application/cloudevents+json")
	return req
}

// CreateInvalidStructuredRequestV1Alpha1 creates an invalid structured cloudevent as http request. The `type` is missing.
func CreateInvalidStructuredRequestV1Alpha1(t *testing.T) *http.Request {
	t.Helper()
	s := `{
			"specversion":"1.0",
			"source":"/default/sap.kyma/id",
			"id":"8945ec08-256b-11eb-9928-acde48001122",
			"data": {
				"foo":"bar"
			}
	}`
	reader := strings.NewReader(s)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", reader)
	req.Header.Add("Content-Type", "application/cloudevents+json")
	return req
}

// CreateValidBinaryRequestV1Alpha1 creates a valid binary cloudevent as http request.
func CreateValidBinaryRequestV1Alpha1(t *testing.T) *http.Request {
	t.Helper()
	reader := strings.NewReader(`{"foo":"bar"}`)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", reader)
	req.Header.Add("Ce-Specversion", "1.0")
	req.Header.Add("Ce-Type", "sap.kyma.custom.testapp1023.order.created.v1")
	req.Header.Add("Ce-Source", "/default/sap.kyma/id")
	req.Header.Add("Ce-ID", "8945ec08-256b-11eb-9928-acde48001122")
	return req
}

// CreateInvalidBinaryRequestV1Alpha1 creates an invalid binary cloudevent as http request. The `type` is missing.
func CreateInvalidBinaryRequestV1Alpha1(t *testing.T) *http.Request {
	t.Helper()
	reader := strings.NewReader(`{"foo":"bar"}`)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", reader)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Ce-Specversion", "1.0")
	req.Header.Add("Ce-Source", "/default/sap.kyma/id")
	req.Header.Add("Ce-ID", "8945ec08-256b-11eb-9928-acde48001122")
	return req
}

type GenericSenderStub struct {
	SleepDuration time.Duration
	Err           sender.PublishError
	ReceivedEvent *cev2event.Event
	BackendURL    string
}

func (g *GenericSenderStub) Send(_ context.Context, event *cev2event.Event) sender.PublishError {
	g.ReceivedEvent = event
	time.Sleep(g.SleepDuration)
	return g.Err
}

func (g *GenericSenderStub) URL() string {
	return g.BackendURL
}

func NewApplicationListerOrDie(ctx context.Context, appName string) *application.Lister {
	app := applicationtest.NewApplication(appName, nil)
	appLister := fake.NewApplicationListerOrDie(ctx, app)
	return appLister
}
