package publish

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidatePublish_MissingSource(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Source = nil
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSource, err.Details[0].Field)
}

func Test_ValidatePublish_MissingSourceType(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Source.SourceType = ""
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceType, err.Details[0].Field)
}

func Test_ValidatePublish_MissingSourceNamespace(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Source.SourceNamespace = ""
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceNamespace, err.Details[0].Field)
}

func Test_ValidatePublish_MissingSourceEnvironment(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Source.SourceEnvironment = ""
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceEnvironment, err.Details[0].Field)
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

func Test_ValidatePublish_InvalidSourceEnvironment(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Source.SourceEnvironment = ".invalid."
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceEnvironment, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidSourceNamespace(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Source.SourceNamespace = ".invalid."
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceNamespace, err.Details[0].Field)
}

func Test_ValidatePublish_InvalidSourceType(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	publishRequest.Source.SourceType = ".invalid."
	err := ValidatePublish(&publishRequest)
	assert.NotEqual(t, len(err.Details), 0)
	assert.Equal(t, http.StatusBadRequest, err.Status)
	assert.Equal(t, FieldSourceType, err.Details[0].Field)
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

func Test_ValidatePublish_Success(t *testing.T) {
	publishRequest := buildTestPublishRequest()
	err := ValidatePublish(&publishRequest)
	assert.Nil(t, err)
}

func TestSourceEnvironmentRegex(t *testing.T) {
	testRegex(t, isValidSourceEnvironment, "stage", true)        // alphabet
	testRegex(t, isValidSourceEnvironment, "prod123", true)      // alphanumeric
	testRegex(t, isValidSourceEnvironment, "my.s3.bucket", true) // . allowed
	testRegex(t, isValidSourceEnvironment, "my-s3-bucket", true) // - allowed
	testRegex(t, isValidSourceEnvironment, "my_s3_bucket", true) // _ allowed
	testRegex(t, isValidSourceEnvironment, "1stage", false)      // cannot start with number
	testRegex(t, isValidSourceEnvironment, ".stage", false)      // cannot start with symbol
	testRegex(t, isValidSourceEnvironment, "stage.", false)      // cannot end with symbol
}

func TestSourceNamespaceRegex(t *testing.T) {
	testRegex(t, isValidSourceNamespace, "kafka", true)              // alphabet
	testRegex(t, isValidSourceNamespace, "kafka10", true)            // alphanumeric
	testRegex(t, isValidSourceNamespace, "kafka.apache.org", true)   // . allowed
	testRegex(t, isValidSourceNamespace, "kafka-apache-org", true)   // - allowed
	testRegex(t, isValidSourceNamespace, "kafka_apache_org", true)   // _ allowed
	testRegex(t, isValidSourceNamespace, "1kafka.apache.org", false) // cannot start with number
	testRegex(t, isValidSourceNamespace, ".kafka.apache.org", false) // cannot start with symbol
	testRegex(t, isValidSourceNamespace, "kafka.apache.org.", false) // cannot end with symbol
}

func TestSourceTypeRegex(t *testing.T) {
	testRegex(t, isValidSourceType, "commerce", true)       // alphabet
	testRegex(t, isValidSourceType, "s3", true)             // alphanumeric
	testRegex(t, isValidSourceType, "marketing.beta", true) // . allowed
	testRegex(t, isValidSourceType, "marketing-beta", true) // - allowed
	testRegex(t, isValidSourceType, "marketing_beta", true) // _ allowed
	testRegex(t, isValidSourceType, "1marketing", false)    // cannot start with number
	testRegex(t, isValidSourceType, ".marketing", false)    // cannot start with symbol
	testRegex(t, isValidSourceType, "marketing.", false)    // cannot end with symbol
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
		Source: &EventSource{
			SourceEnvironment: "test-source-environment",
			SourceNamespace:   "test-source-namespace",
			SourceType:        "test-source-type",
		},
	}
	return publishRequest
}
