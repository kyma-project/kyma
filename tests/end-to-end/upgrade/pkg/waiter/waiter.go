package waiter

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// WaitAtMost function is a util for waiting some maximum time for some action
func WaitAtMost(conditionFunc wait.ConditionFunc, timeout time.Duration, stop <-chan struct{}) error {
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
	return wait.PollUntil(1000*time.Millisecond, conditionFunc, stopCh)
}
