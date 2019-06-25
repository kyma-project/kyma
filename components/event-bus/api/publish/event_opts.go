package publish

const (
	// DefaultMaxSourceIDLength default value
	DefaultMaxSourceIDLength = 253
	// DefaultMaxEventTypeLength default value
	DefaultMaxEventTypeLength = 253
	// DefaultMaxEventTypeVersionLength default value
	DefaultMaxEventTypeVersionLength = 4
)

// EventOptions represents the event options.
type EventOptions struct {
	MaxSourceIDLength         int
	MaxEventTypeLength        int
	MaxEventTypeVersionLength int
}

// GetDefaultEventOptions returns a new default event options instance.
func GetDefaultEventOptions() *EventOptions {
	options := EventOptions{
		DefaultMaxSourceIDLength,
		DefaultMaxEventTypeLength,
		DefaultMaxEventTypeVersionLength,
	}
	return &options
}
