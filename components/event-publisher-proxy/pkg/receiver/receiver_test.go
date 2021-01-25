package receiver

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

// a mocked http.Handler
type testHandler struct{}

func (h *testHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

var _ http.Handler = (*testHandler)(nil)

func TestNewHttpMessageReceiver(t *testing.T) {
	port, err := testingutils.GeneratePort()
	if err != nil {
		t.Fatalf("failed to generate port: %v", err)
	}
	r := NewHttpMessageReceiver(port)
	if r == nil {
		t.Fatalf("Could not create HttpMessageReceiver")
	}
	if r.port != port {
		t.Errorf("Port should be: %d is: %d", port, r.port)
	}
}

// Test that tht receiver shuts down properly then receiving stop signal
func TestStartListener(t *testing.T) {
	timeout := time.Second * 10
	r := fixtureReceiver()

	ctx := context.Background()
	// used to simulate sending a stop signal
	ctx, cancelFunc := context.WithCancel(ctx)

	// start receiver
	wg := sync.WaitGroup{}
	started := make(chan bool, 1)
	defer close(started)
	go func(t *testing.T) {
		wg.Add(1)
		started <- true
		t.Log("starting receiver in goroutine")
		if err := r.StartListen(ctx, &testHandler{}); err != nil {
			t.Fatalf("error while starting HTTPMessageReceiver: %v", err)
		}
		t.Log("receiver goroutine ends here")
		wg.Done()
	}(t)

	// wait for receiver to start
	<-started

	// stop it
	cancelFunc()
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	// wait for it
	t.Log("Waiting for receiver to stop")
	select {
	// receiver dit shut down properly
	case <-c:
		t.Log("Waiting for receiver to stop [done]")
		break
	// receiver dit shut down in time
	case <-time.Tick(timeout):
		t.Fatalf("Expected receiver to shutdown after timeout: %v\n", timeout)
	}
}

func fixtureReceiver() *HttpMessageReceiver {
	return NewHttpMessageReceiver(0)
}
