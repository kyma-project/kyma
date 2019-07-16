package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/stretchr/testify/assert"
)

const (
	// TestSourceID used in the tests
	TestSourceID = "test-source-id"
	// TestEventType used in the tests
	TestEventType = "test-event-type"
	// TestEventTypeVersion used in the tests
	TestEventTypeVersion = "v1"
	// TestEventTypeVersionInvalid used in the tests
	TestEventTypeVersionInvalid = "#"
	// TestEventTime used in the tests
	TestEventTime = "2012-11-01T22:08:41+00:00"
	// TestEventTimeInvalid used in the tests
	TestEventTimeInvalid = "2012-11-01T22"
	// TestEventID used in the tests
	TestEventID = "4ea567cf-812b-49d9-a4b2-cb5ddf464094"
	// TestEventIDInvalid used in the tests
	TestEventIDInvalid = "4ea567cf"
	// TestSourceIDInvalid used in the tests
	TestSourceIDInvalid = "source/Id"
	// TestData used in the tests
	TestData = "{'key':'value'}"
	// TestDataEmpty used in the tests
	TestDataEmpty = ""

	// event payload format
	eventFormat            = "{%v}"
	sourceIDFormat         = "\"source-id\":\"%v\""
	eventTypeFormat        = "\"event-type\":\"%v\""
	eventTypeVersionFormat = "\"event-type-version\":\"%v\""
	eventIDFormat          = "\"event-id\":\"%v\""
	eventTimeFormat        = "\"event-time\":\"%v\""
	dataFormat             = "\"data\":\"%v\""
)

type eventBuilder struct {
	bytes.Buffer
}

func (b *eventBuilder) build(format string, values ...interface{}) *eventBuilder {
	if b.Len() > 0 {
		_, _ = b.WriteString(",")
	}
	source := fmt.Sprintf(format, values...)
	_, _ = b.WriteString(source)
	return b
}

func (b *eventBuilder) String() string {
	return fmt.Sprintf(eventFormat, b.Buffer.String())
}

// BuildDefaultTestPayload returns a default test payload.
func BuildDefaultTestPayload() string {
	return BuildTestPayload(TestSourceID, TestEventType, TestEventTypeVersion, TestEventID, TestEventTime, TestData)
}

// BuildTestPayload returns a test payload.
func BuildTestPayload(sourceID, eventType, eventTypeVersion, eventID, eventTime, data string) string {
	builder := new(eventBuilder).
		build(sourceIDFormat, sourceID).
		build(eventTypeFormat, eventType).
		build(eventTypeVersionFormat, eventTypeVersion).
		build(eventIDFormat, eventID).
		build(eventTimeFormat, eventTime).
		build(dataFormat, data)
	payload := builder.String()
	return payload
}

// BuildDefaultTestBadPayload returns a default test bad payload.
func BuildDefaultTestBadPayload() string {
	builder := new(eventBuilder).
		build(sourceIDFormat, TestSourceID).
		build(eventTypeFormat, TestEventType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventIDFormat, TestEventID).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	_, _ = builder.WriteString(",") // spoil the payload
	payload := builder.String()
	return payload
}

// BuildDefaultTestPayloadWithoutSourceID returns a default test payload without the source-id.
func BuildDefaultTestPayloadWithoutSourceID() string {
	builder := new(eventBuilder).
		build(eventTypeFormat, TestEventType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventIDFormat, TestEventID).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

// BuildDefaultTestPayloadWithoutEventType returns a default test payload without the event-type.
func BuildDefaultTestPayloadWithoutEventType() string {
	builder := new(eventBuilder).
		build(sourceIDFormat, TestSourceID).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventIDFormat, TestEventID).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

// BuildDefaultTestPayloadWithoutEventTypeVersion returns a default test payload without the event-type-version.
func BuildDefaultTestPayloadWithoutEventTypeVersion() string {
	builder := new(eventBuilder).
		build(sourceIDFormat, TestSourceID).
		build(eventTypeFormat, TestEventType).
		build(eventIDFormat, TestEventID).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

// BuildDefaultTestPayloadWithoutEventTime returns a default test payload without the event-time
func BuildDefaultTestPayloadWithoutEventTime() string {
	builder := new(eventBuilder).
		build(sourceIDFormat, TestSourceID).
		build(eventTypeFormat, TestEventType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventIDFormat, TestEventID).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

// BuildDefaultTestPayloadWithoutData returns a default test payload without the data
func BuildDefaultTestPayloadWithoutData() string {
	builder := new(eventBuilder).
		build(sourceIDFormat, TestSourceID).
		build(eventTypeFormat, TestEventType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventIDFormat, TestEventID).
		build(eventTimeFormat, TestEventTime)
	payload := builder.String()
	return payload
}

// BuildDefaultTestPayloadWithEmptyData returns a default test payload with empty data
func BuildDefaultTestPayloadWithEmptyData() string {
	builder := new(eventBuilder).
		build(sourceIDFormat, TestSourceID).
		build(eventTypeFormat, TestEventType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventIDFormat, TestEventID).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestDataEmpty)
	payload := builder.String()
	return payload
}

// PerformPublishV1Request performs a test publish request.
func PerformPublishV1Request(t *testing.T, publishURL string, payload string) ([]byte, int) {
	res, err := http.Post(publishURL+"/v1/events", "application/json", strings.NewReader(payload))

	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Fatal(err)
	}

	return body, res.StatusCode
}

// PerformPublishV1RequestWithHeaders performs a test publish request with HTTP headers.
func PerformPublishV1RequestWithHeaders(t *testing.T, publishURL string, payload string, headers map[string]string) ([]byte, int) {
	req, _ := http.NewRequest("POST", publishURL+"/v1/events", strings.NewReader(payload))

	req.Header.Set("Content-Type", "application/json")

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Fatal(err)
	}

	return body, res.StatusCode
}

// AssertExpectedError asserts an expected status-code and error.
func AssertExpectedError(t *testing.T, body []byte, actualStatusCode int, expectedStatusCode int, errorField interface{}, errorType interface{}) {
	var responseError api.Error
	err := json.Unmarshal(body, &responseError)
	assert.Equal(t, expectedStatusCode, actualStatusCode)
	assert.Nil(t, err)
	if errorType != nil {
		assert.Equal(t, errorType, responseError.Type)
	}
	if errorField != nil {
		assert.NotNil(t, responseError.Details)
		assert.NotEqual(t, len(responseError.Details), 0)
		assert.Equal(t, errorField, responseError.Details[0].Field)
	}
}

// PerformPublishV2RequestWithHeaders performs a test publish request with HTTP headers.
func PerformPublishV2RequestWithHeaders(t *testing.T, publishURL string, payload string, headers map[string]string) ([]byte, int) {
	req, _ := http.NewRequest("POST", publishURL+"/v2/events", strings.NewReader(payload))

	req.Header.Set("Content-Type", "application/json")

	for header, value := range headers {
		req.Header.Set(header, value)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Fatal(err)
	}

	return body, res.StatusCode
}
