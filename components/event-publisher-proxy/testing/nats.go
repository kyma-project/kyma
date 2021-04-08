package testing

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	cev2 "github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func StartNatsServer() *server.Server {
	opts := test.DefaultTestOptions
	opts.Port = server.RANDOM_PORT
	return test.RunServer(&opts)
}

func SubscribeToEventOrFail(t *testing.T, connection *nats.Conn, eventType string, validator nats.MsgHandler) {
	if _, err := connection.Subscribe(eventType, validator); err != nil {
		t.Fatalf("Failed to subscribe to event with error: %v", err)
	}
}

func ValidateNatsSubjectOrFail(t *testing.T, subject string, notify ...chan bool) nats.MsgHandler {
	return func(msg *nats.Msg) {
		for _, n := range notify {
			n <- true
		}
		if msg != nil && msg.Subject != subject {
			t.Errorf("invalid NATS subject, expected [%s] but found [%s]", subject, msg.Subject)
		}
	}
}

func ValidateNatsMessageDataOrFail(t *testing.T, data string, notify ...chan bool) nats.MsgHandler {
	return func(msg *nats.Msg) {
		for _, n := range notify {
			n <- true
		}

		event := cev2.New(cev2.CloudEventsVersionV1)
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			t.Errorf("failed to unmarshal message with error: %v", err)
		}

		if eventData := string(event.Data()); data != eventData {
			t.Errorf("invalid message data, expected [%s] but found [%s]", data, eventData)
		}
	}
}

func WaitForChannelOrTimeout(done chan bool, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout is reached %v", timeout)
	}
}
