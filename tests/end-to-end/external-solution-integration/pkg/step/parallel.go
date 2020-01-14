package step

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type Parallelized struct {
	steps []Step
}

func (p *Parallelized) Name() string {
	names := make([]string, len(p.steps))
	for i, step := range p.steps {
		names[i] = step.Name()
	}
	joined := strings.Join(names, ", ")
	return fmt.Sprintf("Parallelized: %s", joined)
}

func (p *Parallelized) Run() error {
	return p.inParallel(func(step Step) error {
		return step.Run()
	})
}

func (p *Parallelized) Cleanup() error {
	return p.inParallel(func(step Step) error {
		return step.Cleanup()
	})
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

func Parallel(steps ...Step) *Parallelized {
	return &Parallelized{steps: steps}
}
