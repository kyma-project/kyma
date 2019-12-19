package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

const testTimes = 3
const informerResyncPeriod = 10 * time.Second

func TestPluggableContainer(t *testing.T) {
	appCfg := application.Config{
		Gateway: gateway.Config{
			StatusRefreshPeriod: 1 * time.Second,
		},
	}
	pluggable, err := application.New(&rest.Config{}, appCfg, informerResyncPeriod, nil)
	require.NoError(t, err)

	pluggable.SetFakeClient()

	for i := 0; i < testTimes; i++ {
		require.NotPanics(t, func() {
			err := pluggable.Enable()

			require.NoError(t, err)
			<-pluggable.Pluggable.SyncCh

			checkExportedFields(t, pluggable, true)
		})
		require.NotPanics(t, func() {
			err := pluggable.Disable()
			require.NoError(t, err)

			checkExportedFields(t, pluggable, false)
		})
	}
}

func checkExportedFields(t *testing.T, resolver *application.PluggableContainer, enabled bool) {
	assert.NotNil(t, resolver.Resolver)
	require.NotNil(t, resolver.ApplicationRetriever)
	assert.NotNil(t, resolver.ApplicationRetriever.Application())

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	val, err := resolver.Resolver.ApplicationsQuery(ctx, nil, nil, nil)
	if enabled {
		require.NoError(t, err)
	} else {
		require.Nil(t, val)
		require.Error(t, err)
	}
}
