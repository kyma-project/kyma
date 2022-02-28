package handlers

import (
	"errors"
	"testing"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	evtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/nats-io/nats.go"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	defaultStreamName = "kyma-eventing"
)

type jetStreamClient struct {
	nats.JetStreamContext

	natsConn *nats.Conn
}

func TestJetStream_Initialize_NoStreamExists(t *testing.T) {
	g := NewWithT(t)
	// Given
	natsServer, _ := startNATSServer(evtesting.WithJetStreamEnabled())
	defer natsServer.Shutdown()

	natsConfig := defaultNatsConfig(natsServer.ClientURL())
	defaultLogger := getLogger(g, kymalogger.INFO)

	jsClient := getJetStreamClient(natsConfig.URL)
	defer jsClient.natsConn.Close()

	jsBackend := NewJetStream(natsConfig, nil, defaultLogger)
	// No stream exists
	_, err := jsClient.StreamInfo(natsConfig.JSStreamName)
	g.Expect(errors.Is(err, nats.ErrStreamNotFound)).To(BeTrue())

	// When
	initErr := jsBackend.Initialize(env.Config{})

	// Then
	// A stream is created
	g.Expect(initErr).To(BeNil())
	g.Expect(jsClient.StreamInfo(natsConfig.JSStreamName)).ShouldNot(BeNil())
}

func TestJetStream_Initialize_StreamExists(t *testing.T) {
	g := NewWithT(t)
	// Given
	natsServer, _ := startNATSServer(evtesting.WithJetStreamEnabled())
	defer natsServer.Shutdown()

	natsConfig := defaultNatsConfig(natsServer.ClientURL())
	defaultLogger := getLogger(g, kymalogger.INFO)

	jsClient := getJetStreamClient(natsConfig.URL)
	defer jsClient.natsConn.Close()

	jsBackend := NewJetStream(natsConfig, nil, defaultLogger)
	// A stream already exists
	createdStreamInfo, err := jsClient.AddStream(&nats.StreamConfig{
		Name:    natsConfig.JSStreamName,
		Storage: nats.MemoryStorage,
	})
	g.Expect(createdStreamInfo).ToNot(BeNil())
	g.Expect(err).To(BeNil())

	// When
	initErr := jsBackend.Initialize(env.Config{})

	// Then
	// No new stream should be created
	g.Expect(initErr).To(BeNil())
	reusedStreamInfo, err := jsClient.StreamInfo(natsConfig.JSStreamName)
	g.Expect(err).To(BeNil())
	g.Expect(reusedStreamInfo.Created).To(Equal(createdStreamInfo.Created))
}

func defaultNatsConfig(url string) env.NatsConfig {
	return env.NatsConfig{
		URL:                     url,
		JSStreamName:            defaultStreamName,
		JSStreamStorageType:     JetStreamStorageTypeMemory,
		JSStreamRetentionPolicy: JetStreamRetentionPolicyInterest,
	}
}

// getJetStreamClient creates a client with JetStream context, or fails the caller test.
func getJetStreamClient(serverURL string) *jetStreamClient {
	conn, err := nats.Connect(serverURL)
	if err != nil {
		ginkgo.Fail(err.Error())
	}
	jsCtx, err := conn.JetStream()
	if err != nil {
		conn.Close()
		ginkgo.Fail(err.Error())
	}
	return &jetStreamClient{
		JetStreamContext: jsCtx,
		natsConn:         conn,
	}
}
