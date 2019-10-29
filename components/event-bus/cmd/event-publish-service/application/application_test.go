package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

type testInput struct {
	specVersion string
	id          string
	typ         string
	typeVersion string
	source      string
	time        string
	data        string
}

type testExpectation struct {
	statusCode int
	emptyBody  bool
	// TODO(nachtmaar): validate api.Response object in success case
	// apiError api.Error
	// TODO(nachtmaar): validate publish.Error object in error case
	// apiResponse api.Response
}

var tableTest = []struct {
	testInput       testInput
	testExpectation testExpectation
	name            string
}{
	{
		name: "valid CE v0.3",
		testInput: testInput{
			specVersion: "0.3",
			id:          testV2.TestEventID,
			typ:         testV2.TestType,
			typeVersion: testV2.TestEventTypeVersion,
			source:      testV2.TestSource,
			data:        testV2.TestData,
			time:        testV2.TestEventTime,
		},
		testExpectation: testExpectation{
			statusCode: 200,
			emptyBody:  false,
		},
	},
	{
		name: "valid CE v1.0",
		testInput: testInput{
			specVersion: "1.0",
			id:          testV2.TestEventID,
			typ:         testV2.TestType,
			typeVersion: testV2.TestEventTypeVersion,
			source:      testV2.TestSource,
			data:        testV2.TestData,
			time:        testV2.TestEventTime,
		},
		testExpectation: testExpectation{
			statusCode: 200,
			emptyBody:  false,
		},
	},
	// these tests are a replacement for the tests in api/publish/v2/validators_test.go
	{
		name: "invalid CE v0.3 no source id",
		testInput: testInput{
			specVersion: "0.3",
			id:          testV2.TestEventID,
			typ:         testV2.TestType,
			typeVersion: testV2.TestEventTypeVersion,
			// ups ...
			source: "",
			data:   testV2.TestData,
			time:   testV2.TestEventTime,
		},
		testExpectation: testExpectation{
			statusCode: 400,
			emptyBody:  false,
			// TODO(nachtmaar): something like this maybe ?
			// apiError: api.Error{
			// 	Status:   400,
			// 	Type:     "",
			// 	Message:  "Some unexpected internal error occurred, please contact support.",
			// 	MoreInfo: "",
			// 	Details: []api.ErrorDetail{
			// 		{
			// 			Field:    "",
			// 			Type:     "",
			// 			Message:  "source: REQUIRED",
			// 			MoreInfo: "",
			// 		},
			// 	},
			// },
		},
	},
}

func TestTable(t *testing.T) {
	for _, tt := range tableTest {
		t.Run(tt.name+"_structured", func(t *testing.T) {
			testCloudEventStructuredEncoding(t, tt.testInput, tt.testExpectation)
		})
		t.Run(tt.name+"_binary", func(t *testing.T) {
			testCloudEventBinaryEncoding(t, tt.testInput, tt.testExpectation)
		})
	}
}

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

// test cloudevents structured encoding
func testCloudEventStructuredEncoding(t *testing.T, testInput testInput, testExpectation testExpectation) {
	payload := testV2.BuildPublishV2TestPayload(testInput.source, testInput.typ, testInput.typeVersion, testInput.id, testInput.time, testInput.data)
	body, statusCode := testV2.PerformPublishV2RequestStructured(t, server.URL, payload)
	handleTableTestResult(t, testExpectation, body, statusCode)
}

// test cloudevents binary encoding
func testCloudEventBinaryEncoding(t *testing.T, testInput testInput, testExpectation testExpectation) {
	headers := testV2.BuildPublishV2TestHeader(testInput.source, testInput.typ, testInput.typeVersion, testInput.id, testInput.time)
	body, statusCode := testV2.PerformPublishV2RequestBinary(t, server.URL, testInput.data, headers)
	handleTableTestResult(t, testExpectation, body, statusCode)
}

// Implements the validation logic for a `testExpectation` based on a `testInput` HTTP request to /v2 endpoint
func handleTableTestResult(t *testing.T, testExpectation testExpectation, body []byte, statusCode int) {
	// check body
	if testExpectation.emptyBody {
		assert.Nil(t, body)
	} else {
		assert.NotNil(t, body)
	}

	// check status code
	assert.Equal(t, testExpectation.statusCode, statusCode)

	// valid successful http request
	if (statusCode / 100) == 2 {
		// get the response
		publishResponse := api.Response{}
		err := json.Unmarshal(body, &publishResponse)
		assert.Nil(t, err)
		fmt.Println(publishResponse)
	} else {
		// valid unsuccessful http request
		apiError := api.Error{}

		err := json.Unmarshal(body, &apiError)
		assert.Nil(t, err)

		// status code is also expected in returned error object
		assert.Equal(t, testExpectation.statusCode, apiError.Status)
		fmt.Println(apiError)
	}
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

func Test_PublishV2WithCorrectPayloadBinary_ShouldSuceed(t *testing.T) {
	// prepare and send payload
	headers := testV2.BuildPublishV2TestHeader(testV2.TestSource, testV2.TestType, testV2.TestEventTypeVersion, testV2.TestEventID, testV2.TestEventTime)
	body, statusCode := testV2.PerformPublishV2RequestBinary(t, server.URL, testV2.TestData, headers)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusOK, statusCode)
}

func Test_PublishV2WithCorrectPayloadStructured_ShouldSuceed(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayload(testV2.TestSource, testV2.TestType, testV2.TestEventTypeVersion, testV2.TestEventID, testV2.TestEventTime, testV2.TestData)
	body, statusCode := testV2.PerformPublishV2RequestStructured(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusOK, statusCode)
}

func Test_PublishV2WithInvalidSpecVersion_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithInvalidSpecversion()
	body, statusCode := testV2.PerformPublishV2RequestStructured(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

// CloudEvents structured mode
func Test_PublishV2WithoutAnyCEFields_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildV2PayloadWithoutCEFields()
	body, statusCode := testV2.PerformPublishV2RequestStructured(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func Test_PublishV2WithoutAnyCEId_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithoutID()
	body, statusCode := testV2.PerformPublishV2RequestStructured(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func Test_PublishV2WithoutAnyCESource_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithoutSource()
	body, statusCode := testV2.PerformPublishV2RequestStructured(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func Test_PublishV2WithoutAnyCEWithoutSpecversion_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithoutSpecversion()
	body, statusCode := testV2.PerformPublishV2RequestStructured(t, server.URL, payload)

	// assert
	assert.NotNil(t, body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}

func Test_PublishV2WithoutAnyCEWithoutType_ShouldFail(t *testing.T) {
	// prepare and send payload
	payload := testV2.BuildPublishV2TestPayloadWithoutType()
	body, statusCode := testV2.PerformPublishV2RequestStructured(t, server.URL, payload)

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
	assert.Equal(t, api.Published, publishResponse.Status)
	assert.Equal(t, "Message successfully published to the channel", publishResponse.Reason)
}
