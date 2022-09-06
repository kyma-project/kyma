package executor

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestToolkit(t *testing.T) {
	t.Run("Should return no error when verify function returns true", func(t *testing.T) {
		// given
		executeAndWait := ExecuteAndWaitForCondition{
			RetryableExecuteFunc: func() error {
				return nil
			},
			ConditionMetFunc: func() bool {
				return true
			},
			Tick:    10 * time.Second,
			Timeout: 1 * time.Minute,
		}

		// when
		err := executeAndWait.Do()

		//then
		require.NoError(t, err)
	})

	t.Run("Retry when exec function fails", func(t *testing.T) {
		// given
		counter := 1

		executeAndWait := ExecuteAndWaitForCondition{

			RetryableExecuteFunc: func() error {
				if counter < 3 {
					counter++
					return errors.New("failed")
				}

				return nil
			},
			ConditionMetFunc: func() bool {
				return true
			},
			Tick:    10 * time.Second,
			Timeout: 1 * time.Minute,
		}

		// when
		err := executeAndWait.Do()

		//then
		require.NoError(t, err)
		require.Greater(t, counter, 2)
	})

	t.Run("Return error when exec function constantly fails", func(t *testing.T) {
		// given
		executeAndWait := ExecuteAndWaitForCondition{

			RetryableExecuteFunc: func() error {
				return errors.New("call failed")
			},
			ConditionMetFunc: func() bool {
				return true
			},
			Tick:    10 * time.Second,
			Timeout: 1 * time.Minute,
		}

		// when
		err := executeAndWait.Do()

		//then
		require.Error(t, err)
	})

	t.Run("Return error when verify function constantly returns false", func(t *testing.T) {
		// given
		executeAndWait := ExecuteAndWaitForCondition{

			RetryableExecuteFunc: func() error {
				return nil
			},
			ConditionMetFunc: func() bool {
				return false
			},
			Tick:    10 * time.Second,
			Timeout: 1 * time.Minute,
		}

		// when
		err := executeAndWait.Do()

		//then
		require.Error(t, err)
	})
}
