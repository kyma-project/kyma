package storage_test

import (
	"github.com/kyma-project/kyma/components/application-connector/application-broker/internal/storage"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-connector/application-broker/internal/storage/testdata"
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
