package internal_test

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

func TestChartRefGobEncodeDecode(t *testing.T) {
	for sym, exp := range map[string]internal.ChartRef{
		"A":          {Name: "NameA", Version: *semver.MustParse("0.0.1")},
		"empty/name": {Name: "NameA"},
		"empty/all":  {},
	} {
		t.Run(sym, func(t *testing.T) {
			// GIVEN:
			buf := bytes.Buffer{}
			enc := gob.NewEncoder(&buf)
			dec := gob.NewDecoder(&buf)
			var got internal.ChartRef

			// WHEN:
			err := enc.Encode(&exp)
			require.NoError(t, err)

			err = dec.Decode(&got)
			require.NoError(t, err)

			// THEN:
			assert.Equal(t, exp.Name, got.Name)
			assert.Equal(t, exp.Version.String(), got.Version.String())
		})
	}
}

func TestCanBeProvision(t *testing.T) {
	// Given
	namespace := internal.Namespace("test-bundle-namespace")
	collection := []*internal.Instance{
		{ServiceID: "a1", Namespace: "test-bundle-namespace"},
		{ServiceID: "a2", Namespace: "test-bundle-namespace"},
		{ServiceID: "a3", Namespace: "test-bundle-namespace"},
		{ServiceID: "a2", Namespace: "other-bundle-namespace"},
	}

	bundleExist := internal.Bundle{
		Metadata: internal.BundleMetadata{
			ProvisionOnlyOnce: true,
		},
		ID: "a1",
	}
	bundleNotExist := internal.Bundle{
		Metadata: internal.BundleMetadata{
			ProvisionOnlyOnce: true,
		},
		ID: "a5",
	}
	bundleManyProvision := internal.Bundle{
		Metadata: internal.BundleMetadata{
			ProvisionOnlyOnce: false,
		},
		ID: "a1",
	}

	// WHEN/THEN
	assert.False(t, bundleExist.IsProvisioningAllowed(namespace, collection))
	assert.True(t, bundleExist.IsProvisioningAllowed("other-bundle-namespace", collection))
	assert.True(t, bundleExist.IsProvisioningAllowed("other-ns", collection))
	assert.True(t, bundleNotExist.IsProvisioningAllowed(namespace, collection))
	assert.True(t, bundleManyProvision.IsProvisioningAllowed(namespace, collection))
}
