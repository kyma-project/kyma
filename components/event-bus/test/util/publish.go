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
	// publish request
	TestSourceID                = "test-source-id"
	TestSourceIDInHeader        = "test-source-id-in-header"
	TestEventType               = "test-event-type"
	TestEventTypeVersion        = "v1"
	TestEventTypeVersionInvalid = "#"
	TestEventTime               = "2012-11-01T22:08:41+00:00"
	TestEventTimeInvalid        = "2012-11-01T22"
	TestEventID                 = "4ea567cf-812b-49d9-a4b2-cb5ddf464094"
	TestEventIDInvalid          = "4ea567cf"
	TestSourceIdInvalid         = "source/Id"
	TestData                    = "{'key':'value'}"
	TestDataEmpty               = ""

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

func buildTestPublishRequest(sourceID, eventType, eventTypeVersion, eventID, eventTime, data string) api.PublishRequest {
	publishRequest := api.PublishRequest{
		Data:             data,
		EventID:          eventID,
		EventTime:        eventTime,
		EventType:        eventType,
		EventTypeVersion: eventTypeVersion,
		SourceID:         sourceID,
	}
	return publishRequest
}

func BuildDefaultTestSubjectAndPayload() (string, string) {
	subject := BuildDefaultTestSubject()
	payload := BuildDefaultTestPayload()
	return subject, payload
}

func BuildDefaultTestSubject() string {
	return buildTestSubject(TestSourceID, TestEventType, TestEventTypeVersion)
}

func buildTestSubject(sourceID, eventType, eventTypeVersion string) string {
	return encodeSubject(buildTestPublishRequest(sourceID, eventType, eventTypeVersion, TestEventID, TestEventTime, TestData))
}

func BuildDefaultTestPayload() string {
	return BuildTestPayload(TestSourceID, TestEventType, TestEventTypeVersion, TestEventID, TestEventTime, TestData)
}

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

func BuildDefaultTestPayloadWithoutSourceId() string {
	builder := new(eventBuilder).
		build(eventTypeFormat, TestEventType).
		build(eventTypeVersionFormat, TestEventTypeVersion).
		build(eventIDFormat, TestEventID).
		build(eventTimeFormat, TestEventTime).
		build(dataFormat, TestData)
	payload := builder.String()
	return payload
}

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

func encodeSubject(r api.PublishRequest) string {
	return fmt.Sprintf("%s.%s.%s", r.SourceID, r.EventType, r.EventTypeVersion)
}

func PerformPublishRequest(t *testing.T, publishURL string, payload string) ([]byte, int) {
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

func PerformPublishRequestWithHeaders(t *testing.T, publishURL string, payload string, headers map[string]string) ([]byte, int) {
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

func VerifyReceivedMsg(t *testing.T, a string, b []byte) {
	var bReq api.PublishRequest
	err := json.Unmarshal(b, &bReq)
	if err != nil {
		t.Error(err)
	}
	var aReq api.PublishRequest
	err = json.Unmarshal([]byte(a), &aReq)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, aReq.EventID, bReq.EventID)
	assert.Equal(t, aReq.SourceID, bReq.SourceID)
}

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
