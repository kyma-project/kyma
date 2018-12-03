package executor

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPeriodic(t *testing.T) {
	// GIVEN
	var called int32 = 0

	worker := func(stopCh <-chan struct{}) {
		atomic.AddInt32(&called, 1)
	}
	stopCh := make(chan struct{})
	svc := NewPeriodic(50*time.Millisecond, worker)

	// WHEN
	svc.Run(stopCh)
	time.Sleep(120 * time.Millisecond)
	close(stopCh)

	// THEN
	// expecting 3 calls, first at after 0ms, second - after 10ms, third after 20ms
	assert.Equal(t, int32(3), atomic.LoadInt32(&called))
}
