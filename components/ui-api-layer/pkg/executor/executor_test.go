package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPeriodic(t *testing.T) {
	/*
		This test is very simple - it has hardcoded time.Sleep. Running it at machine with high load may cause the test fail.
		The best way is to modify executor_test (and adapt executor) to not rely on time.Sleep.
	*/

	// GIVEN
	called := 0
	worker := func(stopCh <-chan struct{}) {
		called = called + 1
	}
	stopCh := make(chan struct{})
	svc := NewPeriodic(50*time.Millisecond, worker)

	// WHEN
	svc.Run(stopCh)
	time.Sleep(120 * time.Millisecond)
	close(stopCh)

	// THEN
	// expecting 3 calls, first at after 0ms, second - after 10ms, third after 20ms
	assert.Equal(t, 3, called)
}
