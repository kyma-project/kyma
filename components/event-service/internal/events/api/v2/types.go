package api

import "github.com/kyma-project/kyma/components/event-service/internal/events/api"

// EventRequestV2 implements the service definition of EventRequestV2
type EventRequestV2 struct {
	EventType           string       `json:"type"`
	EventTypeVersion    string       `json:"eventtypeversion"`
	EventID             string       `json:"id"`
	EventTime           string       `json:"time"`
	SpecVersion         string       `json:"specversion"`
	DataContentEncoding string       `json:"datacontentencoding,omitempty"`
	Data                api.AnyValue `json:"data"`
}

// PublishEventParametersV2 holds parameters to PublishEvent
type PublishEventParametersV2 struct {
	EventRequestV2 EventRequestV2 `json:"publishrequest,omitempty"`
}

// SendEventParametersV2 implements the request to the outbound messaging API
type SendEventParametersV2 struct {
	Source              string       `json:"source"`
	Type                string       `json:"type"`
	EventTypeVersion    string       `json:"eventtypeversion"`
	ID                  string       `json:"id"`
	Time                string       `json:"time"`
	SpecVersion         string       `json:"specversion"`
	DataContentEncoding string       `json:"datacontentencoding"`
	Data                api.AnyValue `json:"data"`
}
