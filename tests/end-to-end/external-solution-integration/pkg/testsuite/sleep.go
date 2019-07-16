package testsuite

import "time"

type Sleep struct {
	d time.Duration
}

func (s Sleep) Name() string {
	return "sleep"
}

func (s Sleep) Run() error {
	time.Sleep(s.d)
	return nil
}

func (s Sleep) Cleanup() error {
	return nil
}

func NewSleep(d time.Duration) *Sleep {
	return &Sleep{d:d}
}


