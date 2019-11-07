package http

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	cloudevents "github.com/cloudevents/sdk-go"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter"
	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/source"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const startPort = 54321

var tests = []struct {
	name         string
	giveEvent    func() cloudevents.Event
	giveEncoding cloudevents.HTTPEncoding
	// the expected status code and error message then sending to the adapter with cloudevents sdk
	// if none given, assume sending adapter was ok
	wantAdapterError string
}{
	{
		name: "accepts CE v1.0 binary",
		giveEvent: func() cloudevents.Event {
			event := cloudevents.NewEvent(cloudevents.VersionV1)
			_ = event.Context.SetType("foo")
			_ = event.Context.SetID("foo")
			// event.Context.SetSource("will be replaced by adapter anyways, but we need a valid event here")
			_ = event.Context.SetSource("foo")
			return event
		},
		giveEncoding: cloudevents.HTTPBinaryV1,
	},
	{
		name: "accepts CE v1.0 structured",
		giveEvent: func() cloudevents.Event {
			event := cloudevents.NewEvent(cloudevents.VersionV1)
			_ = event.Context.SetType("foo")
			_ = event.Context.SetID("foo")
			// event.Context.SetSource("will be replaced by adapter anyways, but we need a valid event here")
			_ = event.Context.SetSource("foo")
			return event
		},
		giveEncoding: cloudevents.HTTPStructuredV1,
	},
	{
		name: "declines CE < v1.0 binary",
		giveEvent: func() cloudevents.Event {
			event := cloudevents.NewEvent(cloudevents.VersionV03)
			_ = event.Context.SetSpecVersion(cloudevents.VersionV03)
			_ = event.Context.SetType("foo")
			_ = event.Context.SetID("foo")
			// event.Context.SetSource("will be replaced by adapter anyways, but we need a valid event here")
			_ = event.Context.SetSource("foo")
			return event
		},
		wantAdapterError: "error sending cloudevent: 400 Bad Request",
		giveEncoding:     cloudevents.HTTPBinaryV03,
	},
	{
		name: "declines CE < v1.0 structured",
		giveEvent: func() cloudevents.Event {
			event := cloudevents.NewEvent(cloudevents.VersionV03)
			_ = event.Context.SetSpecVersion(cloudevents.VersionV03)
			_ = event.Context.SetType("foo")
			_ = event.Context.SetID("foo")
			// event.Context.SetSource("will be replaced by adapter anyways, but we need a valid event here")
			_ = event.Context.SetSource("foo")
			return event
		},
		wantAdapterError: "error sending cloudevent: 400 Bad Request",
		giveEncoding:     cloudevents.HTTPStructuredV03,
	},
}

type handler struct {
	requests []*http.Request
}

func (h *handler) startSink(t *testing.T) string {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("received sink request")
		if _, err := fmt.Fprintln(w, "Hello, cloudevents client"); err != nil {
			t.Error(err)
		}
		h.requests = append(h.requests, r)
	}))
	sinkURI := ts.URL

	return sinkURI
}

type config struct {
	sinkURI       string
	namespace     string
	metricsConfig string
	loggingConfig string
	source        string
	port          int
}

func (c config) GetSinkURI() string {
	return c.sinkURI
}

func (c config) GetNamespace() string {
	return c.namespace
}

func (c config) GetMetricsConfigJson() string {
	return c.metricsConfig
}

func (c config) GetLoggingConfigJson() string {
	return c.loggingConfig
}

func (c config) GetSource() string {
	return c.source
}

func (c config) GetPort() int {
	return c.port
}

// TestAdapter tests the http-adapter by
// - spinning up the adapter
// - sending a CE event
// - receiving the CE event enriched by application source from adapter using a mocked server in the test
// - the sinkURI is set to the mocked http server
func TestAdapter(t *testing.T) {
	// TODO(nachtmaar):
	return
	t.Parallel()

	for idx, tt := range tests {
		// https://gist.github.com/posener/92a55c4cd441fc5e5e85f27bca008721#how-to-solve-this
		tt := tt
		idx := idx

		t.Logf("running test %s", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// receive channel for http.Request from sink
			handler := handler{
				requests: []*http.Request{},
			}

			sinkURI := handler.startSink(t)
			t.Logf("sink URI: %q", sinkURI)

			adapterPort := startPort + idx
			adapterURI := fmt.Sprintf("http://localhost:%d", adapterPort)

			c := config{
				sinkURI:   sinkURI,
				namespace: "foo",
				// TODO(nachtmaar):
				metricsConfig: "",
				loggingConfig: "",

				source: "guenther",
				port:   adapterPort,
			}

			// start http-adapter
			startHttpAdapter(t, c, context.Background())

			waitAdapterReady(t, adapterURI)
			eventResponse, err := sendEvent(t, adapterURI, tt.giveEvent(), tt.giveEncoding)
			// TODO(nachtmaar):
			fmt.Println(eventResponse)
			t.Log("waiting for sink response")

			if tt.wantAdapterError != "" {
				if err == nil || err.Error() != tt.wantAdapterError {
					t.Fatalf("Expected the cloudevents error to be: %q, but got: %q", tt.wantAdapterError, err)
				} else {
					// done with testing
					return
				}
			}

			// TODO: validate sink request: trace headers etc ...
			if len(handler.requests) != 1 {
				t.Fatalf("Exactly one sink request expected, got: %d", len(handler.requests))
			}
			sinkRequest := handler.requests[0]

			t.Log("ensure source set on event")
			ensureSourceSet(t, sinkRequest, c.GetSource())

			t.Logf("test %q done", tt.name)
		})
	}
	fmt.Println("waiting for tests to complete")

	fmt.Println("tests end")
}

