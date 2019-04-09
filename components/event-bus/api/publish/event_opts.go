package publish

const (
	DefaultMaxSourceIDLength         = 253
	DefaultMaxEventTypeLength        = 253
	DefaultMaxEventTypeVersionLength = 4
)

type EventOptions struct {
	MaxSourceIDLength         int
	MaxEventTypeLength        int
	MaxEventTypeVersionLength int
}

func GetDefaultEventOptions() *EventOptions {
	options := EventOptions{
		DefaultMaxSourceIDLength,
		DefaultMaxEventTypeLength,
		DefaultMaxEventTypeVersionLength,
	}
	return &options
}
