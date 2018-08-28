package config_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrokerFlavorGetNsFromBrokerURL(t *testing.T) {
	sut := config.NewBrokerFlavorFromConfig(&config.Config{ClusterScopedBroker: false})

	t.Run("namespace-scope broker", func(t *testing.T) {
		actNs, err := sut.GetNsFromBrokerURL("http://reb-ns-for-stage.kyma-system.svc.cluster.local/v2/catalog")
		require.NoError(t, err)
		assert.Equal(t, "stage", actNs)
	})

	t.Run("URL does not match ns-pattern", func(t *testing.T) {
		_, err := sut.GetNsFromBrokerURL("https://core-reb.kyma-system.svc.cluster.local/v2/catalog")
		assert.EqualError(t, err, "url:https://core-reb.kyma-system.svc.cluster.local/v2/catalog does not match pattern reb-ns-for-([a-z][a-z0-9-]*)\\.")
	})

}
