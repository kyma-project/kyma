//nolint:lll // output directly from prometheus
package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/legacy/legacytest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender/common"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender/jetstream"

	eclogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/cleaner"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/applicationtest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/application/fake"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/builder"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/cloudevents/eventtype/eventtypetest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/histogram/mocks"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/metricstest"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/sender"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func TestHandler_publishCloudEvents(t *testing.T) {
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
			name: "Publish structured Cloudevent for Subscription v1alpha1",
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
			name: "Publish binary Cloudevent for Subscription v1alpha1",
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
				request: CreateValidStructuredRequest(t),
			},
			wantStatus: 204,
			wantTEF: metricstest.MakeTEFBackendDuration(204, "FOO") +
				metricstest.MakeTEFBackendRequests(204, "FOO") +
				metricstest.MakeTEFEventTypePublished(204, "testapp1023", "order.created.v1"),
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
				request: CreateValidBinaryRequest(t),
			},
			wantStatus: 204,
			wantTEF: metricstest.MakeTEFBackendDuration(204, "FOO") +
				metricstest.MakeTEFBackendRequests(204, "FOO") +
				metricstest.MakeTEFEventTypePublished(204, "testapp1023", "order.created.v1"),
		},
		{
			name: "Publish invalid structured CloudEvent",
			fields: fields{
				Sender:           &GenericSenderStub{},
				collector:        metrics.NewCollector(latency),
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateInvalidStructuredRequest(t),
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
				request: CreateInvalidBinaryRequest(t),
			},
			wantStatus: 400,
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
				request: CreateValidBinaryRequest(t),
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
					Err: jetstream.ErrNoSpaceLeftOnDevice,
				},
				collector:        metrics.NewCollector(latency),
				eventTypeCleaner: &eventtypetest.CleanerStub{},
			},
			args: args{
				request: CreateValidBinaryRequest(t),
			},
			wantStatus: 507,
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

			app := applicationtest.NewApplication("appName1", nil)
			appLister := fake.NewApplicationListerOrDie(context.Background(), app)

			ceBuilder := builder.NewGenericBuilder("prefix", cleaner.NewJetStreamCleaner(logger), appLister, logger)

			h := &Handler{
				Sender:             tt.fields.Sender,
				Logger:             logger,
				collector:          tt.fields.collector,
				eventTypeCleaner:   tt.fields.eventTypeCleaner,
				ceBuilder:          ceBuilder,
				Options:            &options.Options{},
				OldEventTypePrefix: testingutils.OldEventTypePrefix,
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

func TestHandler_publishLegacyEventsAsCE(t *testing.T) {
	// define common given variables
	appLister := NewApplicationListerOrDie(context.Background(), "testapp")

	// set mock for latency metrics
	latency := new(mocks.BucketsProvider)
	latency.On("Buckets").Return(nil)

	tests := []struct {
		name                   string
		givenSender            sender.GenericSender
		givenLegacyTransformer legacy.RequestToCETransformer
		givenCollector         metrics.PublishingMetricsCollector
		givenRequest           *http.Request
		wantHTTPStatus         int
		wantTEF                string
	}{
		{
			name: "Send valid legacy event",
			givenSender: &GenericSenderStub{
				BackendURL: "FOO",
			},
			givenLegacyTransformer: legacy.NewTransformer("namespace", "im.a.prefix", appLister),
			givenCollector:         metrics.NewCollector(latency),
			givenRequest:           legacytest.ValidLegacyRequestOrDie(t, "v1", "testapp", "object.created"),
			wantHTTPStatus:         http.StatusOK,
			wantTEF: metricstest.MakeTEFBackendDuration(204, "FOO") +
				metricstest.MakeTEFBackendRequests(204, "FOO") +
				metricstest.MakeTEFEventTypePublished(204, "testapp", "object.created.v1"),
		},
		{
			name: "Send valid legacy event but cannot send to backend due to target not found (e.g. stream missing)",
			givenSender: &GenericSenderStub{
				Err:        common.ErrBackendTargetNotFound,
				BackendURL: "FOO",
			},
			givenLegacyTransformer: legacy.NewTransformer("namespace", "im.a.prefix", appLister),
			givenCollector:         metrics.NewCollector(latency),
			givenRequest:           legacytest.ValidLegacyRequestOrDie(t, "v1", "testapp", "object.created"),
			wantHTTPStatus:         http.StatusNotFound,
			wantTEF: metricstest.MakeTEFBackendDuration(404, "FOO") +
				metricstest.MakeTEFBackendRequests(404, "FOO") +
				metricstest.MakeTEFBackendErrors(),
		},
		{
			name: "Send valid legacy event but cannot send to backend due to full storage",
			givenSender: &GenericSenderStub{
				Err:        common.ErrInsufficientStorage,
				BackendURL: "FOO",
			},
			givenLegacyTransformer: legacy.NewTransformer("namespace", "im.a.prefix", appLister),
			givenCollector:         metrics.NewCollector(latency),
			givenRequest:           legacytest.ValidLegacyRequestOrDie(t, "v1", "testapp", "object.created"),
			wantHTTPStatus:         507,
			wantTEF: metricstest.MakeTEFBackendDuration(507, "FOO") +
				metricstest.MakeTEFBackendRequests(507, "FOO") +
				metricstest.MakeTEFBackendErrors(),
		},
		{
			name: "Send valid legacy event but cannot send to backend",
			givenSender: &GenericSenderStub{
				Err:        common.BackendPublishError{},
				BackendURL: "FOO",
			},
			givenLegacyTransformer: legacy.NewTransformer("namespace", "im.a.prefix", appLister),
			givenCollector:         metrics.NewCollector(latency),
			givenRequest:           legacytest.ValidLegacyRequestOrDie(t, "v1", "testapp", "object.created"),
			wantHTTPStatus:         500,
			wantTEF: metricstest.MakeTEFBackendDuration(500, "FOO") +
				metricstest.MakeTEFBackendRequests(500, "FOO") +
				metricstest.MakeTEFBackendErrors(),
		},
		{
			name: "Send invalid legacy event",
			givenSender: &GenericSenderStub{
				BackendURL: "FOO",
			},
			givenLegacyTransformer: legacy.NewTransformer("namespace", "im.a.prefix", appLister),
			givenCollector:         metrics.NewCollector(latency),
			givenRequest:           legacytest.InvalidLegacyRequestOrDie(t, "v1", "testapp", "object.created"),
			wantHTTPStatus:         400,
			// this is a client error. We do record an error metric for requests that cannot even be decoded correctly.
			wantTEF: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			logger, err := eclogger.New("text", "debug")
			require.NoError(t, err)

			ceBuilder := builder.NewGenericBuilder("prefix", cleaner.NewJetStreamCleaner(logger), appLister, logger)

			h := &Handler{
				Sender:            tt.givenSender,
				Logger:            logger,
				LegacyTransformer: tt.givenLegacyTransformer,
				collector:         tt.givenCollector,
				ceBuilder:         ceBuilder,
				Options:           &options.Options{},
			}
			writer := httptest.NewRecorder()

			// when
			h.publishLegacyEventsAsCE(writer, tt.givenRequest)

			// then
			require.Equal(t, tt.wantHTTPStatus, writer.Result().StatusCode)
			body, err := io.ReadAll(writer.Result().Body)
			require.NoError(t, err)

			if tt.wantHTTPStatus == http.StatusOK {
				ok := &api.PublishResponse{}
				err = json.Unmarshal(body, ok)
				require.NoError(t, err)
			} else {
				nok := &api.Error{}
				err = json.Unmarshal(body, nok)
				require.NoError(t, err)
			}

			metricstest.EnsureMetricMatchesTextExpositionFormat(t, h.collector, tt.wantTEF)
		})
	}
}

