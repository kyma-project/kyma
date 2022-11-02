package receiver

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"

	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

// a mocked http.Handler
type testHandler struct{}

func (h *testHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

var _ http.Handler = (*testHandler)(nil)

func TestNewHttpMessageReceiver(t *testing.T) {
	port := testingutils.GeneratePortOrDie()
	r := NewHTTPMessageReceiver(port)
	if r.Port != port {
		t.Errorf("Port should be: %d is: %d", port, r.Port)
	}
}

// Test that the receiver shutdown when receiving stop signal
func TestStartListener(t *testing.T) {
	timeout := time.Second * 10
	r := fixtureReceiver()
	mockedLogger, _ := logger.New("json", "info")
	ctx := context.Background()

	// used to simulate sending a stop signal
	ctx, cancelFunc := context.WithCancel(ctx)

	// start receiver
	wg := sync.WaitGroup{}
	start := make(chan bool, 1)
	defer close(start)
	wg.Add(1)
	go func(t *testing.T) {
		defer wg.Done()
		start <- true
		t.Log("starting receiver in goroutine")
		if err := r.StartListen(ctx, &testHandler{}, mockedLogger); err != nil {
			t.Errorf("error while starting HTTPMessageReceiver: %v", err)
		}
		t.Log("receiver goroutine ends here")
	}(t)

	// wait for goroutine to start
	<-start

	// stop it
	cancelFunc()
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	t.Log("Waiting for receiver to stop")
	select {
	// receiver shutdown properly
	case <-c:
		t.Log("Waiting for receiver to stop [done]")
		break
	// receiver shutdown in time
	case <-time.Tick(timeout):
		t.Fatalf("Expected receiver to shutdown after timeout: %v\n", timeout)
	}
}

func fixtureReceiver() *HTTPMessageReceiver {
	return &HTTPMessageReceiver{Port: 0, Host: "localhost"}
}
