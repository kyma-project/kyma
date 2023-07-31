package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/histogram/mocks"
)

func TestNewCollector(t *testing.T) {
	// given
	latency := new(mocks.BucketsProvider)
	latency.On("Buckets").Return(nil)

	// when
	collector := NewCollector(latency)

	// then
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.backendLatency)
	assert.NotNil(t, collector.backendLatency.MetricVec)
	assert.NotNil(t, collector.eventType)
	assert.NotNil(t, collector.eventType.MetricVec)
	latency.AssertExpectations(t)
}
