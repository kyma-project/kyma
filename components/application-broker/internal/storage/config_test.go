package storage_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/testdata"
)

func TestConfigParse(t *testing.T) {
	// GIVEN:
	in, err := ioutil.ReadFile("testdata/ConfigAllMemory.input.yaml")
	require.NoError(t, err)

	exp := testdata.GoldenConfigMemorySingleAll()

	// WHEN:
	got, err := storage.ConfigParse(in)

	// THEN:
	assert.EqualValues(t, exp, *got)
}
