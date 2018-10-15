package publish

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidatePublish_MissingSourceId(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.SourceID = ""
	err := ValidatePublish(&publishRequest)
	assert.Equal(t, len(err.Details), 1)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, ErrorMessageMissingSourceId, err.Message)
	assert.Equal(t, ErrorTypeMissingFieldOrHeader, err.Details[0].Type)
}

func Test_ValidatePublish_MissingEventType(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventType = ""
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventType, err.Details[0].Field)
}

func Test_ValidatePublish_MissingEventTypeVersion(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventTypeVersion = ""
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTypeVersion, err.Details[0].Field)
}

func Test_ValidatePublish_MissingEventTime(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventTime = ""
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTime, err.Details[0].Field)
}

func Test_ValidatePublish_MissingData(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Data = nil
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldData, err.Details[0].Field)
}

func Test_ValidatePublish_EmptyData(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Data = ""
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldData, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventType(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventType = "invalid/event-type"
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventType, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventTypeVersion(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventTypeVersion = "$"
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTypeVersion, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventTime(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventTime = "invalid-time"
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTime, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventID(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventID = "invalid-id"
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventId, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidSourceId_In_Payload(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.SourceIdFromHeader = false
	publishRequest.SourceID = "invalid/sourceId"
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceId, err.Details[0].Field)
	assert.Equal(t, ErrorTypeInvalidField, err.Details[0].Type)
}

func Test_ValidatePublish_InvalidSourceId_In_Header(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.SourceIdFromHeader = true
	publishRequest.SourceID = "invalid/sourceId"
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, HeaderSourceId, err.Details[0].Field)
	assert.Equal(t, ErrorTypeInvalidHeader, err.Details[0].Type)
}

func Test_ValidatePublish_Success(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	err := ValidatePublish(&publishRequest)
	assert.Nil(t, err)
}

func TestEventTypeRegex(t *testing.T) {
	testRegex(t, isValidEventType, "created", true)         // alphabet
	testRegex(t, isValidEventType, "created1", true)        // alphanumeric
	testRegex(t, isValidEventType, "order.created", true)   // . allowed
	testRegex(t, isValidEventType, "order-created", true)   // - allowed
	testRegex(t, isValidEventType, "order_created", true)   // _ allowed
	testRegex(t, isValidEventType, "1order.created", false) // cannot start with number
	testRegex(t, isValidEventType, ".order.created", false) // cannot start with symbol
	testRegex(t, isValidEventType, "order.created.", false) // cannot end with symbol
}

func TestEventTypeVersionRegex(t *testing.T) {
	testRegex(t, isValidEventTypeVersion, "beta", true) // alphabet
	testRegex(t, isValidEventTypeVersion, "v1", true)   // alphanumeric
	testRegex(t, isValidEventTypeVersion, "v.1", false) // . not allowed
	testRegex(t, isValidEventTypeVersion, "v-1", false) // - not allowed
	testRegex(t, isValidEventTypeVersion, "v_1", false) // _ not allowed
	testRegex(t, isValidEventTypeVersion, "1v", true)   // can start with number
	testRegex(t, isValidEventTypeVersion, ".v1", false) // cannot start with symbol
	testRegex(t, isValidEventTypeVersion, "v1.", false) // cannot end with symbol
}

func testRegex(t *testing.T, match func(s string) bool, target string, expected bool) {
	assert.Equal(t, expected, match(target))
}

func buildTestPublishRequest() PublishRequest {
	publishRequest := PublishRequest{
		Data:             "{'key':'value'}",
		EventID:          "4ea567cf-812b-49d9-a4b2-cb5ddf464094",
		EventTime:        "2012-11-01T22:08:41+00:00",
		EventType:        "test-event-type",
		EventTypeVersion: "v1",
		SourceID:         "stage.com.org.commerce",
	}
	return publishRequest
}
