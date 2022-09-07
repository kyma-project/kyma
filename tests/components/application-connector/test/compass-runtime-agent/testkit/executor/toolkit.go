package executor

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"time"
)

type RetryableExecuteFunc func() error
type ConditionMet func() bool

type ExecuteAndWaitForCondition struct {
	RetryableExecuteFunc RetryableExecuteFunc
	ConditionMetFunc     ConditionMet
	Tick                 time.Duration
	Timeout              time.Duration
}

func (e ExecuteAndWaitForCondition) Do() error {

	err := retry.Do(func() error {
		return e.RetryableExecuteFunc()
	})

	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), e.Timeout)
	defer cancel()

	ticker := time.NewTicker(e.Tick)

	for {
		select {
		case <-ticker.C:
			{
				res := e.ConditionMetFunc()

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
