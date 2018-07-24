package stanutil

import (
	"errors"
	"fmt"
	"log"

	stan "github.com/nats-io/go-nats-streaming"
)

// Connect creates a new NATS-Streaming connection
func Connect(clusterID string, clientID string, natsURL string) (*stan.Conn, error) {
	sc, err := stan.Connect(clusterID, clientID, stan.NatsURL(natsURL))
	if err != nil {
		log.Printf("Can't connect to: %s ; error: %v; NATS URL: %s", clusterID, err, natsURL)
	}
	return &sc, err
}

// Close must be the last call to close the connection
func Close(sc *stan.Conn) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from: %v", r)
			log.Printf("Close(): %v\n", err.Error())
		}
	}()

	if sc == nil {
		err = errors.New("can't close empty stan connection")
		return
	}
	err = (*sc).Close()
	if err != nil {
		log.Printf("Can't close connection: %+v\n", err)
	}
	return
}

// Publish a message to a subject
func Publish(sc *stan.Conn, subj string, msg *[]byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from: %v", r)
			log.Printf("Publish(): %v\n", err.Error())
		}
	}()

	if sc == nil {
		err = errors.New("cant'publish on empty stan connection")
		return
	}
	err = (*sc).Publish(subj, *msg)
	if err != nil {
		log.Printf("Error during publish: %v\n", err)
	}
	return
}

// IsConnected ....
func IsConnected(sc *stan.Conn) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
			log.Printf("IsConnected() recovered: %v\n", r)
		}
	}()

	if sc != nil && sc != (*stan.Conn)(nil) && (*sc).NatsConn() != nil {
		ok = (*sc).NatsConn().IsConnected()
	} else {
		ok = false
	}
	return
}
