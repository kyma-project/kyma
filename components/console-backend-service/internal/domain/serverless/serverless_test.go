package serverless

import (
	"context"

	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimes = 3
const informerResyncPeriod = 30 * time.Second

func TestPluggableContainer(t *testing.T) {
	svcFactory := fake.NewSimpleFakeServiceFactory(informerResyncPeriod)
	require.NotNil(t, svcFactory)

	pluggable, err := New(svcFactory, Config{}, nil)
	require.NoError(t, err)

	for i := 0; i < testTimes; i++ {
		require.NotPanics(t, func() {
			err := pluggable.Enable()
			require.NoError(t, err)
			<-pluggable.Pluggable.SyncCh

			checkInternalMethod(t, pluggable, true)
		})
		require.NotPanics(t, func() {
			err := pluggable.Disable()
			require.NoError(t, err)

			checkInternalMethod(t, pluggable, false)
		})
	}
}

func checkInternalMethod(t *testing.T, resolver *PluggableContainer, enabled bool) {
	assert.NotNil(t, resolver.Resolver)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	val, err := resolver.Resolver.FunctionsQuery(ctx, "default")
	if enabled {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
		require.Nil(t, val)
	}
}
