package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrokerModeGetNsFromURL(t *testing.T) {
	// GIVEN
	sut, err := NewNsBrokerService()
	require.NoError(t, err)
	// WHEN
	actNs, err := sut.GetNsFromBrokerURL("http://ab-ns-for-stage.kyma-system.svc.cluster.local/v2/catalog")
	// THEN
	require.NoError(t, err)
	assert.Equal(t, "stage", actNs)
}

func TestBrokerModeErrorOnGetNsFromURL(t *testing.T) {
	// GIVEN
	sut, err := NewNsBrokerService()
	require.NoError(t, err)
	// WHEN
	_, err = sut.GetNsFromBrokerURL("https://core-ab.kyma-system.svc.cluster.local/v2/catalog")
	// THEN
	assert.EqualError(t, err, "url:https://core-ab.kyma-system.svc.cluster.local/v2/catalog does not match pattern ab-ns-for-([a-z][a-z0-9-]*)\\.")
}
