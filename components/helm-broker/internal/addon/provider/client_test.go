package provider_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon/provider"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryClientSuccess(t *testing.T) {
	// given
	log := spy.NewLogDummy()
	fakeRepo := &fakeRepository{path: "../testdata"}

	tmpDir, err := ioutil.TempDir("../../../tmp", "RepositoryLoaderTest")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	bundleLoader, err := provider.NewClient(fakeRepo, addon.NewLoader(tmpDir, log), log)
	require.NoError(t, err)

	bundleEntry := addon.EntryDTO{
		Name:    "redis",
		Version: "0.0.1",
	}

	// when
	gotIdx, gotIdxErr := bundleLoader.GetIndex()
	gotBundle, gotBundleErr := bundleLoader.GetCompleteAddon(bundleEntry)

	// then
	require.NoError(t, gotIdxErr)
	assert.NotEmpty(t, gotIdx)

	require.NoError(t, gotBundleErr)
	assert.NotEmpty(t, gotBundle)
}

// fakeRepository provide access to bundles repository
type fakeRepository struct {
	path string
}

// IndexReader returns index.yaml file from fake repository
func (p *fakeRepository) IndexReader() (io.ReadCloser, error) {
	fName := fmt.Sprintf("%s/%s", p.path, "index.yaml")
	return os.Open(fName)
}

// BundleLoadInfo returns info how to load bundle
func (p *fakeRepository) BundleLoadInfo(name addon.Name, version addon.Version) (provider.LoadType, string, error) {
	return provider.ArchiveLoadType, p.BundleDocURL(name, version), nil
}

// BundleDocURL returns download url for given bundle
func (p *fakeRepository) BundleDocURL(name addon.Name, version addon.Version) string {
	return fmt.Sprintf("%s/%s-%s.tgz", p.path, name, version)
}

// Cleanup added to fulfil the interface expectation
func (p *fakeRepository) Cleanup() error {
	return nil
}
