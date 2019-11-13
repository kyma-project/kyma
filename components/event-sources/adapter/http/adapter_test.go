package http

import (
	"context"
	"net/http"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter"
	"knative.dev/pkg/source"
)

type statsReporter struct {
	eventCount int
}

func (r *statsReporter) ReportEventCount(args *source.ReportArgs, responseCode int) error {
	r.eventCount += 1
	return nil
}

type cloudEventsClient struct {
	sent  bool
	event cloudevents.Event
}

func (c *cloudEventsClient) StartReceiver(ctx context.Context, fn interface{}) error {
	panic("only implemented for mock")
}

func (c *cloudEventsClient) Send(ctx context.Context, event cloudevents.Event) (context.Context, *cloudevents.Event, error) {
	c.sent = true
	c.event = event
	return ctx, &event, nil
}

const (
	ApplicationSource = "application-source"
	Port              = 8080
	SinkURI           = "some URI"
	NameSpace         = "some namespace"
)

const (
	validHTTPMethod = http.MethodPost
	validURI        = "/"
)

func testLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	return logger
}
func getValidEvent() cloudevents.Event {
	validEvent := cloudevents.NewEvent(cloudevents.VersionV1)
	validEvent.SetSource("foo")
	validEvent.SetType("foo")
	validEvent.SetID("foo")
	return validEvent
}

// Given a sink response code and a valid event
// - test the status code sent to the client
// - ensure event count increased by 1
// - cloudevents source got replaced
func TestStatusCodes(t *testing.T) {

	var testsCases = []struct {
		name                 string
		giveSinkResponseCode int
		wantResponseCode     int
	}{
		{
			name:                 "accept CE v1.0 healthy sink",
			giveSinkResponseCode: http.StatusOK,
			wantResponseCode:     http.StatusOK,
		},
		{
			name:                 "accept CE v1.0 sink 2xx",
			giveSinkResponseCode: http.StatusAccepted,
			wantResponseCode:     http.StatusOK,
		},
		{
			name:                 "accept CE v1.0 broken sink",
			giveSinkResponseCode: http.StatusInternalServerError,
			wantResponseCode:     http.StatusInternalServerError,
		},
	}

	for _, tt := range testsCases {
		t.Run(tt.name, func(t *testing.T) {

			logger := testLogger(t)
			defer logger.Sync() // flushes buffer, if any
			ceClient := &cloudEventsClient{}
			statsReporter := &statsReporter{}

			ha := httpAdapter{
				ceClient:      ceClient,
				statsReporter: statsReporter,
				accessor: &envConfig{
					EnvConfig: adapter.EnvConfig{
						SinkURI:   SinkURI,
						Namespace: NameSpace,
					},
					EventSource: ApplicationSource,
					Port:        Port,
				},
				adapterContext: nil,
				logger:         logger,
			}
			er := cloudevents.EventResponse{}

			// set sink status code
			// cloudevents.StartReceiver sets the TransportContext in the adapter
			// for the tests we need to provide our own since we directly call serveHTTP
			tctx := cloudeventshttp.TransportContext{Header: http.Header{}, Method: validHTTPMethod, URI: validURI, StatusCode: tt.giveSinkResponseCode}
			ctx := cloudeventshttp.WithTransportContext(context.Background(), tctx)

			// call adapter with event
			_ = ha.serveHTTP(ctx, getValidEvent(), &er)

			// response code expected
			if tt.wantResponseCode != 0 && tt.wantResponseCode != er.Status {
				t.Errorf("Expected status code: %d, got: %d", tt.wantResponseCode, er.Status)
			}

		})
	}
}