// CreateValidStructuredRequestV1Alpha2 creates a structured cloudevent as http request.
func CreateValidStructuredRequest(t *testing.T) *http.Request {
	t.Helper()
	reader := strings.NewReader(`{
		"specversion":"1.0",
		"type":"order.created.v1",
		"source":"testapp1023",
		"id":"8945ec08-256b-11eb-9928-acde48001122",
		"data":{
			"foo":"bar"
		}
		}`)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", reader)
	req.Header.Add("Content-Type", "application/cloudevents+json")
	return req
}

// CreateBrokenRequest creates a structured cloudevent request that cannot be parsed.
func CreateBrokenRequest(t *testing.T) *http.Request {
	t.Helper()
	reader := strings.NewReader("I AM JUST A BROKEN REQUEST")
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", reader)
	req.Header.Add("Content-Type", "application/cloudevents+json")
	return req
}

// CreateInvalidStructuredRequestV1Alpha2 creates an invalid structured cloudevent as http request.
// The `type` is missing.
func CreateInvalidStructuredRequest(t *testing.T) *http.Request {
	t.Helper()
	reader := strings.NewReader(`{
		"specversion":"1.0",
		"source":"testapp1023",
		"id":"8945ec08-256b-11eb-9928-acde48001122",
		"data":{
			"foo":"bar"
		}}`)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", reader)
	req.Header.Add("Content-Type", "application/cloudevents+json")
	return req
}

// CreateValidBinaryRequestV1Alpha2 creates a valid binary cloudevent as http request.
func CreateValidBinaryRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", strings.NewReader(`{"foo":"bar"}`))
	req.Header.Add("Ce-Specversion", "1.0")
	req.Header.Add("Ce-Type", "order.created.v1")
	req.Header.Add("Ce-Source", "testapp1023")
	req.Header.Add("Ce-ID", "8945ec08-256b-11eb-9928-acde48001122")
	return req
}

// CreateInvalidBinaryRequestV1Alpha2 creates an invalid binary cloudevent as http request. The `type` is missing.
func CreateInvalidBinaryRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://localhost/publish", strings.NewReader(`{"foo":"bar"}`))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Ce-Specversion", "1.0")
	req.Header.Add("Ce-Source", "testapp1023")
	req.Header.Add("Ce-ID", "8945ec08-256b-11eb-9928-acde48001122")
	return req
}
