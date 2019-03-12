package runner

import "github.com/sirupsen/logrus"

// Task is a wrapper around the UpgradeTest which orchestrate it with its name
type task struct {
	name string
	UpgradeTest
}

// Name returns task name
func (t *task) Name() string {
	return t.name
}

// taskFn is a signature for task function.
// Required to unify the way how UpgradeTest methods are executed.
type taskFn func(stopCh <-chan struct{}, log logrus.FieldLogger, namespace string) error
