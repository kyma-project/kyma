package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Scheduler takes steps and executes them
type Scheduler struct {
	logger *zap.Logger
}

// New returns a new scheduler
func New(logger *zap.Logger) *Scheduler {
	return &Scheduler{
		logger: logger,
	}
}

// Defines contract for scheduler
type (
	StepFunc func(ctx context.Context) error

	Step struct {
		Name string
		Do   StepFunc
	}

	Steps []Step
)

// MustExecute takes steps and executes them
func (s *Scheduler) MustExecute(ctx context.Context, steps Steps) {
	var (
		logger = s.logger.Sugar().With("execution_id", uuid.New())

		logStart = logger.With("action", "start").Info
		logDone  = logger.With("action", "done").Info
		logFatal = logger.With("action", "failed").Fatalw

		stepCounter = 0
		startTime   = time.Now()
	)

	logger.Infof("Started execution of %d steps...", len(steps))
	for _, step := range steps {
		logStart(step.Name)
		stepCounter++

		err := step.Do(ctx)
		if err != nil {
			logFatal(step.Name, "total_execution_time", time.Since(startTime).String(), "step_executed", stepCounter, "error", err)
		}
		logDone(step.Name)
	}
	logger.Infow("All steps executed without errors.", "total_execution_time", time.Since(startTime).String(), "step_executed", stepCounter)
}
