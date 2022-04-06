package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestJetstreamMessageSender(t *testing.T) {
	testCases := []struct {
		name                      string
		givenStream               bool
		givenNatsConnectionClosed bool
		wantError                 bool
		wantStatusCode            int
	}{
		{
			name:                      "send in jetstream mode should not succeed if stream doesn't exist",
			givenStream:               false,
			givenNatsConnectionClosed: false,
			wantError:                 true,
			wantStatusCode:            http.StatusNotFound,
		},
		{
			name:                      "send in jetstream mode should succeed if NATS connection is open and the stream exists",
			givenStream:               true,
			givenNatsConnectionClosed: false,
			wantError:                 false,
			wantStatusCode:            http.StatusNoContent,
		},
		{
			name:                      "send in jetstream mode should fail if NATS connection is not open",
			givenNatsConnectionClosed: true,
			wantError:                 true,
			wantStatusCode:            http.StatusBadGateway,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// given
			testEnv := setupTestEnvironment(t)
			natsServer, connection := testEnv.Server, testEnv.Connection

			defer func() {
				natsServer.Shutdown()
				connection.Close()
			}()

			if tc.givenStream {
				addStream(t, connection, getStreamConfig())
			}

			ce := createCloudEvent(t)

			// when
			ctx := context.Background()
			sender := NewJetstreamMessageSender(context.Background(), connection, testEnv.Config, logrus.New())

			if tc.givenNatsConnectionClosed {
				connection.Close()
			}

			// then
			status, err := sender.Send(ctx, ce)
			assert.Equal(t, tc.wantError, err != nil)
			assert.Equal(t, tc.wantStatusCode, status)
		})
	}
}

func TestStreamExists(t *testing.T) {
	testCases := []struct {
		name                      string
		givenStream               bool
		givenNatsConnectionClosed bool
		wantResult                bool
		wantError                 error
	}{
		{
			name:                      "Stream doesn't exist and should return false",
			givenStream:               false,
			givenNatsConnectionClosed: false,
			wantResult:                false,
			wantError:                 nats.ErrStreamNotFound,
		},
		{
			name:                      "Stream exists and should return true",
			givenStream:               true,
			givenNatsConnectionClosed: false,
			wantResult:                true,
			wantError:                 nil,
		},
		{
			name:                      "Connection closed and error should happen",
			givenStream:               true,
			givenNatsConnectionClosed: true,
			wantResult:                false,
			wantError:                 nats.ErrConnectionClosed,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// given
			testEnv := setupTestEnvironment(t)
			natsServer, connection, sender := testEnv.Server, testEnv.Connection, testEnv.Sender

			defer func() {
				connection.Close()
				natsServer.Shutdown()
			}()

			if tc.givenStream {
				addStream(t, connection, getStreamConfig())
			}

			// close the connection to provoke the error
			if tc.givenNatsConnectionClosed {
				connection.Close()
			}

			// when
			result, err := sender.streamExists(connection)

			// then
			assert.Equal(t, result, tc.wantResult)
			assert.Equal(t, err, tc.wantError)
		})
	}
}

func TestJSSubjectPrefix(t *testing.T) {

	testCases := []struct {
		name              string
		givenPrefix       string
		givenEventSubject string
		wantSubject       string
	}{
		{
			name:              "With empty prefix",
			givenEventSubject: "custom.test",
			givenPrefix:       "",
			wantSubject:       "custom.test",
		},
		{
			name:              "With non-empty prefix",
			givenEventSubject: "custom.test",
			givenPrefix:       "prefix",
			wantSubject:       "prefix.custom.test",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// given
			config := &env.NatsConfig{
				JSStreamSubjectPrefix: tc.givenPrefix,
			}
			s := JetstreamMessageSender{envCfg: config}
			ce := createCloudEvent(t)
			ce.SetType(tc.givenEventSubject)

			// when
			subject := s.getJsSubjectToPublish(ce.Type())

			// then
			assert.Equal(t, subject, tc.wantSubject)
		})
	}
}

// helper functions and structs

type TestEnvironment struct {
	Connection *nats.Conn
	Config     *env.NatsConfig
	Logger     *logrus.Logger
	Sender     *JetstreamMessageSender
	Server     *server.Server
	JsContext  *nats.JetStreamContext
}

// setupTestEnvironment sets up the resources and mocks required for testing.
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	natsServer := testingutils.StartNatsServer(true)
	require.NotNil(t, natsServer)

	connection, err := testingutils.ConnectToNatsServer(natsServer.ClientURL())
	require.NotNil(t, connection)
	require.NoError(t, err)

	natsConfig := CreateNatsJsConfig(natsServer.ClientURL(), testingutils.MessagingEventTypePrefix)
	logger := logrus.New()
	jsCtx, err := connection.JetStream()
	require.NoError(t, err)

	sender := &JetstreamMessageSender{
		connection: connection,
		envCfg:     natsConfig,
		logger:     logger,
	}

	return &TestEnvironment{
		Connection: connection,
		Config:     natsConfig,
		Logger:     logger,
		Sender:     sender,
		Server:     natsServer,
		JsContext:  &jsCtx,
	}
}

// createCloudEvent build a cloud event.
func createCloudEvent(t *testing.T) *event.Event {
	builder := testingutils.NewCloudEventBuilder(
		testingutils.WithCloudEventType(testingutils.CloudEventType),
	)
	payload, _ := builder.BuildStructured()
	newEvent := cloudevents.NewEvent()
	newEvent.SetType(testingutils.CloudEventType)
	err := json.Unmarshal([]byte(payload), &newEvent)
	assert.NoError(t, err)

	return &newEvent
}

// getStreamConfig inits a testing stream config.
func getStreamConfig() *nats.StreamConfig {
	return &nats.StreamConfig{
		Name:      testingutils.MessagingEventTypePrefix,
		Subjects:  []string{fmt.Sprintf("%s.>", testingutils.MessagingEventTypePrefix)},
		Storage:   nats.MemoryStorage,
		Retention: nats.InterestPolicy,
	}
}

// addStream creates a stream for the test events.
func addStream(t *testing.T, connection *nats.Conn, config *nats.StreamConfig) {
	js, err := connection.JetStream()
	assert.NoError(t, err)
	_, err = js.AddStream(config)
	assert.NoError(t, err)
}

func CreateNatsJsConfig(url string, streamName string) *env.NatsConfig {
	return &env.NatsConfig{
		JSStreamName:  streamName,
		URL:           url,
		ReconnectWait: time.Second,
	}
}
