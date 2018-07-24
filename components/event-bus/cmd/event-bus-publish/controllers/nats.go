package controllers

import (
	"log"
	"time"

	ns "github.com/kyma-project/kyma/components/event-bus/internal/stanutil"
	stan "github.com/nats-io/go-nats-streaming"
)

const (
	numPollers     = 2                // number of Poller goroutines to launch
	statusInterval = 10 * time.Second // how often to log status to stdout
)

//TODO Add reconnect logic

// Publisher sends a message to a specific topic
type Publisher interface {
	Start() error
	Stop() error
	Publish(subject, message string) error
	IsReady() (bool, error)
}

// streamingController is a NATS Streaming Controller that handles all communication aspects with NATS.
// It is also responsible of the NATS connection management.
type streamingController struct {
	clientID  string
	natsURL   string
	clusterID string
	natsConn  *stan.Conn
}

// Start Start the NATS Streaming controller
func (sc *streamingController) Start() error {
	var err error
	if sc.natsConn, err = ns.Connect(sc.clusterID, sc.clientID, sc.natsURL); err != nil {
		log.Printf(" Create new connection failed: %+v", err)
		return err
	}
	return nil
}

// Stop Stop the Streaming Controller and close the NATS connection
func (sc *streamingController) Stop() error {
	return ns.Close(sc.natsConn)
}

// Publish sends a message to NATS
func (sc *streamingController) Publish(subject, message string) error {
	messageArray := []byte(message)
	if err := ns.Publish(sc.natsConn, subject, &messageArray); err != nil {
		log.Printf(" Failed to publish message to NATS Streaming: %+v", err)
		return err
	}
	return nil
}

func (sc *streamingController) IsReady() (bool, error) {
	return ns.IsConnected(sc.natsConn), nil
}

//GetPublisher Factory function to created and return a new Publisher instance
func GetPublisher(clientID, natsURL, clusterID string) Publisher {
	publisher := streamingController{}
	publisher.clientID = clientID
	publisher.natsURL = natsURL
	publisher.clusterID = clusterID
	return &publisher
}
