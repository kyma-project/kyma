package v2

import "github.com/nats-io/nats.go"

type ConnClosedHandler func(conn *nats.Conn)
