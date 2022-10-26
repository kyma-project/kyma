package api

// AnyValue implements the service definition of AnyValue
type AnyValue interface {
}

// PublishRequestV1 implements the service definition of PublishRequestV1
type PublishRequestV1 struct {
	EventType        string   `json:"event-type,omitempty"`
	EventTypeVersion string   `json:"event-type-version,omitempty"`
	EventID          string   `json:"event-id,omitempty"`
	EventTime        string   `json:"event-time,omitempty"`
	Data             AnyValue `json:"data,omitempty"`
}

// PublishEventParametersV1 holds parameters to PublishEvent
type PublishEventParametersV1 struct {
	PublishrequestV1 PublishRequestV1 `json:"publishrequest,omitempty"`
}

// PublishResponse implements the service definition of PublishResponse
type PublishResponse struct {
	EventID string `json:"event-id,omitempty"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
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

// PublishEventResponses holds responses of PublishEvent
type PublishEventResponses struct {
	Ok    *PublishResponse
	Error *Error
}
