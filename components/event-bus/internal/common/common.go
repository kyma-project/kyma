package common

import (
	"fmt"
	"strings"
	"sync"
)

// EventDetails represents the event details.
type EventDetails struct {
	eventType        string
	eventTypeVersion string
	sourceID         string
}

// escapePeriodsAndBackSlashes Replaces all occurrences of the `.` with `\.` as well as `\` with `\\`
func escapePeriodsAndBackSlashes(in *string) string {
	s := strings.Replace(*in, `\`, `\\`, -1)
	return strings.Replace(s, `.`, `\.`, -1)
}

// Encode formats the event details into a NATS Streaming compliant subject name literal.
// Encoded subject is constructed by using the Period (`.`) character be added between tokens as a delimiter.
// Period character in a token literal will be escaped with a forward slash (`\.`), ex: `env.prod` will be `env\.prod`.
// Return value is a string literal composed of the event details tokens as: `sourceID + eventType + eventTypeVersion`.
func (e *EventDetails) Encode() string {
	return fmt.Sprintf(`%s.%s.%s`,
		escapePeriodsAndBackSlashes(&e.sourceID),
		escapePeriodsAndBackSlashes(&e.eventType),
		escapePeriodsAndBackSlashes(&e.eventTypeVersion))
}

// StatusReady represents the subscription ready status.
type StatusReady struct {
	mu    sync.RWMutex
	ready bool
}

// SetReady sets the subscription ready status to true.
func (s *StatusReady) SetReady() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.ready {
		s.ready = true
		return true
	}
	return false
}

// SetNotReady sets the subscription ready status to false.
func (s *StatusReady) SetNotReady() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ready {
		s.ready = false
	}
}
