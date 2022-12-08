package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/histogram/mocks"
)

func TestNewCollector(t *testing.T) {
	// given
	const bucketsFunc = "Buckets"
	latency := new(mocks.BucketsProvider)
	latency.On(bucketsFunc).Return(nil)
	latency.Test(t)

	// when
	collector := NewCollector(latency)

	// then
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.backendErrors)
	assert.NotNil(t, collector.backendErrors.MetricVec)
	assert.NotNil(t, collector.backendLatency)
	assert.NotNil(t, collector.backendLatency.MetricVec)
	assert.NotNil(t, collector.eventType)
	assert.NotNil(t, collector.eventType.MetricVec)
	assert.NotNil(t, collector.backendRequests)
	assert.NotNil(t, collector.backendRequests.MetricVec)
	latency.AssertCalled(t, bucketsFunc)
	latency.AssertNumberOfCalls(t, bucketsFunc, 2)
	latency.AssertExpectations(t)
}
