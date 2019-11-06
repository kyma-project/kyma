package main

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	httpadapter "github.com/kyma-project/kyma/components/event-sources/adapter/http"
	"knative.dev/eventing/pkg/adapter"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

const EventSource = "somesource"
const Port = 54321

var tests = []struct {
	name           string
	giveEvent      func() cloudevents.Event
	wantStatusCode int
}{
	{
		name: "accepts CE v1.0",
		giveEvent: func() cloudevents.Event {
			event := cloudevents.NewEvent(cloudevents.VersionV1)
			event.Context.SetType("foo")
			event.Context.SetID("foo")
			//event.Context.SetSource("will be replaced by adapter anyways, but we need a valid event here")
			event.Context.SetSource("foo")
			return event
		},
		wantStatusCode: http.StatusOK,
	},
	{
		name: "declines CE < v1.0",
		giveEvent: func() cloudevents.Event {
			event := cloudevents.NewEvent(cloudevents.VersionV03)
			event.Context.SetType("foo")
			event.Context.SetID("foo")
			//event.Context.SetSource("will be replaced by adapter anyways, but we need a valid event here")
			event.Context.SetSource("foo")
			return event
		},
		wantStatusCode: http.StatusBadRequest,
	},
}

type handler struct {
	requests []*http.Request
}

func (h handler) startSink(t *testing.T) string {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("received sink request")
		if _, err := fmt.Fprintln(w, "Hello, cloudevents client"); err != nil {
			t.Error(err)
		}
		h.requests = append(h.requests, r)
	}))
	sinkURI := ts.URL

	//sinkURIChan := make(chan string)
	//// start mock sink
	//t.Log("starting sink")
	//go func() {
	//	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		t.Log("received sink request")
	//		if _, err := fmt.Fprintln(w, "Hello, cloudevents client"); err != nil {
	//			t.Error(err)
	//		}
	//		sinkRequests <- *r
	//	}))
	//	sinkURIChan <- ts.URL
	//	//defer ts.Close()
	//}()
	//sinkURI := <-sinkURIChan
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

	//testsReady := make(chan bool, len(tests))
	var wg sync.WaitGroup

	for idx, tt := range tests {
		t.Logf("running test %s", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			wg.Add(1)

			adapterPort := Port + idx
			adapterURI := fmt.Sprintf("http://localhost:%d", adapterPort)

			// receive channel for http.Request from sink
			handler := handler{}

			sinkURI := handler.startSink(t)
			t.Logf("sink URI: %q", sinkURI)

			//setEnvironmentVariables(t, sinkURI, Port)
			c := config{
				sinkURI:   sinkURI,
				namespace: "foo",
				// TODO(nachtmaar):
				metricsConfig: "",
				loggingConfig: "",

				source: "guenther",
				port:   adapterPort,
			}
			hector := func() adapter.EnvConfigAccessor { return &c }

			// start http-adapter
			go adapter.Main("application-source", hector, httpadapter.NewAdapter)

			// TODO(nachtmaar): remove sleep by using readiness probe
			time.Sleep(10 * time.Second)
			sendEvent(t, adapterURI, tt.giveEvent())
			//t.Logf("received event response, event: %v", eventResponse)
			//
			//// TODO(nachtmaar): validate eventResponse
			//t.Logf("event response: %v", eventResponse)
			t.Log("waiting for sink response")

			// TODO: validate sink request: trace headers etc ...
			sinkRequest := handler.requests[0]

			if len(handler.requests) > 1 {
				t.Errorf("Only one sink request expected, got: %d", len(handler.requests))
			}

			//sinkRequest := receiveFromSink(t, sinkRequests)
			t.Log("ensure source set on event")
			ensureSourceSet(t, sinkRequest, EventSource)
			//testsReady <- true
		})
	}
	//wg.Wait()
	//for _, tt := range tests {
	//	fmt.Printf("waiting for test: %q\n", tt.name)
	//	<-testsReady
	//	fmt.Printf("waiting for test: %q[done]\n", tt.name)
	//}

	fmt.Println("test end")
}

// ensureSourceSet checks that the http adapter sets the event source on the event which is sent to the sink
func ensureSourceSet(t *testing.T, sinkReponse *http.Request, wantEventSource string) {
	t.Helper()
	giveEventSource := sinkReponse.Header.Get("CE-Source")
	if giveEventSource != wantEventSource {
		t.Errorf("Adapter is supposed to set the event source to: %q, got: %q", wantEventSource, giveEventSource)
	}
}

// receiveFromSink receives a http request which was send to the sink by the adapter
// it receives the request from a channel
func receiveFromSink(t *testing.T, sinkRequests chan http.Request) *http.Request {
	t.Helper()
	select {
	case sinkRequest := <-sinkRequests:
		t.Logf("got sink request: %v\n", sinkRequest)
		return &sinkRequest
	case <-time.After(5 * time.Second):
		t.Fatal("no cloud event received on sink side")
		return nil
	}
}

// sendEvent sends a cloudevent to the adapter
func sendEvent(t *testing.T, adapterURI string, event cloudevents.Event) *cloudevents.Event {
	t.Helper()
	transport, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(adapterURI),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV1),
	)
	if err != nil {
		t.Fatal(err)
	}
	client, err := cloudevents.NewClient(transport)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("sending event to http adapter, event: %s", event)
	_, eventResponse, err := client.Send(context.Background(), event)
	if err != nil {
		fmt.Println(err.Error())
		t.Fatal(err)
	}
	return eventResponse
}

//// setEnvironmentVariables sets all environment variables required by the http adapter
//func setEnvironmentVariables(t *testing.T, sinkURI string, port int) {
//	t.Helper()
//	// set required environment variables
//	envs := map[string]string{
//		"SINK_URI":           sinkURI,
//		"NAMESPACE":          "foo",
//		"K_METRICS_CONFIG":   "metrics",
//		"K_LOGGING_CONFIG":   "logging",
//		"APPLICATION_SOURCE": EventSource,
//		// some probably unused port
//		"PORT": strconv.Itoa(port),
//	}
//	for k, v := range envs {
//		if err := os.Setenv(k, v); err != nil {
//			t.Fatal(err)
//		}
//	}
//}

