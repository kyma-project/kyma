package application_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/application/gateway"
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

			checkExportedFields(t, pluggable)
		})
		require.NotPanics(t, func() {
			err := pluggable.Disable()
			require.NoError(t, err)

			checkExportedFields(t, pluggable)
		})
	}
}

func checkExportedFields(t *testing.T, resolver *application.PluggableContainer) {
	assert.NotNil(t, resolver.Resolver)
	assert.NotNil(t, resolver.AppLister)
}
