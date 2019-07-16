package v2

import (
	"net/http"
	"testing"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"

	"github.com/stretchr/testify/assert"
)

func Test_ValidatePublish_MissingSourceId(t *testing.T) {
	event := buildTestPublishRequest()
	event.Source = ""
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.Equal(t, len(err.Details), 1)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, ErrorMessageMissingSourceID, err.Message)
	assert.Equal(t, api.ErrorTypeMissingField, err.Details[0].Type)
}

func Test_ValidatePublish_MissingEventType(t *testing.T) {
	event := buildTestPublishRequest()
	event.Type = ""
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventType, err.Details[0].Field)
}

func Test_ValidatePublish_MissingEventTypeVersion(t *testing.T) {
	event := buildTestPublishRequest()
	event.TypeVersion = ""
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTypeVersion, err.Details[0].Field)
}

func Test_ValidatePublish_MissingEventTime(t *testing.T) {
	event := buildTestPublishRequest()
	event.Time = ""
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTime, err.Details[0].Field)
}

func Test_ValidatePublish_MissingData(t *testing.T) {
	event := buildTestPublishRequest()
	event.Data = nil
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldData, err.Details[0].Field)
}

func Test_ValidatePublish_EmptyData(t *testing.T) {
	event := buildTestPublishRequest()
	event.Data = ""
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, api.FieldData, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventType(t *testing.T) {
	event := buildTestPublishRequest()
	event.Type = "invalid/event-type"
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventType, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventTypeVersion(t *testing.T) {
	event := buildTestPublishRequest()
	event.TypeVersion = "$"
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTypeVersion, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventTime(t *testing.T) {
	event := buildTestPublishRequest()
	event.Time = "invalid-time"
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventTime, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidEventID(t *testing.T) {
	event := buildTestPublishRequest()
	event.ID = "invalid-id"
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldEventID, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidSourceId_In_Payload(t *testing.T) {
	event := buildTestPublishRequest()
	event.Source = "source/Id"
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceID, err.Details[0].Field)
	assert.Equal(t, api.ErrorTypeInvalidField, err.Details[0].Type)
}

func Test_ValidatePublish_Success(t *testing.T) {
	event := buildTestPublishRequest()
	err := ValidatePublish(&event, api.GetDefaultEventOptions())
	assert.Nil(t, err)
}

func Test_ValidatePublish_InvalidSourceIdLength(t *testing.T) {
	event := buildTestPublishRequest()
	opts := api.GetDefaultEventOptions()
	opts.MaxSourceIDLength = 1
	err := ValidatePublish(&event, opts)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, api.FieldSourceID, err.Details[0].Field)
	assert.Equal(t, api.ErrorTypeInvalidFieldLength, err.Details[0].Type)
}

func Test_ValidatePublish_InvalidEventTypeLength(t *testing.T) {
	event := buildTestPublishRequest()
	opts := api.GetDefaultEventOptions()
	opts.MaxEventTypeLength = 1
	err := ValidatePublish(&event, opts)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, api.FieldEventType, err.Details[0].Field)
	assert.Equal(t, api.ErrorTypeInvalidFieldLength, err.Details[0].Type)
}

func Test_ValidatePublish_InvalidEventTypeVersionLength(t *testing.T) {
	event := buildTestPublishRequest()
	opts := api.GetDefaultEventOptions()
	opts.MaxEventTypeVersionLength = 1
	err := ValidatePublish(&event, opts)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, api.FieldEventTypeVersion, err.Details[0].Field)
	assert.Equal(t, api.ErrorTypeInvalidFieldLength, err.Details[0].Type)
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

func buildTestPublishRequest() EventRequestV2 {
	event := EventRequestV2{
		ID:          "4ea567cf-812b-49d9-a4b2-cb5ddf464094",
		Source:      "ec-default",
		SpecVersion: "0.3",
		Time:        "2012-11-01T22:08:41+00:00",
		Type:        "test-event-type",
		TypeVersion: "v1",
		Data:        "{'key':'value'}",
	}
	return event
}
