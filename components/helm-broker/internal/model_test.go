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
