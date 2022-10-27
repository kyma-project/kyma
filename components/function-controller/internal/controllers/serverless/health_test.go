package serverless_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHealthChecker_Checker(t *testing.T) {
	log := zap.NewNop().Sugar()
	t.Run("Success", func(t *testing.T) {
		//GIVEN
		timeout := 10 * time.Second
		checker, inCh, outCh := serverless.NewHealthChecker(timeout, log)

		//WHEN
		go func() {
			check := <-inCh
			require.Equal(t, check.Object.GetName(), serverless.HealthEvent)
			outCh <- true
		}()
		err := checker.Checker(nil)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Timeout", func(t *testing.T) {
		//GIVEN
		timeout := time.Second
		checker, inCh, _ := serverless.NewHealthChecker(timeout, log)

		//WHEN
		go func() {
			check := <-inCh
			require.Equal(t, check.Object.GetName(), serverless.HealthEvent)
		}()
		err := checker.Checker(nil)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "reconcile didn't send confirmation")

	})

	t.Run("Can't send check event", func(t *testing.T) {
		//GIVEN
		timeout := time.Second
		checker, _, _ := serverless.NewHealthChecker(timeout, log)

		//WHEN
		err := checker.Checker(nil)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout when sending check event")
	})
}

func TestHealthName(t *testing.T) {
	//GIVEN
	//WHEN
	// This const is longer than 253 characters to avoid collisions with real k8s objects.
	require.Greater(t, len(serverless.HealthEvent), 253)
	//THEN
}
