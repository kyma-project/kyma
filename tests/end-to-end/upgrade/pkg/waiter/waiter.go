package waiter

import (
	"time"
	"k8s.io/apimachinery/pkg/util/wait"
)

func New(timeout time.Duration, conditionFunc wait.ConditionFunc, stop <-chan struct{}) error {
	timeoutCh := time.After(timeout)
	stopCh := make(chan struct{})
	go func() {
		select {
		case <-timeoutCh:
			close(stopCh)
		case <-stop:
			close(stopCh)
		}
	}()
	return wait.PollUntil(500*time.Millisecond, conditionFunc, stopCh)
}