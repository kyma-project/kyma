package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"knative.dev/eventing/pkg/kncloudevents"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestAdapter tests the http-adapter by
// - spinning up the adapter
// - sending a CE event
// - receiving the CE event enriched by application source from adapter using a mocked server in the test
// - the sinkURI is set to the mocked http server
func TestAdapter(t *testing.T) {
	var wg sync.WaitGroup
	//sinkURI := "https://dummy-kn-channel.default.svc.cluster.local"
	sinkURI := "http://localhost:55555"
	port := 54321
	adapterURI := fmt.Sprintf("http://localhost:%d", port)

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

	wg.Add(1)
	go fakeMain(&wg)

	client, err := kncloudevents.NewDefaultClient(adapterURI)
	if err != nil {
		t.Fatal(err)
	}
	// TODO(nachtmaar): remove sleep
	time.Sleep(5 * time.Second)
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
	// TODO(nachtmaar): validate eventResponse
	t.Log(eventResponse)

	// TODO(nachtmaar): stop process when all tests completed
	wg.Wait()
}

func fakeMain(wg *sync.WaitGroup) {
	main()
	wg.Done()
}
