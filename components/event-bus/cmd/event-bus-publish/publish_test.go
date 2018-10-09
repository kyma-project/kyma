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
	subject, requestPayload := buildDefaultTestSubjectAndPayload()
	sub, _ := sc.Subscribe(subject, func(m *stan.Msg) {
		msg = m
	})
	responseBody, statusCode := performPublishRequest(t, publishServer.URL, requestPayload)
	verifyPublish(t, statusCode, sub, responseBody, requestPayload)
}

func TestPublishWithSourceIdInHeader(t *testing.T) {
	subject := buildDefaultTestSubject()
	requestPayload := buildDefaultTestPayloadWithoutSourceId()
	sub, _ := sc.Subscribe(subject, func(m *stan.Msg) {
		msg = m
	})
	responseBody, statusCode := performPublishRequestWithHeaders(t, publishServer.URL, requestPayload, map[string]string{api.HeaderSourceId: testSourceID})

	verifyPublish(t, statusCode, sub, responseBody, buildDefaultTestPayload())
}

func TestPublishWithSourceIdInPayloadAndHeaderAndPayloadOneIsGivenPrecedence(t *testing.T) {
	subject := buildDefaultTestSubject()
	requestPayload := buildDefaultTestPayload()
	sub, _ := sc.Subscribe(subject, func(m *stan.Msg) {
		msg = m
	})
	responseBody, statusCode := performPublishRequestWithHeaders(t, publishServer.URL, requestPayload, map[string]string{api.HeaderSourceId: testSourceIDInHeader})
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
	verifyReceivedMsg(t, requestPayload, msg.Data)
	sub.Unsubscribe()
}

func TestStatus(t *testing.T) {
	log.Println("started nats and publish app")
	res, err := http.Get(publishServer.URL + "/v1/status/ready")
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)
}

func TestPublishWithBadPayload(t *testing.T) {
	payload := buildDefaultTestBadPayload()
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, nil, api.ErrorTypeBadPayload)
}

func TestPublishWithoutSourceId(t *testing.T) {
	payload := buildDefaultTestPayloadWithoutSourceId()
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldSourceId+"/"+api.HeaderSourceId, api.ErrorTypeValidationViolation)
}

func TestPublishWithoutEventType(t *testing.T) {
	payload := buildDefaultTestPayloadWithoutEventType()
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventType, api.ErrorTypeValidationViolation)
}

func TestPublishWithoutEventTypeVersion(t *testing.T) {
	payload := buildDefaultTestPayloadWithoutEventTypeVersion()
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTypeVersion, api.ErrorTypeValidationViolation)
}

func TestPublishWithoutEventTime(t *testing.T) {
	payload := buildDefaultTestPayloadWithoutEventTime()
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTime, api.ErrorTypeValidationViolation)
}

func TestPublishWithoutData(t *testing.T) {
	payload := buildDefaultTestPayloadWithoutData()
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldData, api.ErrorTypeValidationViolation)
}

func TestPublishWithEmptyData(t *testing.T) {
	payload := buildDefaultTestPayloadWithEmptyData()
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldData, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidEventTypeVersion(t *testing.T) {
	payload := buildTestPayload(testSourceID, testEventType, testEventTypeVersionInvalid, testEventID,
		testEventTime, testData)
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTypeVersion, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidEventTime(t *testing.T) {
	payload := buildTestPayload(testSourceID, testEventType, testEventTypeVersion, testEventID,
		testEventTimeInvalid, testData)
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventTime, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidEventId(t *testing.T) {
	payload := buildTestPayload(testSourceID, testEventType, testEventTypeVersion, testEventIDInvalid,
		testEventTime, testData)
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldEventId, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidSourceIdInPayload(t *testing.T) {
	payload := buildTestPayload(testSourceIdInvalid, testEventType, testEventTypeVersion, testEventID,
		testEventTime, testData)
	body, statusCode := performPublishRequest(t, publishServer.URL, payload)
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.FieldSourceId, api.ErrorTypeValidationViolation)
}

func TestPublishInvalidSourceIdInHeader(t *testing.T) {
	payload := buildDefaultTestPayloadWithoutSourceId()
	body, statusCode := performPublishRequestWithHeaders(t, publishServer.URL, payload, map[string]string{api.HeaderSourceId: testSourceIdInvalid})
	assertExpectedError(t, body, statusCode, http.StatusBadRequest, api.HeaderSourceId, api.ErrorTypeValidationViolation)
}
