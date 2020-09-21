package receiver

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"
)

// a mocked http.Handler
type testHandler struct{}

func (h *testHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

var _ http.Handler = (*testHandler)(nil)

func TestNewHttpMessageReceiver(t *testing.T) {
	port := 9091
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
	go func() {
		wg.Add(1)
		t.Log("starting receiver in goroutine")
		if err := r.StartListen(ctx, &testHandler{}); err != nil {
			t.Fatalf("error while starting HTTPMessageReceiver: %v", err)
		}
		t.Log("receiver goroutine ends here")
		wg.Done()
	}()

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
	return NewHttpMessageReceiver(9091)
}
