package jetstream

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

func TestJetStreamMessageSender(t *testing.T) {
	testCases := []struct {
		name                      string
		givenStream               bool
		givenStreamMaxBytes       int64
		givenNATSConnectionClosed bool
		wantErr                   error
		wantStatusCode            int
	}{
		{
			name:                      "send in jetstream mode should not succeed if stream doesn't exist",
			givenStream:               false,
			givenNATSConnectionClosed: false,
			wantErr:                   ErrCannotSendToStream,
		},
		{
			name:                      "send in jetstream mode should not succeed if stream is full",
			givenStream:               true,
			givenStreamMaxBytes:       1,
			givenNATSConnectionClosed: false,
			wantErr:                   ErrNoSpaceLeftOnDevice,
		},
		{
			name:                      "send in jetstream mode should succeed if NATS connection is open and the stream exists",
			givenStream:               true,
			givenStreamMaxBytes:       5000,
			givenNATSConnectionClosed: false,
			wantErr:                   nil,
		},
		{
			name:                      "send in jetstream mode should fail if NATS connection is not open",
			givenNATSConnectionClosed: true,
			wantErr:                   ErrNotConnected,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// arrange
			testEnv := setupTestEnvironment(t)
			natsServer, connection, mockedLogger := testEnv.Server, testEnv.Connection, testEnv.Logger

			defer func() {
				natsServer.Shutdown()
				connection.Close()
			}()

			if tc.givenStream {
				sc := getStreamConfig(tc.givenStreamMaxBytes)
				cc := getConsumerConfig()
				addStream(t, connection, sc)
				addConsumer(t, connection, sc, cc)
			}

			ce := createCloudEvent(t)

			ctx := context.Background()
			sender := NewSender(context.Background(), connection, testEnv.Config, &options.Options{}, mockedLogger)

			if tc.givenNATSConnectionClosed {
				connection.Close()
			}

			// act
			err := sender.Send(ctx, ce)

			testEnv.Logger.WithContext().Errorf("err: %v", err)

			// assert
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// helper functions and structs

type TestEnvironment struct {
	Connection *nats.Conn
	Config     *env.NATSConfig
	Logger     *logger.Logger
	Sender     *Sender
	Server     *server.Server
	JsContext  *nats.JetStreamContext
}

// setupTestEnvironment sets up the resources and mocks required for testing.
func setupTestEnvironment(t *testing.T) *TestEnvironment {
	natsServer := testingutils.StartNATSServer()
	require.NotNil(t, natsServer)

	connection, err := testingutils.ConnectToNATSServer(natsServer.ClientURL())
	require.NotNil(t, connection)
	require.NoError(t, err)

	natsConfig := CreateNATSJsConfig(natsServer.ClientURL())

	mockedLogger, err := logger.New("json", "info")
	require.NoError(t, err)

	jsCtx, err := connection.JetStream()
	require.NoError(t, err)

	sender := &Sender{
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
	jsType := fmt.Sprintf("%s.%s", testingutils.StreamName, testingutils.CloudEventTypeWithPrefix)
	builder := testingutils.NewCloudEventBuilder(
		testingutils.WithCloudEventType(jsType),
	)
	payload, _ := builder.BuildStructured()
	newEvent := cloudevents.NewEvent()
	newEvent.SetType(jsType)
	err := json.Unmarshal([]byte(payload), &newEvent)
	assert.NoError(t, err)

	return &newEvent
}

// getStreamConfig inits a testing stream config.
func getStreamConfig(maxBytes int64) *nats.StreamConfig {
	return &nats.StreamConfig{
		Name:      testingutils.StreamName,
		Subjects:  []string{fmt.Sprintf("%s.>", env.JetStreamSubjectPrefix)},
		Storage:   nats.MemoryStorage,
		Retention: nats.InterestPolicy,
		Discard:   nats.DiscardNew,
		MaxBytes:  maxBytes,
	}
}

func getConsumerConfig() *nats.ConsumerConfig {
	return &nats.ConsumerConfig{
		Durable:       "test",
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: fmt.Sprintf("%v.%v", env.JetStreamSubjectPrefix, testingutils.CloudEventTypeWithPrefix),
	}
}

// addStream creates a stream for the test events.
func addStream(t *testing.T, connection *nats.Conn, config *nats.StreamConfig) {
	js, err := connection.JetStream()
	assert.NoError(t, err)
	info, err := js.AddStream(config)
	t.Logf("%+v", info)
	assert.NoError(t, err)
}

func addConsumer(t *testing.T, connection *nats.Conn, sc *nats.StreamConfig, config *nats.ConsumerConfig) {
	js, err := connection.JetStream()
	assert.NoError(t, err)
	info, err := js.AddConsumer(sc.Name, config)
	t.Logf("%+v", info)
	assert.NoError(t, err)
}

func CreateNATSJsConfig(url string) *env.NATSConfig {
	return &env.NATSConfig{
		JSStreamName:    testingutils.StreamName,
		URL:             url,
		ReconnectWait:   time.Second,
		EventTypePrefix: testingutils.OldEventTypePrefix,
	}
}

func TestSender_URL(t *testing.T) {
	type fields struct {
		envCfg *env.NATSConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "URL is correct",
			want: "FOO",
			fields: fields{
				envCfg: &env.NATSConfig{
					URL: "FOO",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sender{
				envCfg: tt.fields.envCfg,
			}
			assert.Equalf(t, tt.want, s.URL(), "URL()")
		})
	}
}

func TestSender_getJsSubjectToPublish(t *testing.T) {
	t.Parallel()

	type fields struct {
		opts *options.Options
	}
	tests := []struct {
		name    string
		fields  fields
		subject string
		want    string
	}{
		{
			name:    "Appends JS prefix for v1alpha1 subscription",
			subject: "sap.kyma.custom.noapp.order.created.v1",
			want:    "kyma.sap.kyma.custom.noapp.order.created.v1",
			fields: fields{
				opts: &options.Options{},
			},
		},
		{
			name:    "Appends JS prefix for v1alpha2 exact type matching subscription",
			subject: "sap.kyma.custom.noapp.order.created.v1",
			want:    "kyma.sap.kyma.custom.noapp.order.created.v1",
			fields: fields{
				opts: &options.Options{},
			},
		},
		{
			name:    "Does not append JS prefix for v1alpha2 standard type matching subscription",
			subject: "kyma.noapp.order.created.v1",
			want:    "kyma.noapp.order.created.v1",
			fields: fields{
				opts: &options.Options{},
			},
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &Sender{
				opts:   tc.fields.opts,
				envCfg: CreateNATSJsConfig(""),
			}
			assert.Equal(t, tc.want, s.getJsSubjectToPublish(tc.subject))
		})
	}
}
