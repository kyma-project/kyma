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
	assert.NotNil(t, collector.errors)
	assert.NotNil(t, collector.errors.MetricVec)
	assert.NotNil(t, collector.latency)
	assert.NotNil(t, collector.latency.MetricVec)
	assert.NotNil(t, collector.eventType)
	assert.NotNil(t, collector.eventType.MetricVec)
	assert.NotNil(t, collector.requests)
	assert.NotNil(t, collector.requests.MetricVec)

	// then
	latency.AssertCalled(t, bucketsFunc)
	latency.AssertNumberOfCalls(t, bucketsFunc, 1)
	latency.AssertExpectations(t)
}
