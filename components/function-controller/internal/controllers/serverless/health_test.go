package serverless_test

import (
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestHealthChecker_Checker(t *testing.T) {
	log := zap.NewNop().Sugar()
	t.Run("Success", func(t *testing.T) {
		//GIVEN
		timeout := 10 * time.Second
		inCh := make(chan event.GenericEvent, 1)
		outCh := make(chan bool, 1)
		checker := serverless.NewHealthChecker(inCh, outCh, timeout, log)

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

	t.Run("timeout", func(t *testing.T) {
		//GIVEN
		timeout := time.Second
		inCh := make(chan event.GenericEvent, 1)
		outCh := make(chan bool, 1)
		checker := serverless.NewHealthChecker(inCh, outCh, timeout, log)

		//WHEN
		go func() {
			check := <-inCh
			require.Equal(t, check.Object.GetName(), serverless.HealthEvent)
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
		checker := serverless.NewHealthChecker(inCh, outCh, timeout, log)

		e := event.GenericEvent{
			Object: &corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					Name: "",
				},
			},
		}
		inCh <- e
		//WHEN
		err := checker.Checker(nil)

		//THEN
		require.Error(t, err)
	})

	t.Run("send a few events", func(t *testing.T) {
		//GIVEN
		timeout := 10 * time.Second
		inCh := make(chan event.GenericEvent, 1)
		outCh := make(chan bool, 1)
		checker := serverless.NewHealthChecker(inCh, outCh, timeout, log)
		var wg sync.WaitGroup

		//WHEN

		wg.Add(3)

		go func() {
			outCh <- true
			wg.Done()
		}()
		go func() {
			outCh <- true
			wg.Done()
		}()

		go func() {
			check := <-inCh
			require.Equal(t, check.Object.GetName(), serverless.HealthEvent)
			wg.Done()
		}()

		err := checker.Checker(nil)

		//THEN

		wg.Wait()

		require.NoError(t, err)
		require.Len(t, outCh, 0)
	})
}

func TestHealthName(t *testing.T) {
	//GIVEN
	//WHEN
	// This const is should be longer than 253 characters to avoid collisions with real k8s objects.
	require.Greater(t, len(serverless.HealthEvent), 253)
	//THEN
}
