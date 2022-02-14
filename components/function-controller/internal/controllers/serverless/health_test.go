package serverless_test

import (
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestHealthChecker_Checker(t *testing.T) {
	logger := zap.NewNop()
	t.Run("Success", func(t *testing.T) {
		//GIVEN
		timeout := 10 * time.Second
		inCh := make(chan event.GenericEvent, 1)
		outCh := make(chan bool, 1)
		checker := serverless.NewHealthChecker(inCh, outCh, timeout, logger)

		//WHEN
		go func() {
			check := <-inCh
			require.Equal(t, check.Meta.GetName(), serverless.HealthEvent)
			outCh <- true
		}()
		err := checker.Checker(nil)

		//THEN
		require.NoError(t, err)
	})

	t.Run("timeout", func(t *testing.T) {
		//GIVEN
		timeout := time.Second
		inCh := make(chan event.GenericEvent, 1)
		outCh := make(chan bool, 1)
		checker := serverless.NewHealthChecker(inCh, outCh, timeout, logger)

		//WHEN
		go func() {
			check := <-inCh
			require.Equal(t, check.Meta.GetName(), serverless.HealthEvent)
		}()
		err := checker.Checker(nil)

		//THEN
		require.Error(t, err)
	})

	t.Run("can't send check event", func(t *testing.T) {
		//GIVEN
		timeout := time.Second
		inCh := make(chan event.GenericEvent, 1)
		outCh := make(chan bool, 1)
		checker := serverless.NewHealthChecker(inCh, outCh, timeout, logger)
		inCh <- event.GenericEvent{Meta: &ctrl.ObjectMeta{Name: ""}}
		//WHEN
		err := checker.Checker(nil)

		//THEN
		require.Error(t, err)
	})
}

func TestHealthName(t *testing.T) {
	//GIVEN
	//WHEN
	// This const is should be longer than 253 characters to avoid collisions with real k8s objects.
	require.Greater(t, len(serverless.HealthEvent), 253)
	//THEN
}
