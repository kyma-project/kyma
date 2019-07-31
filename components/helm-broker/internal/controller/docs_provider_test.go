package controller

import (
	"context"
	"testing"

	"fmt"

	"github.com/Masterminds/semver"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDocsProvider_EnsureClusterDocsTopic(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	const id = "123"

	for tn, tc := range map[string]struct {
		givenAddon addon.CompleteAddon
	}{
		"URL set":   {fixAddonWithDocsURL(id, "test", "url", "url2")},
		"empty URL": {fixAddonWithEmptyDocs(id, "test", "url")},
	} {
		t.Run(tn, func(t *testing.T) {
			c := fake.NewFakeClient()
			cdt := fixClusterDocsTopic(id)
			docsProvider := NewDocsProvider(c)

			// when
			err = docsProvider.EnsureClusterDocsTopic(tc.givenAddon.Addon)
			require.NoError(t, err)

			// then
			err = c.Get(context.Background(), client.ObjectKey{Namespace: cdt.Namespace, Name: cdt.Name}, cdt)
			require.NoError(t, err)
			assert.Equal(t, tc.givenAddon.Addon.Docs[0].Template, cdt.Spec.CommonDocsTopicSpec)
		})
	}
}

func TestDocsProvider_EnsureClusterDocsTopic_UpdateIfExist(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	const id = "123"
	cdt := fixClusterDocsTopic(id)
	addonWithEmptyDocsURL := fixAddonWithEmptyDocs(id, "test", "url")
	addonWithEmptyDocsURL.Addon.Docs[0].Template.Description = "new description"

	c := fake.NewFakeClient(cdt)
	docsProvider := NewDocsProvider(c)

	// when
	err = docsProvider.EnsureClusterDocsTopic(addonWithEmptyDocsURL.Addon)
	require.NoError(t, err)

	// then
	err = c.Get(context.Background(), client.ObjectKey{Namespace: cdt.Namespace, Name: cdt.Name}, cdt)
	require.NoError(t, err)
	assert.Equal(t, addonWithEmptyDocsURL.Addon.Docs[0].Template, cdt.Spec.CommonDocsTopicSpec)
}

func TestDocsProvider_EnsureClusterDocsTopicRemoved(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	const id = "123"
	cdt := fixClusterDocsTopic(id)
	c := fake.NewFakeClient(cdt)
	docsProvider := NewDocsProvider(c)

	// when
	err = docsProvider.EnsureClusterDocsTopicRemoved(id)
	require.NoError(t, err)

	// then
	err = c.Get(context.Background(), client.ObjectKey{Namespace: cdt.Namespace, Name: cdt.Name}, cdt)
	assert.True(t, errors.IsNotFound(err))
}

func TestDocsProvider_EnsureClusterDocsTopicRemoved_NotExists(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	const id = "123"
	cdt := fixClusterDocsTopic(id)
	c := fake.NewFakeClient()
	docsProvider := NewDocsProvider(c)

	// when
	err = docsProvider.EnsureClusterDocsTopicRemoved(id)
	require.NoError(t, err)

	// then
	err = c.Get(context.Background(), client.ObjectKey{Namespace: cdt.Namespace, Name: cdt.Name}, cdt)
	assert.True(t, errors.IsNotFound(err))
}

func TestDocsProvider_EnsureDocsTopic(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	dt := fixDocsTopic()

	for tn, tc := range map[string]struct {
		givenAddon addon.CompleteAddon
	}{
		"URL set":   {fixAddonWithDocsURL(dt.Name, "test", "url", "url2")},
		"empty URL": {fixAddonWithEmptyDocs(dt.Name, "test", "url")},
	} {
		t.Run(tn, func(t *testing.T) {
			c := fake.NewFakeClient(dt)
			docsProvider := NewDocsProvider(c)

			// when
			err = docsProvider.EnsureDocsTopic(tc.givenAddon.Addon, dt.Namespace)
			require.NoError(t, err)

			// then
			result := v1alpha1.DocsTopic{}
			err = c.Get(context.Background(), client.ObjectKey{Namespace: dt.Namespace, Name: dt.Name}, &result)
			require.NoError(t, err)
			assert.Equal(t, tc.givenAddon.Addon.Docs[0].Template, result.Spec.CommonDocsTopicSpec)
		})
	}
}

func TestDocsProvider_EnsureDocsTopic_UpdateIfExist(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	dt := fixDocsTopic()

	addonWithEmptyDocsURL := fixAddonWithEmptyDocs(dt.Name, "test", "url")
	addonWithEmptyDocsURL.Addon.Docs[0].Template.Description = "new description"

	c := fake.NewFakeClient(dt)
	docsProvider := NewDocsProvider(c)

	// when
	err = docsProvider.EnsureDocsTopic(addonWithEmptyDocsURL.Addon, dt.Namespace)
	require.NoError(t, err)

	// then
	result := v1alpha1.DocsTopic{}
	err = c.Get(context.Background(), client.ObjectKey{Namespace: dt.Namespace, Name: dt.Name}, &result)
	require.NoError(t, err)
	assert.Equal(t, addonWithEmptyDocsURL.Addon.Docs[0].Template, result.Spec.CommonDocsTopicSpec)
}

func TestDocsProvider_EnsureDocsTopicRemoved(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	dt := fixDocsTopic()
	c := fake.NewFakeClient(dt)
	docsProvider := NewDocsProvider(c)

	// when
	err = docsProvider.EnsureDocsTopicRemoved(dt.Name, dt.Namespace)
	require.NoError(t, err)

	// then
	result := v1alpha1.DocsTopic{}
	err = c.Get(context.Background(), client.ObjectKey{Namespace: dt.Namespace, Name: dt.Name}, &result)
	assert.True(t, errors.IsNotFound(err))
}

func TestDocsProvider_EnsureDocsTopicRemoved_NotExists(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	dt := fixDocsTopic()
	c := fake.NewFakeClient()
	docsProvider := NewDocsProvider(c)

	// when
	err = docsProvider.EnsureDocsTopicRemoved(dt.Name, dt.Namespace)
	require.NoError(t, err)

	// then
	result := v1alpha1.DocsTopic{}
	err = c.Get(context.Background(), client.ObjectKey{Namespace: dt.Namespace, Name: dt.Name}, &result)
	assert.True(t, errors.IsNotFound(err))
}

func fixClusterDocsTopic(id string) *v1alpha1.ClusterDocsTopic {
	return &v1alpha1.ClusterDocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name: id,
		},
	}
}

