package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/publisher"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/test/fake"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/publish/opts"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
	test "github.com/kyma-project/kyma/components/event-bus/test/util"
	"github.com/stretchr/testify/assert"
)

var (
	server      *httptest.Server
	knativeLib  *knative.KnativeLib
	application *KnativePublishApplication
)

func TestMain(m *testing.M) {
	// init and start the knative publish application
	options := opts.GetDefaultOptions()
	knativeLib = &knative.KnativeLib{}
	mockPublisher := fake.NewMockKnativePublisher()
	tracer := trace.StartNewTracer(options.TraceOptions)
	application = StartKnativePublishApplication(options, knativeLib, &mockPublisher, &tracer)

	// init the publish server
	server = httptest.NewServer(application.ServeMux())

	// start running the tests
	exitCode := m.Run()

	// cleanup
	tracer.Stop()
	server.Close()
	os.Exit(exitCode)
}

func Test_KnativePublishApplication_ShouldStart(t *testing.T) {
	assert.NotNil(t, (*application).tracer)
	assert.NotNil(t, (*application).serveMux)
	assert.NotNil(t, (*application).knativePublisher)
	assert.NotNil(t, (*application).knativeLib)
	assert.Equal(t, true, (*application).started)
}

func Test_PublishWithSourceIdInPayload_ShouldSucceed(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayload()
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusOK, statusCode)

	// get the response
	publishResponse := &api.PublishResponse{}
	err := json.Unmarshal(body, &publishResponse)

	// assert
	assert.Nil(t, err)
	assert.Equal(t, test.TestEventID, publishResponse.EventID)
}

func Test_PublishWithSourceIdInHeader_ShouldSucceed(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayloadWithoutSourceId()
	body, statusCode := test.PerformPublishRequestWithHeaders(t, server.URL, payload, map[string]string{api.HeaderSourceId: test.TestSourceID})

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusOK, statusCode)

	// get the response
	publishResponse := &api.PublishResponse{}
	err := json.Unmarshal(body, &publishResponse)

	// assert
	assert.Nil(t, err)
	assert.Equal(t, test.TestEventID, publishResponse.EventID)
}

func Test_PublishWithChannelNameGreaterThanMaxChannelLength_ShouldFail(t *testing.T) {
	// make the max channel name length to be very low
	application.options.MaxChannelNameLength = 1

	// prepare and send payload
	payload := test.BuildDefaultTestPayload()
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)

	// get the response
	err := &api.Error{}
	marshalErr := json.Unmarshal(body, &err)

	// assert
	assert.Nil(t, marshalErr)
	assert.Equal(t, api.ErrorTypeValidationViolation, err.Type)
	assert.Equal(t, api.ErrorTypeInvalidFieldLength, err.Details[0].Type)
	assert.Equal(t, knative.FieldKnativeChannelName, err.Details[0].Field)

	// restore the max channel name original length
	application.options.MaxChannelNameLength = opts.GetDefaultOptions().MaxChannelNameLength
}

func Test_PublishWithBadPayload_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestBadPayload()
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, nil, api.ErrorTypeBadPayload)
}

func Test_PublishWithoutSourceId_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayloadWithoutSourceId()
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, fmt.Sprintf("%v/%v", api.FieldSourceId, api.HeaderSourceId), api.ErrorTypeValidationViolation)
}

func Test_PublishWithoutEventType_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayloadWithoutEventType()
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventType, api.ErrorTypeValidationViolation)
}

func Test_PublishWithoutEventTypeVersion_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayloadWithoutEventTypeVersion()
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTypeVersion, api.ErrorTypeValidationViolation)
}

func Test_PublishWithoutEventTime_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayloadWithoutEventTime()
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTime, api.ErrorTypeValidationViolation)
}

func Test_PublishWithoutData_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayloadWithoutData()
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldData, api.ErrorTypeValidationViolation)
}

func Test_PublishWithEmptyData_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayloadWithEmptyData()
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldData, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidEventTypeVersion_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildTestPayload(test.TestSourceID, test.TestEventType, test.TestEventTypeVersionInvalid, test.TestEventID, test.TestEventTime, test.TestData)
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTypeVersion, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidEventTime_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildTestPayload(test.TestSourceID, test.TestEventType, test.TestEventTypeVersion, test.TestEventID, test.TestEventTimeInvalid, test.TestData)
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTime, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidEventId_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildTestPayload(test.TestSourceID, test.TestEventType, test.TestEventTypeVersion, test.TestEventIDInvalid, test.TestEventTime, test.TestData)
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventId, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidSourceIdInPayload_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildTestPayload(test.TestSourceIdInvalid, test.TestEventType, test.TestEventTypeVersion, test.TestEventID, test.TestEventTime, test.TestData)
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldSourceId, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidSourceIdInHeader_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayloadWithoutSourceId()
	body, statusCode := test.PerformPublishRequestWithHeaders(t, server.URL, payload, map[string]string{api.HeaderSourceId: test.TestSourceIdInvalid})

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.HeaderSourceId, api.ErrorTypeValidationViolation)
}

func Test_PublishWithTooLargePayload_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := string(make([]byte, opts.GetDefaultOptions().MaxRequestSize+1))
	body, statusCode := test.PerformPublishRequest(t, server.URL, payload)

	// assert
	test.AssertExpectedError(t, body, statusCode, http.StatusRequestEntityTooLarge, nil, api.ErrorTypeRequestBodyTooLarge)
}

func Test_PublishResponseFields(t *testing.T) {
	// prepare and send payload
	payload := test.BuildDefaultTestPayloadWithoutSourceId()
	body, statusCode := test.PerformPublishRequestWithHeaders(t, server.URL, payload, map[string]string{api.HeaderSourceId: test.TestSourceID})

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusOK, statusCode)

	// get the response
	publishResponse := &api.PublishResponse{}
	err := json.Unmarshal(body, &publishResponse)

	// assert
	assert.Nil(t, err)
	assert.Equal(t, test.TestEventID, publishResponse.EventID)
	assert.Equal(t, publisher.PUBLISHED, publishResponse.Status)
	assert.Equal(t, "Message successfully published to the channel", publishResponse.Reason)
}
