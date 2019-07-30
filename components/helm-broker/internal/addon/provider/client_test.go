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

	addonLoader, err := provider.NewClient(fakeRepo, addon.NewLoader(tmpDir, log), log)
	require.NoError(t, err)

	entry := addon.EntryDTO{
		Name:    "redis",
		Version: "0.0.1",
	}

	// when
	gotIdx, gotIdxErr := addonLoader.GetIndex()
	gotAddon, gotAddonErr := addonLoader.GetCompleteAddon(entry)

	// then
	require.NoError(t, gotIdxErr)
	assert.NotEmpty(t, gotIdx)

	require.NoError(t, gotAddonErr)
	assert.NotEmpty(t, gotAddon)
}

// fakeRepository provide access to addons repository
type fakeRepository struct {
	path string
}

// IndexReader returns index.yaml file from fake repository
func (p *fakeRepository) IndexReader() (io.ReadCloser, error) {
	fName := fmt.Sprintf("%s/%s", p.path, "index.yaml")
	return os.Open(fName)
}

// AddonLoadInfo returns info how to load addon
func (p *fakeRepository) AddonLoadInfo(name addon.Name, version addon.Version) (provider.LoadType, string, error) {
	return provider.ArchiveLoadType, p.AddonDocURL(name, version), nil
}

// AddonDocURL returns download url for given addon
func (p *fakeRepository) AddonDocURL(name addon.Name, version addon.Version) string {
	return fmt.Sprintf("%s/%s-%s.tgz", p.path, name, version)
}

// Cleanup added to fulfil the interface expectation
func (p *fakeRepository) Cleanup() error {
	return nil
}
