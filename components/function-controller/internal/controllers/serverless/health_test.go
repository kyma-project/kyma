package serverless_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestHealthChecker_Checker(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		//GIVEN
		timeout := 10 * time.Second
		inCh := make(chan event.GenericEvent, 1)
		outCh := make(chan bool, 1)
		checker := serverless.NewChecker(inCh, outCh, timeout)

		//WHEN
		go func() {
			check := <-inCh
			require.Equal(t, check.Meta.GetName(), serverless.HEALTH_EVENT)
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
		checker := serverless.NewChecker(inCh, outCh, timeout)

		//WHEN
		go func() {
			check := <-inCh
			require.Equal(t, check.Meta.GetName(), serverless.HEALTH_EVENT)
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
		checker := serverless.NewChecker(inCh, outCh, timeout)
		inCh <- event.GenericEvent{Meta: &ctrl.ObjectMeta{Name: ""}}
		//WHEN
		err := checker.Checker(nil)

		//THEN
		require.Error(t, err)
	})
}

func TestHeathlName(t *testing.T) {
	//GIVEN
	//WHEN
	require.Greater(t, len(serverless.HEALTH_EVENT), 64)
	//THEN
}
