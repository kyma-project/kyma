package testsuite

import "time"

// Sleep is a step sleeps for some time
type Sleep struct {
	d time.Duration
}

// Name returns name name of the step
func (s Sleep) Name() string {
	return "sleep"
}

// Run executes the step
func (s Sleep) Run() error {
	time.Sleep(s.d)
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s Sleep) Cleanup() error {
	return nil
}

// NewSleep returns new Sleep
func NewSleep(d time.Duration) *Sleep {
	return &Sleep{d: d}
}
