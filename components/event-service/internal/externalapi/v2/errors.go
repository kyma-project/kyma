package v2

import (
	"github.com/kyma-project/kyma/components/event-service/internal/events/api"
	"github.com/kyma-project/kyma/components/event-service/internal/events/shared"
)

// ErrorResponseMissingFieldEventID returns an error of type PublishEventResponse for missing EventID field
func ErrorResponseMissingFieldEventID() (response *api.PublishEventResponse) {
	return shared.CreateMissingFieldError(shared.FieldEventIDV2)
}

// ErrorResponseMissingFieldEventType returns an error of type PublishEventResponse for missing EventType field
func ErrorResponseMissingFieldEventType() (response *api.PublishEventResponse) {
	return shared.CreateMissingFieldError(shared.FieldEventTypeV2)
}

// ErrorResponseMissingFieldEventTypeVersion returns an error of type PublishEventResponse for missing EventTypeVersion field
func ErrorResponseMissingFieldEventTypeVersion() (response *api.PublishEventResponse) {
	return shared.CreateMissingFieldError(shared.FieldEventTypeVersionV2)
}

// ErrorResponseWrongEventTypeVersion returns an error of type PublishEventResponse for wrong EventTypeVersion field
func ErrorResponseWrongEventTypeVersion() (response *api.PublishEventResponse) {
	return shared.CreateInvalidFieldError(shared.FieldEventTypeVersionV2)
}

// ErrorResponseMissingFieldEventTime returns an error of type PublishEventResponse for missing EventTime field
func ErrorResponseMissingFieldEventTime() (response *api.PublishEventResponse) {
	return shared.CreateMissingFieldError(shared.FieldEventTimeV2)
}

// ErrorResponseWrongEventTime returns an error of type PublishEventResponse for wrong EventTime field
func ErrorResponseWrongEventTime() (response *api.PublishEventResponse) {
	return shared.CreateInvalidFieldError(shared.FieldEventTimeV2)
}

// ErrorResponseWrongEventID returns an error of type PublishEventResponse for wrong EventID field
func ErrorResponseWrongEventID() (response *api.PublishEventResponse) {
	return shared.CreateInvalidFieldError(shared.FieldEventIDV2)
}

// ErrorResponseWrongSpecVersion returns an error of type PublishEventResponse for wrong SpecVersion field
func ErrorResponseWrongSpecVersion() (response *api.PublishEventResponse) {
	return shared.CreateInvalidFieldError(shared.FieldSpecVersionV2)
}

// ErrorResponseMissingFieldData returns an error of type PublishEventResponse for missing Data field
func ErrorResponseMissingFieldData() (response *api.PublishEventResponse) {
	return shared.CreateMissingFieldError(shared.FieldData)
}
