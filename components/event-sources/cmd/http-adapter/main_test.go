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

	setEnvironmentVariable(sinkURI, port, t)

	// start http-adapter
	go startAdapter()

	// TODO(nachtmaar): remove sleep
	time.Sleep(5 * time.Second)
	sendEvent(t, adapterURI)
	eventResponse := sendEvent(t, adapterURI)

	// TODO(nachtmaar): validate eventResponse
	t.Log(eventResponse)
	fmt.Println("starting select")

	receiveFromSink(sinkRequests, t)
	receiveFromSink(sinkRequests, t)
}

func receiveFromSink(sinkRequests chan http.Request, t *testing.T) *http.Request {
	select {
	case sinkRequest := <-sinkRequests:
		fmt.Printf("got sink request: %v\n", sinkRequest)
		return &sinkRequest
	case <-time.After(60 * time.Second):
		t.Error("no cloud event received on sink side")
		return nil
	}
}

func sendEvent(t *testing.T, adapterURI string) *cloudevents.Event {
	client, err := kncloudevents.NewDefaultClient(adapterURI)
	if err != nil {
		t.Fatal(err)
	}
	event := cloudevents.New(cloudevents.CloudEventsVersionV1)
	// TODO(nachtmaar): send custom events, e.g. allow malformed events and different content-types
	event.Context.SetSource("foo")
	event.Context.SetType("foo")
	event.Context.SetID("foo")
	t.Logf("sending event to http adapter: %s", event)
	_, eventResponse, err := client.Send(context.Background(), event)
	if err != nil {
		t.Fatal(err)
	}
	return eventResponse
}

func setEnvironmentVariable(sinkURI string, port int, t *testing.T) {
	// set required environment variables
	envs := map[string]string{
		"SINK_URI":           sinkURI,
		"NAMESPACE":          "foo",
		"K_METRICS_CONFIG":   "metrics",
		"K_LOGGING_CONFIG":   "logging",
		"APPLICATION_SOURCE": "varkes",
		// some probably unused port
		"HTTP_PORT": strconv.Itoa(port),
	}
	for k, v := range envs {
		if err := os.Setenv(k, v); err != nil {
			t.Fatal(err)
		}
	}
}

func startAdapter() {
	main()
}
