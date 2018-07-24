package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal/platform/logger/spy"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchiveBundles(t *testing.T) {
	// GIVEN
	loaderTempDir, err := ioutil.TempDir("", "helm-broker-loader")
	require.NoError(t, err)
	defer os.RemoveAll(loaderTempDir)

	outputDir, err := ioutil.TempDir("", "helm-broker-archive-output")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// WHEN
	err = archiveBundles("testdata/input", outputDir)
	// THEN
	require.NoError(t, err)

	loader := ybundle.NewLoader(loaderTempDir, spy.NewLogDummy())

	quote, err := os.Open(filepath.Join(outputDir, "quote-1.0.1.tgz"))
	assert.NoError(t, err)

	redis, err := os.Open(filepath.Join(outputDir, "redis-0.0.3.tgz"))
	assert.NoError(t, err)

	_, _, err = loader.Load(quote)
	assert.NoError(t, err)

	_, _, err = loader.Load(redis)
	assert.NoError(t, err)
}
