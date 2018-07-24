package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_convertToCloudEvent_Missing_Payload(t *testing.T) {
	_, err := convertToCloudEvent(nil)
	assert.NotNil(t, err)
}

func Test_convertToCloudEvent_Invalid_Payload(t *testing.T) {
	payload := []byte("{\"key\":\"value\"}")
	_, err := convertToCloudEvent(&payload)
	assert.NotNil(t, err)
}

func Test_convertToCloudEvent_Success(t *testing.T) {
	payload := []byte("{\"source\":{\"source-namespace\":\"test-source-namespace\",\"source-type\":\"test-source-type\",\"source-environment\":\"test-source-environment\"},\"event-type\":\"test-event-type\",\"event-type-version\":\"v1\",\"event-id\":\"4ea567cf-812b-49d9-a4b2-cb5ddf464094\",\"event-time\":\"2012-11-01T22:08:41+00:00\",\"data\":\"{'key':'value'}\",\"extensions\":{\"trace-context\":{\"x-b3-flags\":\"0\",\"x-b3-parentspanid\":\"2c94f88423efcd96\",\"x-b3-sampled\":\"true\",\"x-b3-spanid\":\"61c8973b3ccb7417\",\"x-b3-traceid\":\"4ef497da28c4bc6a1c0202f1b0e342db\"}}}")
	_, err := convertToCloudEvent(&payload)
	assert.Nil(t, err)
}

func Test_getTraceContext_Missing_CloudEvent(t *testing.T) {
	traceContext := getTraceContext(nil)
	assert.Nil(t, traceContext)
}

func Test_getTraceContext_Without_TraceContext(t *testing.T) {
	payload := []byte("{\"source\":{\"source-namespace\":\"test-source-namespace\",\"source-type\":\"test-source-type\",\"source-environment\":\"test-source-environment\"},\"event-type\":\"test-event-type\",\"event-type-version\":\"v1\",\"event-id\":\"4ea567cf-812b-49d9-a4b2-cb5ddf464094\",\"event-time\":\"2012-11-01T22:08:41+00:00\",\"data\":\"{'key':'value'}\"}")
	cloudEvent, err := convertToCloudEvent(&payload)
	traceContext := getTraceContext(cloudEvent)

	assert.Nil(t, err)
	assert.NotNil(t, cloudEvent)
	assert.Nil(t, traceContext)
}

func Test_getTraceContext_With_TraceContext(t *testing.T) {
	payload := []byte("{\"source\":{\"source-namespace\":\"test-source-namespace\",\"source-type\":\"test-source-type\",\"source-environment\":\"test-source-environment\"},\"event-type\":\"test-event-type\",\"event-type-version\":\"v1\",\"event-id\":\"4ea567cf-812b-49d9-a4b2-cb5ddf464094\",\"event-time\":\"2012-11-01T22:08:41+00:00\",\"data\":\"{'key':'value'}\",\"extensions\":{\"trace-context\":{\"x-b3-flags\":\"0\",\"x-b3-parentspanid\":\"2c94f88423efcd96\",\"x-b3-sampled\":\"true\",\"x-b3-spanid\":\"61c8973b3ccb7417\",\"x-b3-traceid\":\"4ef497da28c4bc6a1c0202f1b0e342db\"}}}")
	cloudEvent, err := convertToCloudEvent(&payload)
	traceContext := getTraceContext(cloudEvent)

	assert.Nil(t, err)
	assert.NotNil(t, cloudEvent)
	assert.NotNil(t, traceContext)
}
