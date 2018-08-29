package mode_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/config"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/mode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrokerModeGetNsFromURL(t *testing.T) {
	// GIVEN
	sut, err := mode.NewBrokerService(&config.Config{ClusterScopedBrokerEnabled: false})
	require.NoError(t, err)
	// WHEN
	actNs, err := sut.GetNsFromBrokerURL("http://reb-ns-for-stage.kyma-system.svc.cluster.local/v2/catalog")
	// THEN
	require.NoError(t, err)
	assert.Equal(t, "stage", actNs)

}

func TestBrokerModeErrorOnGetNsFromURL(t *testing.T) {
	// GIVEN
	sut, err := mode.NewBrokerService(&config.Config{ClusterScopedBrokerEnabled: false})
	require.NoError(t, err)
	// WHEN
	_, err = sut.GetNsFromBrokerURL("https://core-reb.kyma-system.svc.cluster.local/v2/catalog")
	// THEN
	assert.EqualError(t, err, "url:https://core-reb.kyma-system.svc.cluster.local/v2/catalog does not match pattern reb-ns-for-([a-z][a-z0-9-]*)\\.")
}
