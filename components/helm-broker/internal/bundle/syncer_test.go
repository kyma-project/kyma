package bundle_test

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
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

func TestSyncerWithTwoLoaders(t *testing.T) {
	urlA := "http://kyma-project.io/A"
	urlB := "http://kyma-project.io/B"

	for tn, tc := range map[string]struct {
		existing    []bundle.CompleteBundle
		loadedA     []bundle.CompleteBundle
		loadedB     []bundle.CompleteBundle
		expected    []bundle.CompleteBundle
		idsToRemove []string
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
			idsToRemove: []string{
				"id-a1",
			},
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
		"conflicting bundles": {
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
			idsToRemove: []string{
				"id-a",
			},
		},
		"all bundles conflicting": {
			existing: []bundle.CompleteBundle{},
			loadedA: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-B", urlA), fixBundleWithChart("id-b", "Name-B", urlA)},
			loadedB: []bundle.CompleteBundle{
				fixBundleWithChart("id-a", "Name-A", urlB),
				fixBundleWithChart("id-b", "Name-B", urlB)},
			expected: []bundle.CompleteBundle{},
			idsToRemove: []string{
				"id-a", "id-b",
			},
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

			docsTopicsSvc := &automock.DocsTopicsService{}
			for _, id := range tc.idsToRemove {
				docsTopicsSvc.On("EnsureClusterDocsTopicRemoved", id).Return(nil)
			}
			defer docsTopicsSvc.AssertExpectations(t)

			bStorage, chStorage := populatedInMemoryStorage(tc.existing)
			logSink := spy.NewLogSink()
			syncer := bundle.NewSyncer(bStorage, chStorage, docsTopicsSvc, logSink.Logger)
			syncer.AddProvider(urlA, providerA)
			syncer.AddProvider(urlB, providerB)

			// when
			err := syncer.Execute()

			// then
			require.NoError(t, err)

			gotBundles, err := bStorage.FindAll(internal.ClusterWide)
			require.NoError(t, err)

			assertBundlesAndCharts(t, tc.expected, gotBundles, chStorage)
		})
	}
}

func TestSyncer_CreateDocs(t *testing.T) {
	// given
	urlA := "http://kyma-project.io/A"
	docsURL := "www.myk.com"
	bundleID := "id-a"

	fetchedBundle := fixBundleWithDocsURL(bundleID, "Name-A", urlA, docsURL)
	expectedBundle := fixBundleWithDocsURL(bundleID, "Name-A", urlA, docsURL)

	provider := &automock.Provider{}
	provider.On("ProvideBundles").Return([]bundle.CompleteBundle{fetchedBundle}, nil)
	defer provider.AssertExpectations(t)

	docsTopicsSvc := &automock.DocsTopicsService{}
	docsTopicsSvc.On("EnsureClusterDocsTopic", fetchedBundle.Bundle).Return(nil)
	defer docsTopicsSvc.AssertExpectations(t)

	bStorage, chStorage := populatedInMemoryStorage([]bundle.CompleteBundle{})
	logSink := spy.NewLogSink()
	syncer := bundle.NewSyncer(bStorage, chStorage, docsTopicsSvc, logSink.Logger)
	syncer.AddProvider(urlA, provider)

	// when
	err := syncer.Execute()

	// then
	require.NoError(t, err)

	gotBundles, err := bStorage.FindAll(internal.ClusterWide)
	require.NoError(t, err)

	assertBundlesAndCharts(t, []bundle.CompleteBundle{expectedBundle}, gotBundles, chStorage)
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
			gotChart, err := chStorage.Get(internal.ClusterWide, plan.ChartRef.Name, plan.ChartRef.Version)
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
		bStorage.Upsert(internal.ClusterWide, item.Bundle)
	}
	chStorage := memory.NewChart()
	for _, item := range items {
		for _, ch := range item.Charts {
			chStorage.Upsert(internal.ClusterWide, ch)
		}
	}
	return bStorage, chStorage
}

func fixBundleWithChart(id, name, url string) bundle.CompleteBundle {
	chartName := fmt.Sprintf("chart-%s", name)
	chartVersion := semver.MustParse("1.0.0")
	return bundle.CompleteBundle{
		Bundle: &internal.Bundle{
			ID:            internal.BundleID(id),
			Name:          internal.BundleName(name),
			Description:   "simple description",
			Version:       *semver.MustParse("0.2.3"),
			RepositoryURL: url,
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

func fixBundleWithDocsURL(id, name, url, docsURL string) bundle.CompleteBundle {
	chartName := fmt.Sprintf("chart-%s", name)
	chartVersion := semver.MustParse("1.0.0")
	return bundle.CompleteBundle{
		Bundle: &internal.Bundle{
			ID:            internal.BundleID(id),
			Name:          internal.BundleName(name),
			Description:   "simple description",
			Version:       *semver.MustParse("0.2.3"),
			RepositoryURL: url,
			Plans: map[internal.BundlePlanID]internal.BundlePlan{
				internal.BundlePlanID(fmt.Sprintf("plan-%s", name)): {
					ChartRef: internal.ChartRef{
						Name:    internal.ChartName(chartName),
						Version: *chartVersion,
					},
				},
			},
			Docs: []internal.BundleDocs{
				{
					Template: v1alpha1.ClusterDocsTopicSpec{
						CommonDocsTopicSpec: v1alpha1.CommonDocsTopicSpec{
							Sources: []v1alpha1.Source{
								{
									URL: docsURL,
								},
							},
						},
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

func fixBundleWithEmptyDocs(id, name, url string) bundle.CompleteBundle {
	chartName := fmt.Sprintf("chart-%s", name)
	chartVersion := semver.MustParse("1.0.0")
	return bundle.CompleteBundle{
		Bundle: &internal.Bundle{
			ID:            internal.BundleID(id),
			Name:          internal.BundleName(name),
			Description:   "simple description",
			Version:       *semver.MustParse("0.2.3"),
			RepositoryURL: url,
			Plans: map[internal.BundlePlanID]internal.BundlePlan{
				internal.BundlePlanID(fmt.Sprintf("plan-%s", name)): {
					ChartRef: internal.ChartRef{
						Name:    internal.ChartName(chartName),
						Version: *chartVersion,
					},
				},
			},
			Docs: []internal.BundleDocs{
				{
					Template: v1alpha1.ClusterDocsTopicSpec{
						CommonDocsTopicSpec: v1alpha1.CommonDocsTopicSpec{
							Sources: []v1alpha1.Source{
								{},
							},
						},
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
