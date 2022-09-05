package compass_runtime_agent

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"time"
)

type RetryableExecuteFunc func() error
type ConditionMet func() error

type ExecuteAndWaitForCondition struct {
	retryableFunc    RetryableExecuteFunc
	conditionMetFunc ConditionMet
	tick             time.Duration
	timeout          time.Duration
}

func (e ExecuteAndWaitForCondition) Do() error {

	err := retry.Do(func() error {
		return e.retryableFunc()
	})

	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	ticker := time.NewTicker(e.tick)

	for {
		select {
		case <-ticker.C:
			{
				if e.conditionMetFunc() == nil {
					ticker.Stop()
					return nil
				}
			}
		case <-ctx.Done():
			{
				ticker.Stop()
				return errors.New("Condition not met")
			}
		}
	}
}
