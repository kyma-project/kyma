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

// StepFunc defines contract for step function
type StepFunc func(ctx context.Context) error

// MustExecute takes steps and executes them
func (s *Scheduler) MustExecute(ctx context.Context, steps map[string]StepFunc) {
	var (
		logger = s.logger.Sugar().With("execution_id", uuid.New())

		logStart = logger.With("action", "start").Info
		logDone  = logger.With("action", "done").Info
		logFatal = logger.With("action", "failed").Fatalw

		stepCounter = 0
		startTime   = time.Now()
	)

	logger.Infof("Started execution of %d steps...", len(steps))
	for name, do := range steps {
		logStart(name)
		stepCounter++

		err := do(ctx)
		if err != nil {
			logFatal(name, "total_execution_time", time.Since(startTime).String(), "step_executed", stepCounter, "error", err)
		}
		logDone(name)
	}
	logger.Infow("All steps executed without errors.", "total_execution_time", time.Since(startTime).String(), "step_executed", stepCounter)
}
