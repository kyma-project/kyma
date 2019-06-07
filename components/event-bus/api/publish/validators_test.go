package publish

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidatePublish_MissingSourceId(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.SourceID = ""
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.Equal(t, len(err.Details), 1)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, ErrorMessageMissingSourceID, err.Message)
	assert.Equal(t, ErrorTypeMissingFieldOrHeader, err.Details[0].Type)
}

func Test_ValidatePublish_MissingEventType(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventType = ""
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventType, err.Details[0].Field)
}

func Test_ValidatePublish_MissingEventTypeVersion(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventTypeVersion = ""
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTypeVersion, err.Details[0].Field)
}

func Test_ValidatePublish_MissingEventTime(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventTime = ""
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTime, err.Details[0].Field)
}

func Test_ValidatePublish_MissingData(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Data = nil
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldData, err.Details[0].Field)
}

func Test_ValidatePublish_EmptyData(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Data = ""
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldData, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventType(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventType = "invalid/event-type"
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventType, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventTypeVersion(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventTypeVersion = "$"
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTypeVersion, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventTime(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventTime = "invalid-time"
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTime, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventID(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.EventID = "invalid-id"
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventID, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidSourceId_In_Payload(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.SourceIDFromHeader = false
	publishRequest.SourceID = "source/Id"
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceID, err.Details[0].Field)
	assert.Equal(t, ErrorTypeInvalidField, err.Details[0].Type)
}

func Test_ValidatePublish_InvalidSourceId_In_Header(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.SourceIDFromHeader = true
	publishRequest.SourceID = "source/Id"
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, HeaderSourceID, err.Details[0].Field)
	assert.Equal(t, ErrorTypeInvalidHeader, err.Details[0].Type)
}

func Test_ValidatePublish_Success(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	err := ValidatePublish(&publishRequest, GetDefaultEventOptions())
	assert.Nil(t, err)
}

func Test_ValidatePublish_InvalidSourceIdLength(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	opts := GetDefaultEventOptions()
	opts.MaxSourceIDLength = 1
	err := ValidatePublish(&publishRequest, opts)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceID, err.Details[0].Field)
	assert.Equal(t, ErrorTypeInvalidFieldLength, err.Details[0].Type)
}

func Test_ValidatePublish_InvalidEventTypeLength(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	opts := GetDefaultEventOptions()
	opts.MaxEventTypeLength = 1
	err := ValidatePublish(&publishRequest, opts)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventType, err.Details[0].Field)
	assert.Equal(t, ErrorTypeInvalidFieldLength, err.Details[0].Type)
}

func Test_ValidatePublish_InvalidEventTypeVersionLength(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	opts := GetDefaultEventOptions()
	opts.MaxEventTypeVersionLength = 1
	err := ValidatePublish(&publishRequest, opts)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTypeVersion, err.Details[0].Field)
	assert.Equal(t, ErrorTypeInvalidFieldLength, err.Details[0].Type)
}

func TestSourceIDAndEventTypeRegex(t *testing.T) {
	// prepare test cases
	testCases := map[string]bool{
		"test.test": true,  // . allowed
		"test-test": true,  // - allowed
		"test":      true,  // alphabet
		"test1":     true,  // can end with number
		"1test":     true,  // can start with number
		"TEST":      true,  // uppercase allowed
		"test*test": false, // * not allowed
		"test_test": false, // _ not allowed
		"test.":     false, // cannot end with symbol
		".test":     false, // cannot start with symbol
	}

	// run test cases
	for testCase, expected := range testCases {
		testRegex(t, isValidSourceID, testCase, expected)
		testRegex(t, isValidEventType, testCase, expected)
	}
}

func TestEventTypeVersionRegex(t *testing.T) {
	// prepare test cases
	testCases := map[string]bool{
		"test":      true,  // alphabet
		"test1":     true,  // can end with number
		"1test":     true,  // can start with number
		"TEST":      true,  // uppercase allowed
		"test.test": false, // . allowed
		"test-test": false, // - allowed
		"test*test": false, // * not allowed
		"test_test": false, // _ not allowed
		"test.":     false, // cannot end with symbol
		".test":     false, // cannot start with symbol
	}

	// run test cases
	for testCase, expected := range testCases {
		testRegex(t, isValidEventTypeVersion, testCase, expected)
	}
}

func testRegex(t *testing.T, match func(s string) bool, target string, expected bool) {
	assert.Equal(t, expected, match(target))
}

func buildTestPublishRequest() Request {
	publishRequest := Request{
		Data:             "{'key':'value'}",
		EventID:          "4ea567cf-812b-49d9-a4b2-cb5ddf464094",
		EventTime:        "2012-11-01T22:08:41+00:00",
		EventType:        "test-event-type",
		EventTypeVersion: "v1",
		SourceID:         "ec-default",
	}
	return publishRequest
}
