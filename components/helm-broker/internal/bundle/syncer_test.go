package bundle_test

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle/automock"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/driver/memory"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

func TestSyncerWithOneLoader(t *testing.T) {
	urlA := "http://kyma-project.io/A"
	urlB := "http://kyma-project.io/B"

	for tn, tc := range map[string]struct {
		existing []bundle.CompleteBundle
		loadedA  []bundle.CompleteBundle
		loadedB  []bundle.CompleteBundle
		expected []bundle.CompleteBundle
	}{
		"new bundles for empty storage": {
			existing: []bundle.CompleteBundle{},
			loadedA: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlA)},
			loadedB: []bundle.CompleteBundle{
				fixBundleWithChart("id-b", "Name-B", urlB)},
			expected: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlA),
				fixBundleWithChart("id-b", "Name-B", urlB)},
		},
		"no changes": {
			existing: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlA),
				fixBundleWithChart("id-b", "Name-B", urlB)},
			loadedA: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlA)},
			loadedB: []bundle.CompleteBundle{
				fixBundleWithChart("id-b", "Name-B", urlB)},
			expected: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlA),
				fixBundleWithChart("id-b", "Name-B", urlB)},
		},
		"removal": {
			existing: []bundle.CompleteBundle{
				fixBundleWithChart("id-a0", "Name-A0", urlA),
				fixBundleWithChart("id-a1", "Name-A1", urlA),
				fixBundleWithChart("id-b", "Name-B", urlB)},
			loadedA: []bundle.CompleteBundle{
				fixBundleWithChart("id-a0", "Name-A0", urlA)},
			loadedB: []bundle.CompleteBundle{
				fixBundleWithChart("id-b", "Name-B", urlB)},
			expected: []bundle.CompleteBundle{
				fixBundleWithChart("id-a0", "Name-A0", urlA),
				fixBundleWithChart("id-b", "Name-B", urlB)},
		},
		"moved between repos": {
			existing: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlA),
				fixBundleWithChart("id-b", "Name-B", urlB)},
			loadedA: []bundle.CompleteBundle{
				fixBundleWithChart("id-b", "Name-B", urlA)},
			loadedB: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlB)},
			expected: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlB),
				fixBundleWithChart("id-b", "Name-B", urlA)},
		},
		"double ID": {
			existing: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlA),
				fixBundleWithChart("id-b", "Name-B", urlB)},
			loadedA: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-B", urlA)},
			loadedB: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlB),
				fixBundleWithChart("id-b", "Name-B", urlB)},
			expected: []bundle.CompleteBundle{
				fixBundleWithChart("id-b", "Name-B", urlB)},
		},
		"double ID with empty storage": {
			existing: []bundle.CompleteBundle{},
			loadedA: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-B", urlA)},
			loadedB: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlB),
				fixBundleWithChart("id-b", "Name-B", urlB)},
			expected: []bundle.CompleteBundle{
				fixBundleWithChart("id-b", "Name-B", urlB)},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			providerA := &automock.Provider{}
			providerA.On("ProvideBundles").Return(tc.loadedA, nil)
			defer providerA.AssertExpectations(t)

			providerB := &automock.Provider{}
			providerB.On("ProvideBundles").Return(tc.loadedB, nil)
			defer providerB.AssertExpectations(t)

			bStorage, chStorage := populatedInMemoryStorage(tc.existing)
			logSink := spy.NewLogSink()
			syncer := bundle.NewSyncer(bStorage, chStorage, logSink.Logger)
			syncer.AddProvider(urlA, providerA)
			syncer.AddProvider(urlB, providerB)

			// when
			err := syncer.Execute()

			// then
			require.NoError(t, err)

			gotBundles, err := bStorage.FindAll()
			require.NoError(t, err)

			assertBundlesAndCharts(t, tc.expected, gotBundles, chStorage)
		})
	}
}

func assertBundlesAndCharts(t *testing.T, expected []bundle.CompleteBundle, gotBundles []*internal.Bundle, chStorage storage.Chart) {
	var expectedCharts []*chart.Chart
	assert.Len(t, gotBundles, len(expected))
	for _, item := range expected {
		assert.Contains(t, gotBundles, item.Bundle)
		expectedCharts = append(expectedCharts, item.Charts...)
	}

	var gotCharts []*chart.Chart
	for _, b := range gotBundles {
		for _, plan := range b.Plans {
			gotChart, err := chStorage.Get(plan.ChartRef.Name, plan.ChartRef.Version)
			require.NoError(t, err)
			gotCharts = append(gotCharts, gotChart)
		}
	}

	assert.Len(t, gotCharts, len(expectedCharts))
	for _, ch := range expectedCharts {
		assert.Contains(t, gotCharts, ch)
	}
}

func populatedInMemoryStorage(items []bundle.CompleteBundle) (storage.Bundle, storage.Chart) {
	bStorage := memory.NewBundle()
	for _, item := range items {
		bStorage.Upsert(item.Bundle)
	}
	chStorage := memory.NewChart()
	for _, item := range items {
		for _, ch := range item.Charts {
			chStorage.Upsert(ch)
		}
	}
	return bStorage, chStorage
}

func fixBundleWithChart(id, name, url string) bundle.CompleteBundle {
	chartName := fmt.Sprintf("chart-%s", name)
	chartVersion := semver.MustParse("1.0.0")
	return bundle.CompleteBundle{
		Bundle: &internal.Bundle{
			ID:                  internal.BundleID(id),
			Name:                internal.BundleName(name),
			Description:         "simple description",
			Version:             *semver.MustParse("0.2.3"),
			RemoteRepositoryURL: url,
			Plans: map[internal.BundlePlanID]internal.BundlePlan{
				internal.BundlePlanID(fmt.Sprintf("plan-%s", name)): {
					ChartRef: internal.ChartRef{
						Name:    internal.ChartName(chartName),
						Version: *chartVersion,
					},
				},
			},
		},
		Charts: []*chart.Chart{
			{
				Metadata: &chart.Metadata{
					Name:    chartName,
					Version: chartVersion.String(),
				},
			},
		},
	}
}
