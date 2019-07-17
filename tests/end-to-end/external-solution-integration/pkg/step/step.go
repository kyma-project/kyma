package step

import (
	"github.com/hashicorp/errwrap"
	"github.com/sirupsen/logrus"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

// Step represents a single action in test scenario
type Step interface {
	// Name returns name name of the step
	Name() string
	// Run executes the step
	Run() error
	// Cleanup removes all resources that may possibly created by the step
	Cleanup() error
}

// Runner executes steps in safe manner
type Runner struct {
	log     *logrus.Logger
	cleanup CleanupMode
}

// NewRunner returns new runner
func NewRunner() *Runner {
	log := logrus.New()
	log.SetReportCaller(false)
	return &Runner{
		log:     log,
		cleanup: CleanupMode_Yes,
	}
}

// Run executes steps in specified order. If skipCleanup is false it also executes Step.Cleanup in reverse order
// starting from last executed step
func (r *Runner) Run(steps []Step, skipCleanup bool) error {
	var startedStep int
	var step Step
	var err error

	defer func() {
		if !skipCleanup {
			r.Cleanup(steps[0:(startedStep + 1)])
		}
	}()

	for startedStep, step = range steps {
		r.log.Infof("Step: '%s'", step.Name())
		if err = r.runStep(step); err != nil {
			r.log.Errorf("Error in '%s': %s", step.Name(), err)
			break
		}
	}

	return err
}

// runStep allows to recover in case of panic in step
func (r *Runner) runStep(step Step) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	return step.Run()
}

// Cleanup cleans up given steps in reverse order
func (r *Runner) Cleanup(steps []Step) {
	for i := len(steps) - 1; i >= 0; i-- {
		r.log.Infof("Cleanup: '%s'", steps[i].Name())
		if err := steps[i].Cleanup(); err != nil && !isNotFound(err) {
			r.log.Warnf("Error during '%s' cleanup: %s", steps[i].Name(), err)
		}
	}
}

func isNotFound(err error) bool {
	isNotFound := true
	errwrap.Walk(err, func(e error) {
		if isNotFound && !k8s_errors.IsNotFound(e) {
			isNotFound = false
		}
	})
	return isNotFound
}
