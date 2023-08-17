package v1alpha2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackend_CopyHashes(t *testing.T) {
	// given
	b := Backend{}

	// then
	require.Equal(t, int64(0), b.Ev2hash)
	require.Equal(t, int64(0), b.EventMeshHash)
	require.Equal(t, int64(0), b.WebhookAuthHash)
	require.Equal(t, int64(0), b.EventMeshLocalHash)

	// given
	src := Backend{
		Ev2hash:            int64(1118518533334734626),
		EventMeshHash:      int64(1748405436686967274),
		WebhookAuthHash:    int64(1118518533334734627),
		EventMeshLocalHash: int64(1883494500014499539),
	}

	// when
	b.CopyHashes(src)

	// then
	require.Equal(t, src.Ev2hash, b.Ev2hash)
	require.Equal(t, src.EventMeshHash, b.EventMeshHash)
	require.Equal(t, src.WebhookAuthHash, b.WebhookAuthHash)
	require.Equal(t, src.EventMeshLocalHash, b.EventMeshLocalHash)
}
