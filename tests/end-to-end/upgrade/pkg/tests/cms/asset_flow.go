package cms

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"time"
	"k8s.io/apimachinery/pkg/util/wait"
)

type baseFlow struct {
	namespace string
	log       logrus.FieldLogger
	stop      <-chan struct{}

	dynamicInterface dynamic.Interface
}

func (f *baseFlow) wait(timeout time.Duration, conditionFunc wait.ConditionFunc) error {
	timeoutCh := time.After(timeout)
	stopCh := make(chan struct{})
	go func() {
		select {
		case <-timeoutCh:
			close(stopCh)
		case <-f.stop:
			close(stopCh)
		}
	}()
	return wait.PollUntil(500*time.Millisecond, conditionFunc, stopCh)
}