// use readiness probe to ensure adapter is ready
func waitAdapterReady(t *testing.T, adapterURI string) {
	t.Helper()
	if err := retry.Do(
		func() error {
			resp, err := http.Get(adapterURI + readinessReadiness)
			if err != nil {
				return err
			}
			expectedStatusCode := 200
			if resp.StatusCode != expectedStatusCode {
				return fmt.Errorf("adapter is not ready, expected status code: %d, got: %d", expectedStatusCode, resp.StatusCode)
			}
			return err
		},
	); err != nil {
		t.Fatalf("timeout waiting for adapter readiness: %v", err)
	}
}

// startHttpAdapter starts the adapter with a cloudevents client configured with the test sink as target
func startHttpAdapter(t *testing.T, c config, ctx context.Context) *adapter.Adapter {
	sinkClient, err := kncloudevents.NewDefaultClient(c.GetSinkURI())
	if err != nil {
		t.Fatal("error building cloud event client", zap.Error(err))
	}
	statsReporter, err := source.NewStatsReporter()
	if err != nil {
		t.Errorf("error building statsreporter: %v", err)
	}
	// TODO(nachtmaar): validate metrics reporter called
	httpAdapter := NewAdapter(ctx, c, sinkClient, statsReporter)
	go func() {
		if err := httpAdapter.Start(ctx.Done()); err != nil {
			t.Errorf("start returned an error: %v", err)
		}
	}()
	return &httpAdapter
}

// ensureSourceSet checks that the http adapter sets the event source on the event which is sent to the sink
func ensureSourceSet(t *testing.T, sinkReponse *http.Request, wantEventSource string) {
	t.Helper()
	giveEventSource := sinkReponse.Header.Get("CE-Source")
	if giveEventSource != wantEventSource {
		t.Errorf("Adapter is supposed to set the event source to: %q, got: %q", wantEventSource, giveEventSource)
	}
}

// sendEvent sends a cloudevent to the adapter
// returns an error when not getting status code 2xx
func sendEvent(t *testing.T, adapterURI string, event cloudevents.Event, encoding cloudevents.HTTPEncoding) (*cloudevents.Event, error) {
	t.Helper()
	transport, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(adapterURI),
		cloudevents.WithEncoding(encoding),
	)
	if err != nil {
		return nil, err
	}
	client, err := cloudevents.NewClient(transport)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	t.Logf("sending event to http adapter, event: %s", event)
	// NOTE: then using CE sdk to send an event we get error message and status code in one message: "error sending cloudevent: 400 Bad Request"
	_, eventResponse, err := client.Send(ctx, event)
	if err != nil {
		return nil, err
	}
	return eventResponse, nil
}

func TestAdapterShutdown(t *testing.T) {
	c := config{}

	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)

	httpAdapter := NewAdapter(ctx, c, nil, nil)
	stopChannel := make(chan error)

	// start adapter
	go func() {
		t.Log("starting http adapter in goroutine")
		err := httpAdapter.Start(ctx.Done())
		stopChannel <- err
		t.Log("http adapter goroutine ends here")
	}()

	t.Log("simulate stop signal")
	// call close on internal ctx.Done() channel
	cancelFunc()

	t.Log("waiting for adapter to stop")

	timeout := time.Second * 10
	select {
	case err := <-stopChannel:
		if err != nil {
			t.Fatalf("Expected adapter shutdown to return no error, got: %v\n", err)
		}
	case <-time.Tick(timeout):
		t.Fatalf("Expected adapter to shutdown after timeout: %v\n", timeout)
	}

	t.Log("waiting for adapter to stop [done]")
}
