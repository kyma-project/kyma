package step

import (
	"github.com/sirupsen/logrus"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

type Step interface {
	Name() string
	Run() error
	Cleanup() error
}

type Runner struct {
	log *logrus.Logger
}

func NewRunner() *Runner {
	log := logrus.New()
	log.SetReportCaller(false)
	return &Runner{log: log}
}

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

func (r *Runner) runStep(step Step) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	return step.Run()
}

func (r *Runner) Cleanup(steps []Step) {
	for i := len(steps)-1; i >= 0; i-- {
		r.log.Infof("Cleanup: '%s'", steps[i].Name())
		if err := steps[i].Cleanup(); err != nil && !k8s_errors.IsNotFound(err) {
			r.log.Warnf("Error during '%s' cleanup: %s", steps[i].Name(), err)
		}
	}
}
