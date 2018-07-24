package main

import (
	"bytes"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

func TestYAMLRender_Render(t *testing.T) {
	// GIVEN
	fix := []*internal.Bundle{
		{Name: "A", Description: "Desc A 1", Version: *semver.MustParse("1.2.3")},
		{Name: "A", Description: "Desc A 2", Version: *semver.MustParse("1.2.4")},
		{Name: "B", Description: "Desc B", Version: *semver.MustParse("4.5.6")},
	}

	// YAML imposes order on serialisation which in our implementation uses asc ordering by key as string.
	// It allows us to do simple text comparision on result even with maps.
	exp := `apiVersion: v1
entries:
  A:
  - description: Desc A 1
    name: A
    version: 1.2.3
  - description: Desc A 2
    name: A
    version: 1.2.4
  B:
  - description: Desc B
    name: B
    version: 4.5.6
`

	buf := &bytes.Buffer{}

	// WHEN
	require.NoError(t, render(fix, buf))

	// THEN
	assert.Equal(t, exp, string(buf.String()))
}
