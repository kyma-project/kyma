package bundle_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle/automock"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
)

func TestUpdater_SwitchRepositories(t *testing.T) {

	// GIVEN
	for tn, tc := range map[string]struct {
		reposSize int
		url       string
	}{
		"singleURL": {url: "http://kyma-project.io/A", reposSize: 1},
		"manyURLs":  {url: "http://kyma-project.io/A;http://kyma-project.io/B", reposSize: 2},
	} {
		t.Run(tn, func(t *testing.T) {
			bRemover := &automock.BundleRemover{}
			bRemover.On("RemoveAll").Return(nil)
			defer bRemover.AssertExpectations(t)
			logSink := spy.NewLogSink()

			// WHEN
			updater := bundle.NewUpdater(bRemover, logSink.Logger)
			assert.True(t, updater.IsURLChanged(tc.url))

			// THEN
			repos, err := updater.SwitchRepositories(tc.url)
			assert.NoError(t, err)
			assert.Len(t, repos, tc.reposSize)
			assert.False(t, updater.IsURLChanged(tc.url))

		})
	}

}

func TestUpdater_SwitchRepositories_ErrorOnRemovingAllBundles(t *testing.T) {
	// GIVEN
	bRemover := &automock.BundleRemover{}
	bRemover.On("RemoveAll").Return(errors.New("fix"))
	defer bRemover.AssertExpectations(t)
	logSink := spy.NewLogSink()

	// WHEN
	updater := bundle.NewUpdater(bRemover, logSink.Logger)
	_, err := updater.SwitchRepositories("http://kyma-project.io/A")

	// THEN
	assert.Error(t, err)
}
