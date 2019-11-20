package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/publisher"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/test/fake"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/publish/opts"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
	testV1 "github.com/kyma-project/kyma/components/event-bus/test/util/v1"
	testV2 "github.com/kyma-project/kyma/components/event-bus/test/util/v2"
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

func Test_PublishV1WithSourceIdInPayload_ShouldSucceed(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayload()
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusOK, statusCode)

	// get the response
	publishResponse := &api.Response{}
	err := json.Unmarshal(body, &publishResponse)

	// assert
	assert.Nil(t, err)
	assert.Equal(t, testV1.TestEventID, publishResponse.EventID)
}

func Test_PublishV1WithSourceIdInHeader_ShouldSucceed(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayloadWithoutSourceID()
	body, statusCode := testV1.PerformPublishV1RequestWithHeaders(t, server.URL, payload, map[string]string{api.HeaderSourceID: testV1.TestSourceID})

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusOK, statusCode)

	// get the response
	publishResponse := &api.Response{}
	err := json.Unmarshal(body, &publishResponse)

	// assert
	assert.Nil(t, err)
	assert.Equal(t, testV1.TestEventID, publishResponse.EventID)
}

func Test_PublishWithBadPayload_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestBadPayload()
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, nil, api.ErrorTypeBadPayload)
}

func Test_PublishWithoutSourceId_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayloadWithoutSourceID()
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, fmt.Sprintf("%v/%v", api.FieldSourceID, api.HeaderSourceID), api.ErrorTypeValidationViolation)
}

func Test_PublishWithoutEventType_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayloadWithoutEventType()
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventType, api.ErrorTypeValidationViolation)
}

func Test_PublishWithoutEventTypeVersion_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayloadWithoutEventTypeVersion()
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTypeVersion, api.ErrorTypeValidationViolation)
}

func Test_PublishWithoutEventTime_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayloadWithoutEventTime()
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTime, api.ErrorTypeValidationViolation)
}

func Test_PublishWithoutData_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayloadWithoutData()
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldData, api.ErrorTypeValidationViolation)
}

func Test_PublishWithEmptyData_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayloadWithEmptyData()
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldData, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidEventTypeVersion_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildTestPayload(testV1.TestSourceID, testV1.TestEventType, testV1.TestEventTypeVersionInvalid, testV1.TestEventID, testV1.TestEventTime, testV1.TestData)
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTypeVersion, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidEventTime_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildTestPayload(testV1.TestSourceID, testV1.TestEventType, testV1.TestEventTypeVersion, testV1.TestEventID, testV1.TestEventTimeInvalid, testV1.TestData)
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTime, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidEventId_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildTestPayload(testV1.TestSourceID, testV1.TestEventType, testV1.TestEventTypeVersion, testV1.TestEventIDInvalid, testV1.TestEventTime, testV1.TestData)
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventID, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidSourceIdInPayload_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildTestPayload(testV1.TestSourceIDInvalid, testV1.TestEventType, testV1.TestEventTypeVersion, testV1.TestEventID, testV1.TestEventTime, testV1.TestData)
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldSourceID, api.ErrorTypeValidationViolation)
}

func Test_PublishWithInvalidSourceIdInHeader_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayloadWithoutSourceID()
	body, statusCode := testV1.PerformPublishV1RequestWithHeaders(t, server.URL, payload, map[string]string{api.HeaderSourceID: testV1.TestSourceIDInvalid})

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.HeaderSourceID, api.ErrorTypeValidationViolation)
}

func Test_PublishWithTooLargePayload_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := string(make([]byte, opts.GetDefaultOptions().MaxRequestSize+1))
	body, statusCode := testV1.PerformPublishV1Request(t, server.URL, payload)

	// assert
	testV1.AssertExpectedError(t, body, statusCode, http.StatusRequestEntityTooLarge, nil, api.ErrorTypeRequestBodyTooLarge)
}

func Test_PublishV2WithCorrectPayload_ShouldSuceed(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayload(testV2.TestSource, testV2.TestType, testV2.TestEventTypeVersion, testV2.TestEventID, testV2.TestEventTime, testV2.TestData)
	body, statusCode := testV2.PerformPublishV2RequestWithHeaders(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusOK, statusCode)
}

func Test_PublishV2WithInvalidSpecVersion_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithInvalidSpecversion()
	body, statusCode := testV2.PerformPublishV2RequestWithHeaders(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}
func Test_PublishV2WithoutAnyCEFields_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildV2PayloadWithoutCEFields()
	body, statusCode := testV2.PerformPublishV2RequestWithHeaders(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func Test_PublishV2WithoutAnyCEId_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithoutID()
	body, statusCode := testV2.PerformPublishV2RequestWithHeaders(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func Test_PublishV2WithoutAnyCESource_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithoutSource()
	body, statusCode := testV2.PerformPublishV2RequestWithHeaders(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func Test_PublishV2WithoutAnyCEWithoutSpecversion_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithoutSpecversion()
	body, statusCode := testV2.PerformPublishV2RequestWithHeaders(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func Test_PublishV2WithoutAnyCEWithoutType_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithoutType()
	body, statusCode := testV2.PerformPublishV2RequestWithHeaders(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func Test_PublishV1ResponseFields(t *testing.T) {
	// prepare and send payload
	payload := testV1.BuildDefaultTestPayloadWithoutSourceID()
	body, statusCode := testV1.PerformPublishV1RequestWithHeaders(t, server.URL, payload, map[string]string{api.HeaderSourceID: testV1.TestSourceID})

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusOK, statusCode)

	// get the response
	publishResponse := &api.Response{}
	err := json.Unmarshal(body, &publishResponse)

	// assert
	assert.Nil(t, err)
	assert.Equal(t, testV1.TestEventID, publishResponse.EventID)
	assert.Equal(t, publisher.Published, publishResponse.Status)
	assert.Equal(t, "Message successfully published to the channel", publishResponse.Reason)
}
