package bundle_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryLoaderSuccess(t *testing.T) {
	// given
	log := spy.NewLogDummy()
	fakeRepo := &fakeRepository{path: "testdata"}

	tmpDir, err := ioutil.TempDir("../../tmp", "RepositoryLoaderTest")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	bundleLoader := bundle.NewProvider(fakeRepo, bundle.NewLoader(tmpDir, log), log)

	// when
	result, err := bundleLoader.ProvideBundles(fakeRepo.path)

	// then
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestRepositoryLoader(t *testing.T) {
	// given
	log := spy.NewLogDummy()
	fakeRepo := &fakeRepository{path: "testdata"}

	tmpDir, err := ioutil.TempDir("../../tmp", "RepositoryLoaderTest")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	bundleLoader := bundle.NewProvider(fakeRepo, bundle.NewLoader(tmpDir, log), log)

	// when
	result, err := bundleLoader.ProvideBundles(fakeRepo.path)

	// then
	require.NoError(t, err)
	assert.Len(t, result, 1)
}
