package executor

import (
	"time"
)

type periodic struct {
	refreshPeriod time.Duration
	executionFunc func(stopCh <-chan struct{})
}

// NewPeriodic creates a periodic executor, which calls given executionFunc periodically.
func NewPeriodic(period time.Duration, executionFunc func(stopCh <-chan struct{})) *periodic {
	return &periodic{
		refreshPeriod: period,
		executionFunc: executionFunc,
	}
}

// Run starts the periodic work
func (e *periodic) Run(stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(e.refreshPeriod)
		for {
			e.executionFunc(stopCh)
			select {
			case <-stopCh:
				ticker.Stop()
				return
			case <-ticker.C:
			}
		}
	}()
}
