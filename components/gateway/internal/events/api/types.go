// GENERATED FILE: DO NOT EDIT!

package api

// PublishRequest implements the service definition of PublishRequest
type PublishRequest struct {
	EventType        string   `json:"event-type,omitempty"`
	EventTypeVersion string   `json:"event-type-version,omitempty"`
	EventId          string   `json:"event-id,omitempty"`
	EventTime        string   `json:"event-time,omitempty"`
	Data             AnyValue `json:"data,omitempty"`
}

// PublishResponse implements the service definition of PublishResponse
type PublishResponse struct {
	EventId string `json:"event-id,omitempty"`
}

// AnyValue implements the service definition of AnyValue
type AnyValue interface {
}

// Error implements the service definition of APIError
type Error struct {
	Status   int           `json:"status,omitempty"`
	Type     string        `json:"type,omitempty"`
	Message  string        `json:"message,omitempty"`
	MoreInfo string        `json:"moreInfo,omitempty"`
	Details  []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail implements the service definition of APIErrorDetail
type ErrorDetail struct {
	Field    string `json:"field,omitempty"`
	Type     string `json:"type,omitempty"`
	Message  string `json:"message,omitempty"`
	MoreInfo string `json:"moreInfo,omitempty"`
}

// PublishEventParameters holds parameters to PublishEvent
type PublishEventParameters struct {
	Publishrequest PublishRequest `json:"publishrequest,omitempty"`
}

// PublishEventResponses holds responses of PublishEvent
type PublishEventResponses struct {
	Ok    *PublishResponse
	Error *Error
}

// EventSource implements the Source definition of the outbound messaging API
type EventSource struct {
	SourceNamespace   string `json:"source-namespace,omitempty"`
	SourceType        string `json:"source-type,omitempty"`
	SourceEnvironment string `json:"source-environment,omitempty"`
}

// SendEventParameters implements the request to the outbound messaging API
type SendEventParameters struct {
	Eventsource      EventSource `json:"source,omitempty"`
	EventType        string      `json:"event-type,omitempty"`
	EventTypeVersion string      `json:"event-type-version,omitempty"`
	EventId          string      `json:"event-id,omitempty"`
	EventTime        string      `json:"event-time,omitempty"`
	Data             AnyValue    `json:"data,omitempty"`
}

// SendEventResponse holds the response from outbound messaging API
type SendEventResponse PublishEventResponses
