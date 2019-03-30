package waiter

import (
	"time"
	"k8s.io/apimachinery/pkg/util/wait"
)

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
	return wait.PollUntil(500*time.Millisecond, conditionFunc, stopCh)
}