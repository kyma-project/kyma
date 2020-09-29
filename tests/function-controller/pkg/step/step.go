package step

import (
	"github.com/hashicorp/errwrap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// Step represents a single action in test scenario
type Step interface {
	// Name returns Name Name of the step
	Name() string
	// Run executes the step
	Run() error
	// Cleanup removes all resources that may possibly created by the step
	Cleanup() error
	// OnError is callback in case of error
	OnError(error) error
}

// Runner executes Steps in safe manner
type Runner struct {
	log     *logrus.Logger
	cleanup CleanupMode
}

type RunnerOption func(runner *Runner)

func WithCleanupDefault(mode CleanupMode) RunnerOption {
	return func(runner *Runner) {
		runner.cleanup = mode
	}
}

func WithLogger(logger *logrus.Logger) RunnerOption {
	return func(runner *Runner) {
		runner.log = logger
	}
}

// NewRunner returns new runner
func NewRunner(opts ...RunnerOption) *Runner {
	log := logrus.New()
	log.SetReportCaller(false)

	runner := &Runner{
		log:     log,
		cleanup: CleanupModeYes,
	}
	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

// Run executes Steps in specified order. If skipCleanup is false it also executes Step.Cleanup in reverse order
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
			callbackErr := step.OnError(err)
			if callbackErr != nil {
				//TODO: in case of callbackErr
			}
			return err
		}
	}

	return err
}

// runStep allows to recover in case of panic in step
func (r *Runner) runStep(step Step) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.WithStack(e.(error))
		}
	}()
	return step.Run()
}

// Cleanup cleans up given Steps in reverse order
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
		if isNotFound && !k8serrors.IsNotFound(e) {
			isNotFound = false
		}
	})
	return isNotFound
}