func fixDocsTopic() *v1alpha1.ClusterDocsTopic {
	return &v1alpha1.ClusterDocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}
}

func fixAddonWithDocsURL(id, name, url, docsURL string) addon.CompleteAddon {
	chartName := fmt.Sprintf("chart-%s", name)
	chartVersion := semver.MustParse("1.0.0")
	return addon.CompleteAddon{
		Addon: &internal.Addon{
			ID:            internal.AddonID(id),
			Name:          internal.AddonName(name),
			Description:   "simple description",
			Version:       *semver.MustParse("0.0.1"),
			RepositoryURL: url,
			Plans: map[internal.AddonPlanID]internal.AddonPlan{
				internal.AddonPlanID(fmt.Sprintf("plan-%s", name)): {
					ChartRef: internal.ChartRef{
						Name:    internal.ChartName(chartName),
						Version: *chartVersion,
					},
				},
			},
			Docs: []internal.AddonDocs{
				{
					Template: v1alpha1.CommonDocsTopicSpec{
						Sources: []v1alpha1.Source{
							{
								URL: docsURL,
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

func fixAddonWithEmptyDocs(id, name, url string) addon.CompleteAddon {
	chartName := fmt.Sprintf("chart-%s", name)
	chartVersion := semver.MustParse("1.0.0")
	return addon.CompleteAddon{
		Addon: &internal.Addon{
			ID:            internal.AddonID(id),
			Name:          internal.AddonName(name),
			Description:   "simple description",
			Version:       *semver.MustParse("0.0.1"),
			RepositoryURL: url,
			Plans: map[internal.AddonPlanID]internal.AddonPlan{
				internal.AddonPlanID(fmt.Sprintf("plan-%s", name)): {
					ChartRef: internal.ChartRef{
						Name:    internal.ChartName(chartName),
						Version: *chartVersion,
					},
				},
			},
			Docs: []internal.AddonDocs{
				{
					Template: v1alpha1.CommonDocsTopicSpec{
						Sources: []v1alpha1.Source{
							{},
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
