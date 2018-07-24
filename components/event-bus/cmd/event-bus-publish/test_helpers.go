package main

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
	testUrl                     = "/v1/events/"
	testSourceType              = "test-source-type"
	testSourceNamespace         = "test-source-namespace"
	testSourceEnvironment       = "test-source-environment"
	testEventType               = "test-event-type"
	testEventTypeVersion        = "v1"
	testEventTypeVersionInvalid = "#"
	testEventTime               = "2012-11-01T22:08:41+00:00"
	testEventTimeInvalid        = "2012-11-01T22"
	testEventID                 = "4ea567cf-812b-49d9-a4b2-cb5ddf464094"
	testEventIDInvalid          = "4ea567cf"
	testData                    = "{'key':'value'}"
	testDataEmpty               = ""

	// event payload format
	eventFormat            = "{%v}"
	sourceFormat           = "\"source\":{\"source-namespace\":\"%v\",\"source-type\":\"%v\",\"source-environment\":\"%v\"}"
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
		b.WriteString(",")
	}
	source := fmt.Sprintf(format, values...)
	b.WriteString(source)
	return b
}

func (b *eventBuilder) String() string {
	return fmt.Sprintf(eventFormat, b.Buffer.String())
}

func buildTestPublishRequest(sourceNamespace, sourceType, sourceEnvironment, eventType, eventTypeVersion, eventID, eventTime, data string) api.PublishRequest {
	publishRequest := api.PublishRequest{
		Data:             data,
		EventID:          eventID,
		EventTime:        eventTime,
		EventType:        eventType,
		EventTypeVersion: eventTypeVersion,
		Source: &api.EventSource{
			SourceEnvironment: sourceEnvironment,
			SourceNamespace:   sourceNamespace,
			SourceType:        sourceType,
		},
	}
	return publishRequest
}

func buildDefaultTestPublishRequest() api.PublishRequest {
	return buildTestPublishRequest(testSourceNamespace, testSourceType, testSourceEnvironment, testEventType, testEventTypeVersion, testEventID, testEventTime, testData)
}

func buildDefaultTestSubjectAndPayload() (string, string) {
	subject := buildDefaultTestSubject()
	payload := buildDefaultTestPayload()
	return subject, payload
}

func buildDefaultTestSubject() string {
	return buildTestSubject(testSourceEnvironment, testSourceNamespace, testSourceType, testEventType, testEventTypeVersion)
}

func buildTestSubject(sourceEnvironment, sourceNamespace, sourceType, eventType, eventTypeVersion string) string {
	return encodeSubject(buildTestPublishRequest(sourceNamespace, sourceType, sourceEnvironment, eventType, eventTypeVersion, testEventID, testEventTime, testData))
}

func buildDefaultTestPayload() string {
	return buildTestPayload(testSourceNamespace, testSourceType, testSourceEnvironment, testEventType, testEventTypeVersion, testEventID, testEventTime, testData)
}

func buildTestPayload(sourceNamespace, sourceType, sourceEnvironment, eventType, eventTypeVersion, eventID, eventTime, data string) string {
	builder := new(eventBuilder).
		build(sourceFormat, sourceNamespace, sourceType, sourceEnvironment).
		build(eventTypeFormat, eventType).
		build(eventTypeVersionFormat, eventTypeVersion).
		build(eventIDFormat, eventID).
		build(eventTimeFormat, eventTime).
		build(dataFormat, data)
	payload := builder.String()
	return payload
}

func buildDefaultTestBadPayload() string {
	builder := new(eventBuilder).
		build(sourceFormat, testSourceNamespace, testSourceType, testSourceEnvironment).
		build(eventTypeFormat, testEventType).
		build(eventTypeVersionFormat, testEventTypeVersion).
		build(eventIDFormat, testEventID).
		build(eventTimeFormat, testEventTime).
		build(dataFormat, testData)
	builder.WriteString(",") // spoil the payload
	payload := builder.String()
	return payload
}

func buildDefaultTestPayloadWithoutSource() string {
	builder := new(eventBuilder).
		build(eventTypeFormat, testEventType).
		build(eventTypeVersionFormat, testEventTypeVersion).
		build(eventIDFormat, testEventID).
		build(eventTimeFormat, testEventTime).
		build(dataFormat, testData)
	payload := builder.String()
	return payload
}

func buildDefaultTestPayloadWithoutEventType() string {
	builder := new(eventBuilder).
		build(sourceFormat, testSourceNamespace, testSourceType, testSourceEnvironment).
		build(eventTypeVersionFormat, testEventTypeVersion).
		build(eventIDFormat, testEventID).
		build(eventTimeFormat, testEventTime).
		build(dataFormat, testData)
	payload := builder.String()
	return payload
}

func buildDefaultTestPayloadWithoutEventTypeVersion() string {
	builder := new(eventBuilder).
		build(sourceFormat, testSourceNamespace, testSourceType, testSourceEnvironment).
		build(eventTypeFormat, testEventType).
		build(eventIDFormat, testEventID).
		build(eventTimeFormat, testEventTime).
		build(dataFormat, testData)
	payload := builder.String()
	return payload
}

func buildDefaultTestPayloadWithoutEventTime() string {
	builder := new(eventBuilder).
		build(sourceFormat, testSourceNamespace, testSourceType, testSourceEnvironment).
		build(eventTypeFormat, testEventType).
		build(eventTypeVersionFormat, testEventTypeVersion).
		build(eventIDFormat, testEventID).
		build(dataFormat, testData)
	payload := builder.String()
	return payload
}

func buildDefaultTestPayloadWithoutData() string {
	builder := new(eventBuilder).
		build(sourceFormat, testSourceNamespace, testSourceType, testSourceEnvironment).
		build(eventTypeFormat, testEventType).
		build(eventTypeVersionFormat, testEventTypeVersion).
		build(eventIDFormat, testEventID).
		build(eventTimeFormat, testEventTime)
	payload := builder.String()
	return payload
}

func buildDefaultTestPayloadWithEmptyData() string {
	builder := new(eventBuilder).
		build(sourceFormat, testSourceNamespace, testSourceType, testSourceEnvironment).
		build(eventTypeFormat, testEventType).
		build(eventTypeVersionFormat, testEventTypeVersion).
		build(eventIDFormat, testEventID).
		build(eventTimeFormat, testEventTime).
		build(dataFormat, testDataEmpty)
	payload := builder.String()
	return payload
}

func encodeSubject(r api.PublishRequest) string {
	return fmt.Sprintf("%s.%s.%s.%s.%s",
		r.Source.SourceEnvironment,
		r.Source.SourceNamespace,
		r.Source.SourceType,
		r.EventType,
		r.EventTypeVersion)
}

func performPublishRequest(t *testing.T, publishURL string, payload string) ([]byte, int) {
	res, err := http.Post(publishURL+"/v1/events", "application/json", strings.NewReader(payload))

	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Fatal(err)
	}

	return body, res.StatusCode
}

func verifyReceivedMsg(t *testing.T, a string, b []byte) {
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
}

func assertExpectedError(t *testing.T, body []byte, actualStatusCode int, expectedStatusCode int, errorField interface{}, errorType interface{}) {
	var responseError api.Error
	err := json.Unmarshal(body, &responseError)
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
