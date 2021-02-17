package testkit

type SendEvent struct {
	State   SendEventState
	AppName string
	Payload string
}

// SendEventState represents SendEvent dependencies
type SendEventState interface {
	GetEventSender() *EventSender
}
