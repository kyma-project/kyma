package nats

import (
	"github.com/nats-io/nats.go"
)

type Messager interface {
	Ack(opts ...AckOpt) error
	Msg() *nats.Msg
}

var _ Messager = Msg{}

type MsgHandler func(msg Messager)

// Msg is a wrapper over nats.Msg.
type Msg struct {
	msg *nats.Msg
}

type AckOpt = nats.AckOpt

func NewMsg(msg *nats.Msg) Messager {
	return &Msg{
		msg: msg,
	}
}
func (m Msg) Ack(opts ...AckOpt) error {
	return m.msg.Ack(opts...)
}

func (m Msg) Msg() *nats.Msg {
	return m.msg
}
