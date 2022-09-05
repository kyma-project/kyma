package compass_runtime_agent

import (
	"context"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"time"
)

type RetryableExecuteFunc func() error
type ConditionMet func() error

func ExecuteAndWaitForCondition(retryableFunc RetryableExecuteFunc, conditionMetFunc ConditionMet, tick time.Duration, timeout time.Duration) error {

	err := retry.Do(func() error {
		return retryableFunc()
	})

	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(tick)

	for {
		select {
		case <-ticker.C:
			{
				if conditionMetFunc() == nil {
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