// Test the adapter:
// - sends the event to the sink
// - ensures an event count metric was reported
// - ensure source has been replaced
// - the event that comes in is send to sink except modified source
// - sink will always report 200
func TestServerHTTP_Receive(t *testing.T) {

	var testCases = []struct {
		name                string
		giveEvent           func() cloudevents.Event
		wantResponseCode    int
		wantResponseMessage string
		shouldSendToSink    bool
	}{
		{
			name: "decline CE v0.3",
			giveEvent: func() cloudevents.Event {
				return cloudevents.NewEvent(cloudevents.VersionV03)
			},
			wantResponseCode:    http.StatusBadRequest,
			wantResponseMessage: ErrorResponseCEVersionUnsupported,
			shouldSendToSink:    false,
		},
		{
			name: "decline CE v0.2",
			giveEvent: func() cloudevents.Event {
				return cloudevents.NewEvent(cloudevents.VersionV02)
			},
			wantResponseCode:    http.StatusBadRequest,
			wantResponseMessage: ErrorResponseCEVersionUnsupported,
			shouldSendToSink:    false,
		},
		{
			name: "decline CE v0.1",
			giveEvent: func() cloudevents.Event {
				return cloudevents.NewEvent(cloudevents.VersionV01)
			},
			wantResponseCode:    http.StatusBadRequest,
			wantResponseMessage: ErrorResponseCEVersionUnsupported,
			shouldSendToSink:    false,
		},
		{
			name: "accept valid CE v1.0",
			giveEvent: func() cloudevents.Event {
				event := cloudevents.NewEvent(cloudevents.VersionV1)
				event.SetSource("foo")
				event.SetType("foo")
				event.SetID("foo")
				return event
			},
			wantResponseCode: http.StatusOK,
			shouldSendToSink: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			logger := testLogger(t)
			defer logger.Sync() // flushes buffer, if any
			ceClient := &cloudEventsClient{}
			statsReporter := &statsReporter{}

			ha := httpAdapter{
				ceClient:      ceClient,
				statsReporter: statsReporter,
				accessor: &envConfig{
					EnvConfig: adapter.EnvConfig{
						SinkURI:   SinkURI,
						Namespace: NameSpace,
					},
					EventSource: ApplicationSource,
					Port:        Port,
				},
				adapterContext: nil,
				logger:         logger,
			}
			er := cloudevents.EventResponse{}

			// cloudevents.StartReceiver sets the TransportContext in the adapter
			// for the tests we need to provide our own since we directly call serveHTTP
			tctx := cloudeventshttp.TransportContext{Header: http.Header{}, Method: validHTTPMethod, URI: validURI, StatusCode: http.StatusOK}
			ctx := cloudeventshttp.WithTransportContext(context.Background(), tctx)

			// handle incoming event
			if err := ha.serveHTTP(ctx, tt.giveEvent(), &er); err != nil {
				// response code expected
				if tt.wantResponseCode != 0 && tt.wantResponseCode != er.Status {
					t.Errorf("Expected status code: %d, got: %d", tt.wantResponseCode, er.Status)
				}
				// response message expected
				if tt.wantResponseMessage != "" {
					if tt.wantResponseMessage != err.Error() {
						t.Errorf("Expected status message: %q, got: %q", tt.wantResponseMessage, err.Error())
					}
				}
			}

			isAdapterRespondedWith2XX := er.Status/100 == 2
			ensureAdapterStatusCode(t, er, tt.wantResponseCode)

			if ceClient.sent != tt.shouldSendToSink {
				t.Errorf("The cloudevents client did send to the sink although response code was: %d", er.Status)
			}

			// validations when an event was sent to sink
			if isAdapterRespondedWith2XX {
				if statsReporter.eventCount != 1 {
					t.Errorf("Event metric should be: %d, but is: %d", 1, statsReporter.eventCount)
				}
				if !ceClient.sent {
					t.Errorf("The cloudevents client did not send to the sink")
				}

				ensureSameExpectSource(t, tt.giveEvent(), ceClient.event, ApplicationSource)

				// CE version that comes in should also go out
				if tt.giveEvent().Context.GetSpecVersion() != er.Event.Context.GetSpecVersion() {
					t.Errorf("Event response should be CE %q, got: %q", tt.giveEvent().Context.GetSpecVersion(), er.Event.Context.GetSpecVersion())
				}
			} else {
				if statsReporter.eventCount != 0 {
					t.Errorf("Event metric was reported even though response code was: %d", er.Status)
				}
			}
		})
	}
}

// Ensure that the event which is sent to the sink only differs in the source field
func ensureSameExpectSource(t *testing.T, in cloudevents.Event, out cloudevents.Event, expectedSource string) {
	t.Helper()

	if out.Context.GetSource() != expectedSource {
		t.Errorf("The http adapter did not enrich the event with the source, expected: %q, got: %q", in.Context.GetSource(), expectedSource)
	}
	// required fields
	isSameSpec := in.Context.GetSpecVersion() == out.Context.GetSpecVersion()
	isSameID := in.Context.GetID() == out.Context.GetID()
	isSameType := in.Context.GetType() == out.Context.GetType()
	// optional fields
	isSameDataContentType := in.Context.GetDataContentType() == out.Context.GetDataContentType()
	isSameTime := in.Context.GetTime() == out.Context.GetTime()
	isSameDataSchema := in.Context.GetDataSchema() == out.Context.GetDataSchema()
	isSameSubject := in.Context.GetSubject() == out.Context.GetSubject()

	if !isSameSpec || !isSameDataContentType || !isSameID || !isSameTime || !isSameType || !isSameDataSchema || !isSameSubject {
		t.Errorf("Incoming event (to adapter) and outgoing (to sink) should be the same expect source. Incoming: %+v, Outgoing: %+v", in, out)
	}
}

// check adapter response status code
func ensureAdapterStatusCode(t *testing.T, er cloudevents.EventResponse, expectedStatusCode int) {
	if er.Status != expectedStatusCode {
		t.Errorf("Unexpected status code, expected: %d, got: %d", expectedStatusCode, er.Status)
	}
}
