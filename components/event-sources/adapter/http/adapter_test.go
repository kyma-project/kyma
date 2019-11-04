package http

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter"
	"knative.dev/pkg/source"
	"net/http"
	"testing"
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

var testCases = []struct {
	name                string
	giveEvent           cloudevents.Event
	giveSinkReponseCode int
	wantResponseCode    int
}{
	{
		name:      "decline CE v0.3",
		giveEvent: cloudevents.NewEvent(cloudevents.VersionV03),
		// not required
		giveSinkReponseCode: 0,
		wantResponseCode:    http.StatusBadRequest,
	},
	{
		name:      "decline CE v0.2",
		giveEvent: cloudevents.NewEvent(cloudevents.VersionV02),
		// not required
		giveSinkReponseCode: 0,
		wantResponseCode:    http.StatusBadRequest,
	},
	{
		name:      "decline CE v0.1",
		giveEvent: cloudevents.NewEvent(cloudevents.VersionV01),
		// not required
		giveSinkReponseCode: 0,
		wantResponseCode:    http.StatusBadRequest,
	},
	{
		name:                "accept CE v1.0",
		giveEvent:           cloudevents.NewEvent(cloudevents.VersionV1),
		giveSinkReponseCode: http.StatusOK,
		wantResponseCode:    http.StatusOK,
	},
}

func TestServerHTTP_Succeed(t *testing.T) {

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			logger := testLogger(t)
			defer logger.Sync() // flushes buffer, if any
			ceClient := &cloudEventsClient{}
			statsReporter := &statsReporter{}

			ha := httpAdapter{
				ceClient:      ceClient,
				statsReporter: statsReporter,
				envConfig: &envConfig{
					EnvConfig: adapter.EnvConfig{
						SinkURI:           SinkURI,
						Namespace:         NameSpace,
						MetricsConfigJson: "",
						LoggingConfigJson: "",
					},
					ApplicationSource: ApplicationSource,
					Port:              Port,
				},
				adapterContext: nil,
				logger:         logger,
			}
			er := cloudevents.EventResponse{}

			// cloudevents.StartReceiver sets the TransportContext in the adapter
			// for the tests we need to provide our own since we directly call serveHTTP
			tctx := cloudeventshttp.TransportContext{Header: http.Header{}, Method: validHTTPMethod, URI: validURI, StatusCode: tt.giveSinkReponseCode}
			ctx := cloudeventshttp.WithTransportContext(context.Background(), tctx)

			// handle incoming event
			if err := ha.serveHTTP(ctx, tt.giveEvent, &er); err != nil {
				t.Fatal(err)
			}

			isValidResponse := er.Status/100 == 2

			// check mocks
			if er.Status != tt.wantResponseCode {
				t.Errorf("Unexpected status code, expected: %d, got: %d", tt.wantResponseCode, er.Status)
			}

			// validations when an event was sent to sink
			if isValidResponse {
				if statsReporter.eventCount != 1 {
					t.Errorf("Event metric should be: %d, but is: %d", 1, statsReporter.eventCount)
				}
				if !ceClient.sent {
					t.Errorf("The cloudevents client did not send to the sink")
				}
				if ceClient.event.Context.GetSource() != ApplicationSource {
					t.Errorf("The http adapter did not enrich the event with the source, expected: %q, got: %q", ceClient.event.Context.GetSource(), ApplicationSource)
				}

				// CE version that comes in should also go out
				if tt.giveEvent.Context.GetSpecVersion() != er.Event.Context.GetSpecVersion() {
					t.Errorf("Event response should be CE %q, got: %q", tt.giveEvent.Context.GetSpecVersion(), er.Event.Context.GetSpecVersion())
				}
			} else {
				if statsReporter.eventCount != 0 {
					t.Errorf("Event metric was reported even though response code was: %d", er.Status)
				}
				if ceClient.sent {
					t.Errorf("The cloudevents client did send to the sink although response code was: %d", er.Status)
				}

			}
		})
	}
}
