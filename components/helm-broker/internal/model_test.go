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
	namespace := internal.Namespace("test-addon-namespace")
	collection := []*internal.Instance{
		{ServiceID: "a1", Namespace: "test-addon-namespace"},
		{ServiceID: "a2", Namespace: "test-addon-namespace"},
		{ServiceID: "a3", Namespace: "test-addon-namespace"},
		{ServiceID: "a2", Namespace: "other-addon-namespace"},
	}

	addonExist := internal.Addon{
		Metadata: internal.AddonMetadata{
			ProvisionOnlyOnce: true,
		},
		ID: "a1",
	}
	addonNotExist := internal.Addon{
		Metadata: internal.AddonMetadata{
			ProvisionOnlyOnce: true,
		},
		ID: "a5",
	}
	addonManyProvision := internal.Addon{
		Metadata: internal.AddonMetadata{
			ProvisionOnlyOnce: false,
		},
		ID: "a1",
	}

	// WHEN/THEN
	assert.False(t, addonExist.IsProvisioningAllowed(namespace, collection))
	assert.True(t, addonExist.IsProvisioningAllowed("other-addon-namespace", collection))
	assert.True(t, addonExist.IsProvisioningAllowed("other-ns", collection))
	assert.True(t, addonNotExist.IsProvisioningAllowed(namespace, collection))
	assert.True(t, addonManyProvision.IsProvisioningAllowed(namespace, collection))
}
