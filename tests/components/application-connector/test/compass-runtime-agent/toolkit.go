package compass_runtime_agent

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"time"
)

type RetryableExecuteFunc func() error
type ConditionMet func() bool

type ExecuteAndWaitForCondition struct {
	retryableExecuteFunc RetryableExecuteFunc
	conditionMetFunc     ConditionMet
	tick                 time.Duration
	timeout              time.Duration
}

func (e ExecuteAndWaitForCondition) Do() error {

	err := retry.Do(func() error {
		return e.retryableExecuteFunc()
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
				res := e.conditionMetFunc()

				if res {
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
