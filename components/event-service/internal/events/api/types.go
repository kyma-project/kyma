// GENERATED FILE: DO NOT EDIT!

package api

// PublishRequest implements the service definition of PublishRequest
type PublishRequest struct {
	EventType        string   `json:"event-type,omitempty"`
	EventTypeVersion string   `json:"event-type-version,omitempty"`
	EventID          string   `json:"event-id,omitempty"`
	EventTime        string   `json:"event-time,omitempty"`
	Data             AnyValue `json:"data,omitempty"`
}

// PublishResponse implements the service definition of PublishResponse
type PublishResponse struct {
	EventID string `json:"event-id,omitempty"`
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

// SendEventParameters implements the request to the outbound messaging API
type SendEventParameters struct {
	SourceID         string   `json:"source-id,omitempty"`
	EventType        string   `json:"event-type,omitempty"`
	EventTypeVersion string   `json:"event-type-version,omitempty"`
	EventID          string   `json:"event-id,omitempty"`
	EventTime        string   `json:"event-time,omitempty"`
	Data             AnyValue `json:"data,omitempty"`
}

// SendEventResponse holds the response from outbound messaging API
type SendEventResponse PublishEventResponses
