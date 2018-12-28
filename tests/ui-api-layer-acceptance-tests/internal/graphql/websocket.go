package graphql

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const (
	connectionInitMsg = "connection_init"
	startMsg          = "start"
	connectionAckMsg  = "connection_ack"
	dataMsg           = "data"
)

type clusterWebsocket struct {
	websocket.Conn
}

type operationMessage struct {
	Payload json.RawMessage `json:"payload,omitempty"`
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
}

func newWebsocket(endpoint, token string, headers http.Header) (*clusterWebsocket, error) {
	// pass user token the same way like on UI
	headers.Set("Sec-WebSocket-Protocol", fmt.Sprintf("graphql-ws, %s", token))

	dialer := websocket.Dialer{
		HandshakeTimeout: websocket.DefaultDialer.HandshakeTimeout,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
	}

	connection, _, err := dialer.Dial(endpoint, headers)
	if err != nil {
		return nil, err
	}
	return &clusterWebsocket{
		Conn: *connection,
	}, nil
}

func (c *clusterWebsocket) Handshake() error {
	if err := c.WriteJSON(operationMessage{Type: connectionInitMsg}); err != nil {
		return errors.Wrap(err, "while sending init message")
	}

	var ack operationMessage
	if err := c.ReadJSON(&ack); err != nil {
		return errors.Wrap(err, "while reading ack message")
	}
	if ack.Type != connectionAckMsg {
		log.Fatal(fmt.Errorf("expected ack message, got %#v", ack))
	}

	return nil
}

func (c *clusterWebsocket) Start(query json.RawMessage) error {
	if err := c.WriteJSON(operationMessage{Type: startMsg, ID: "1", Payload: query}); err != nil {
		return errors.Wrap(err, "while starting subscription")
	}

	return nil
}

type Subscription struct {
	Close        func() error
	IsCloseError func(err error, codes ...int) bool
	Next         func(response interface{}, timeout time.Duration) error
}

func errorSubscription(err error) *Subscription {
	return &Subscription{
		Close:        func() error { return nil },
		IsCloseError: func(error, ...int) bool { return false },
		Next:         func(interface{}, time.Duration) error { return err },
	}
}

func newSubscription(connection *clusterWebsocket) *Subscription {
	return &Subscription{
		Close:        connection.Close,
		IsCloseError: websocket.IsCloseError,
		Next: func(response interface{}, timeout time.Duration) error {
			var op operationMessage
			connection.SetReadDeadline(time.Now().Add(timeout))
			err := connection.ReadJSON(&op)
			if err != nil {
				return errors.Wrap(err, "while reading JSON from subscription")
			}
			if op.Type != dataMsg {
				return fmt.Errorf("expected data message, got %s with payload %s", op.Type, op.Payload)
			}

			respDataRaw := map[string]interface{}{}
			err = json.Unmarshal(op.Payload, &respDataRaw)
			if err != nil {
				return errors.Wrap(err, "while decoding response")
			}

			if respDataRaw["errors"] != nil {
				return fmt.Errorf("%s", respDataRaw["errors"])
			}

			return unpack(respDataRaw["data"], response)
		},
	}
}

func unpack(data interface{}, into interface{}) error {
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:      into,
		TagName:     "json",
		ErrorUnused: false,
		ZeroFields:  true,
	})
	if err != nil {
		return errors.Wrap(err, "while initializing mapstructure")
	}

	return d.Decode(data)
}
