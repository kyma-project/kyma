package bundle_test

import (
	"io/ioutil"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle/automock"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryLoaderSuccess(t *testing.T) {
	// given
	log := spy.NewLogDummy()
	tmpDir, err := ioutil.TempDir("../../tmp", "RepositoryLoaderTest")
	require.NoError(t, err)
	bundleLoader := bundle.NewProvider(bundle.NewLocalRepository("testdata"), bundle.NewLoader(tmpDir, log), log)

	// when
	result, err := bundleLoader.ProvideBundles()

	// then
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestRepositoryLoader(t *testing.T) {
	// given
	log := spy.NewLogDummy()
	repo := &automock.Repository{}
	repo.On("IndexReader").Return("")

	tmpDir, err := ioutil.TempDir("../../tmp", "RepositoryLoaderTest")
	require.NoError(t, err)
	bundleLoader := bundle.NewProvider(bundle.NewLocalRepository("testdata"), bundle.NewLoader(tmpDir, log), log)

	// when
	result, err := bundleLoader.ProvideBundles()

	// then
	require.NoError(t, err)
	assert.Len(t, result, 1)
}
