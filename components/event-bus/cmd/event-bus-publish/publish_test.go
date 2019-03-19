package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/application"
	"github.com/kyma-project/kyma/components/event-bus/internal/publish"
	test "github.com/kyma-project/kyma/components/event-bus/test/util"
	"github.com/nats-io/nats-streaming-server/server"
	"github.com/stretchr/testify/assert"

	"github.com/nats-io/go-nats-streaming"
)

const (
	clusterID  = "kyma-nats-streaming"
	iterations = 5
	interval   = 2
)

var (
	publishServer *httptest.Server
	sc            stan.Conn
	msg           *stan.Msg
)

func TestMain(m *testing.M) {
	stanServer, err := server.RunServer(clusterID)
	publishOpts := publish.DefaultOptions()
	publishApplication := application.NewPublishApplication(publishOpts)
	publishServer = httptest.NewServer(publishApplication.ServerMux)
	sc, _ = stan.Connect(clusterID, "kyma-int-test")
	if err != nil {
		panic(err)
	} else {
		retCode := m.Run()
		publishServer.Close()
		publishApplication.Stop()
		stanServer.Shutdown()
		os.Exit(retCode)
	}
}

func TestPublishWithSourceIdInPayload(t *testing.T) {
	subject, requestPayload := test.BuildDefaultTestSubjectAndPayload()
	sub, _ := sc.Subscribe(subject, func(m *stan.Msg) {
		msg = m
	})
	responseBody, statusCode := test.PerformPublishRequest(t, publishServer.URL, requestPayload)
	verifyPublish(t, statusCode, sub, responseBody, requestPayload)
}

func TestPublishWithSourceIdInHeader(t *testing.T) {
	subject := test.BuildDefaultTestSubject()
	requestPayload := test.BuildDefaultTestPayloadWithoutSourceId()
	sub, _ := sc.Subscribe(subject, func(m *stan.Msg) {
		msg = m
	})
	responseBody, statusCode := test.PerformPublishRequestWithHeaders(t, publishServer.URL, requestPayload, map[string]string{api.HeaderSourceId: test.TestSourceID})

	verifyPublish(t, statusCode, sub, responseBody, test.BuildDefaultTestPayload())
}

func TestPublishWithSourceIdInPayloadAndHeaderAndPayloadOneIsGivenPrecedence(t *testing.T) {
	subject := test.BuildDefaultTestSubject()
	requestPayload := test.BuildDefaultTestPayload()
	sub, _ := sc.Subscribe(subject, func(m *stan.Msg) {
		msg = m
	})
	responseBody, statusCode := test.PerformPublishRequestWithHeaders(t, publishServer.URL, requestPayload, map[string]string{api.HeaderSourceId: test.TestSourceIDInHeader})
	verifyPublish(t, statusCode, sub, responseBody, requestPayload)
}

func verifyPublish(t *testing.T, statusCode int, sub stan.Subscription, responseBody []byte, requestPayload string) {
	if statusCode != http.StatusOK {
		t.Errorf("Status code is wrong, have: %d, want: %d", statusCode, http.StatusOK)
	}
	respObj := &api.PublishResponse{}
	err := json.Unmarshal(responseBody, &respObj)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, respObj.EventID)
	assert.NotEmpty(t, respObj.EventID)
	i := 1
	for msg == nil {
		if i > iterations {
			t.Error("Test failed")
			break
		}
		log.Printf("Waiting for receiving published message [%d/%d].", i, iterations)
		time.Sleep(interval * time.Second)
		i++
	}
	test.VerifyReceivedMsg(t, requestPayload, msg.Data)
	sub.Unsubscribe()
}

func TestStatus(t *testing.T) {
	log.Println("started nats and publish app")
	res, err := http.Get(publishServer.URL + "/v1/status/ready")
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)
}

func TestPublishWithBadPayload(t *testing.T) {
	payload := test.BuildDefaultTestBadPayload()
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, nil, api.ErrorTypeBadPayload)
}

func TestPublishWithoutSourceId(t *testing.T) {
	payload := test.BuildDefaultTestPayloadWithoutSourceId()
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldSourceId+"/"+api.HeaderSourceId, api.ErrorTypeValidationViolation)
}

func TestPublishWithoutEventType(t *testing.T) {
	payload := test.BuildDefaultTestPayloadWithoutEventType()
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventType, api.ErrorTypeValidationViolation)
}

func TestPublishWithoutEventTypeVersion(t *testing.T) {
	payload := test.BuildDefaultTestPayloadWithoutEventTypeVersion()
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTypeVersion, api.ErrorTypeValidationViolation)
}

func TestPublishWithoutEventTime(t *testing.T) {
	payload := test.BuildDefaultTestPayloadWithoutEventTime()
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTime, api.ErrorTypeValidationViolation)
}

func TestPublishWithoutData(t *testing.T) {
	payload := test.BuildDefaultTestPayloadWithoutData()
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldData, api.ErrorTypeValidationViolation)
}

func TestPublishWithEmptyData(t *testing.T) {
	payload := test.BuildDefaultTestPayloadWithEmptyData()
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldData, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidEventTypeVersion(t *testing.T) {
	payload := test.BuildTestPayload(test.TestSourceID, test.TestEventType, test.TestEventTypeVersionInvalid, test.TestEventID,
		test.TestEventTime, test.TestData)
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTypeVersion, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidEventTime(t *testing.T) {
	payload := test.BuildTestPayload(test.TestSourceID, test.TestEventType, test.TestEventTypeVersion, test.TestEventID,
		test.TestEventTimeInvalid, test.TestData)
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTime, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidEventId(t *testing.T) {
	payload := test.BuildTestPayload(test.TestSourceID, test.TestEventType, test.TestEventTypeVersion, test.TestEventIDInvalid,
		test.TestEventTime, test.TestData)
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventId, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidSourceIdInPayload(t *testing.T) {
	payload := test.BuildTestPayload(test.TestSourceIdInvalid, test.TestEventType, test.TestEventTypeVersion, test.TestEventID,
		test.TestEventTime, test.TestData)
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldSourceId, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidSourceIdInHeader(t *testing.T) {
	payload := test.BuildDefaultTestPayloadWithoutSourceId()
	body, statusCode := test.PerformPublishRequestWithHeaders(t, publishServer.URL, payload, map[string]string{api.HeaderSourceId: test.TestSourceIdInvalid})
	test.AssertExpectedError(t, body, statusCode, http.StatusBadRequest, api.HeaderSourceId, api.ErrorTypeValidationViolation)
}

func TestPublishWithTooLargePayload(t *testing.T) {
	payload := string(make([]byte, publish.DefaultOptions().MaxRequestSize+1))
	body, statusCode := test.PerformPublishRequest(t, publishServer.URL, payload)
	test.AssertExpectedError(t, body, statusCode, http.StatusRequestEntityTooLarge, nil, api.ErrorTypeRequestBodyTooLarge)
}
