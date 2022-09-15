package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
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
			natsServer, connection, mockedLogger := testEnv.Server, testEnv.Connection, testEnv.Logger

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
			sender := NewJetstreamMessageSender(context.Background(), connection, testEnv.Config, mockedLogger)

			if tc.givenNatsConnectionClosed {
				connection.Close()
			}

			// then
			status, err := sender.Send(ctx, ce)

			testEnv.Logger.WithContext().Errorf("err: %v", err)
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

// helper functions and structs

type TestEnvironment struct {
	Connection *nats.Conn
	Config     *env.NATSConfig
	Logger     *logger.Logger
	Sender     *JetstreamMessageSender
	Server     *server.Server
	JsContext  *nats.JetStreamContext
}

// setupTestEnvironment sets up the resources and mocks required for testing.
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	natsServer := testingutils.StartNATSServer(true)
	require.NotNil(t, natsServer)

	connection, err := testingutils.ConnectToNATSServer(natsServer.ClientURL())
	require.NotNil(t, connection)
	require.NoError(t, err)

	natsConfig := CreateNatsJsConfig(natsServer.ClientURL())

	mockedLogger, err := logger.New("json", "info")
	require.NoError(t, err)

	jsCtx, err := connection.JetStream()
	require.NoError(t, err)

	sender := &JetstreamMessageSender{
		connection: connection,
		envCfg:     natsConfig,
		logger:     mockedLogger,
	}

	return &TestEnvironment{
		Connection: connection,
		Config:     natsConfig,
		Logger:     mockedLogger,
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
		Name:      testingutils.StreamName,
		Subjects:  []string{fmt.Sprintf("%s.>", env.JetStreamSubjectPrefix)},
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

func CreateNatsJsConfig(url string) *env.NATSConfig {
	return &env.NATSConfig{
		JSStreamName:  testingutils.StreamName,
		URL:           url,
		ReconnectWait: time.Second,
	}
}
