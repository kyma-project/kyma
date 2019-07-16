package api

import "github.com/kyma-project/kyma/components/event-service/internal/events/api"

// PublishRequestV1 implements the service definition of PublishRequestV1
type PublishRequestV1 struct {
	EventType        string       `json:"event-type,omitempty"`
	EventTypeVersion string       `json:"event-type-version,omitempty"`
	EventID          string       `json:"event-id,omitempty"`
	EventTime        string       `json:"event-time,omitempty"`
	Data             api.AnyValue `json:"data,omitempty"`
}

// PublishEventParametersV1 holds parameters to PublishEvent
type PublishEventParametersV1 struct {
	PublishrequestV1 PublishRequestV1 `json:"publishrequest,omitempty"`
}

// SendEventParametersV1 implements the request to the outbound messaging API
type SendEventParametersV1 struct {
	SourceID         string       `json:"source-id,omitempty"`
	EventType        string       `json:"event-type,omitempty"`
	EventTypeVersion string       `json:"event-type-version,omitempty"`
	EventID          string       `json:"event-id,omitempty"`
	EventTime        string       `json:"event-time,omitempty"`
	Data             api.AnyValue `json:"data,omitempty"`
}
