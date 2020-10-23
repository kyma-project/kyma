package step

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/hashicorp/go-multierror"
)

type Parallelized struct {
	steps []Step
	name  string
	logf  *logrus.Entry
}

func NewParallelRunner(logf *logrus.Entry, name string, steps ...Step) *Parallelized {
	return &Parallelized{logf: logf, name: name, steps: steps}
}

func (p *Parallelized) Name() string {
	names := make([]string, len(p.steps))
	for i, step := range p.steps {
		names[i] = step.Name()
	}
	joined := strings.Join(names, ", ")
	return fmt.Sprintf("Parallel: %s, %s", p.name, joined)
}

func (p *Parallelized) Run() error {
	return p.inParallel(func(step Step) error {
		p.logf.Infof("Run in parallel: %s", step.Name())
		err := step.Run()
		if err != nil {
			p.logf.Errorf("Step: %s, returned error: %s", step.Name(), err.Error())
			if callbackErr := step.OnError(); callbackErr != nil {
				p.logf.Errorf("while executing OnError on %s, err: %s", step.Name(), callbackErr.Error())
			}
			return errors.Wrapf(err, "while executing step: %s", step.Name())
		}
		return nil
	})
}

func (p *Parallelized) Cleanup() error {
	return p.inParallel(func(step Step) error {
		return step.Cleanup()
	})
}

func (p *Parallelized) OnError() error {
	return nil
}

func (p *Parallelized) inParallel(f func(step Step) error) error {
	wg := &sync.WaitGroup{}
	wg.Add(len(p.steps))
	errs := make(chan error, len(p.steps))
	for _, s := range p.steps {
		go p.runStepInParallel(wg, errs, s, f)
	}
	wg.Wait()
	close(errs)
	var errAll *multierror.Error
	for err := range errs {
		errAll = multierror.Append(errAll, err)
	}
	return errAll.ErrorOrNil()
}

func (p *Parallelized) runStepInParallel(wg *sync.WaitGroup, errs chan<- error, step Step, f func(step Step) error) {
	defer wg.Done()
	defer func() {
		if err := recover(); err != nil {
			errs <- err.(error)
		}
	}()
	errs <- f(step)
}
