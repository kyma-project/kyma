package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"knative.dev/eventing/pkg/kncloudevents"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

const EventSource = "somesource"

// TestAdapter tests the http-adapter by
// - spinning up the adapter
// - sending a CE event
// - receiving the CE event enriched by application source from adapter using a mocked server in the test
// - the sinkURI is set to the mocked http server
func TestAdapter(t *testing.T) {
	port := 54321
	adapterURI := fmt.Sprintf("http://localhost:%d", port)

	// receive channel for http.Request from sink
	sinkRequests := make(chan http.Request, 2)
	// start mock sink
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received sink request")
		if _, err := fmt.Fprintln(w, "Hello, cloudevents client"); err != nil {
			t.Error(err)
		}
		sinkRequests <- *r
	}))
	defer ts.Close()
	sinkURI := ts.URL

	setEnvironmentVariables(t, sinkURI, port)

	// start http-adapter
	go startAdapter()

	// TODO(nachtmaar): remove sleep by using readiness probe
	time.Sleep(5 * time.Second)

	event := cloudevents.New(cloudevents.CloudEventsVersionV1)
	// TODO(nachtmaar): send custom events, e.g. allow malformed events and different content-types
	event.Context.SetSource("foo")
	event.Context.SetType("foo")
	event.Context.SetID("foo")

	eventResponse := sendEvent(t, adapterURI, event)
	t.Logf("received event response: %v", eventResponse)

	// TODO(nachtmaar): validate eventResponse
	t.Log(eventResponse)
	fmt.Println("waiting for sink response")

	// TODO: validate sink request: trace headers etc ...
	sinkRequest := receiveFromSink(t, sinkRequests)
	ensureSourceSet(t, sinkRequest, EventSource)
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
		fmt.Printf("got sink request: %v\n", sinkRequest)
		return &sinkRequest
	case <-time.After(60 * time.Second):
		t.Error("no cloud event received on sink side")
		return nil
	}
}

// sendEvent sends a cloudevent to the adapter
func sendEvent(t *testing.T, adapterURI string, event cloudevents.Event) *cloudevents.Event {
	t.Helper()
	client, err := kncloudevents.NewDefaultClient(adapterURI)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("sending event to http adapter: %s", event)
	_, eventResponse, err := client.Send(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}
	return eventResponse
}

// setEnvironmentVariables sets all environment variables required by the http adapter
func setEnvironmentVariables(t *testing.T, sinkURI string, port int) {
	t.Helper()
	// set required environment variables
	envs := map[string]string{
		"SINK_URI":           sinkURI,
		"NAMESPACE":          "foo",
		"K_METRICS_CONFIG":   "metrics",
		"K_LOGGING_CONFIG":   "logging",
		"APPLICATION_SOURCE": EventSource,
		// some probably unused port
		"PORT": strconv.Itoa(port),
	}
	for k, v := range envs {
		if err := os.Setenv(k, v); err != nil {
			t.Fatal(err)
		}
	}
}

// startAdapter starts the http adapter
func startAdapter() {
	main()
}
