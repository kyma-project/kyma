package bundle_test

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDocsProvider_EnsureClusterDocsTopic(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)
	const id = "123"

	for tn, tc := range map[string]struct {
		givenBundle bundle.CompleteBundle
	}{
		"URL set":   {fixBundleWithDocsURL(id, "test", "url", "url2")},
		"empty URL": {fixBundleWithEmptyDocs(id, "test", "url")},
	} {
		t.Run(tn, func(t *testing.T) {
			c := fake.NewFakeClient()
			cdt := fixClusterDocsTopic(id)
			docsProvider := bundle.NewDocsProvider(c)

			// when
			err = docsProvider.EnsureClusterDocsTopic(tc.givenBundle.Bundle)
			require.NoError(t, err)

			// then
			err = c.Get(context.Background(), client.ObjectKey{Namespace: cdt.Namespace, Name: cdt.Name}, cdt)
			require.NoError(t, err)
			assert.Equal(t, tc.givenBundle.Bundle.Docs[0].Template, cdt.Spec)
		})
	}
}

func TestDocsProvider_EnsureClusterDocsTopic_UpdateIfExist(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	const id = "123"
	cdt := fixClusterDocsTopic(id)
	bundleWithEmptyDocsURL := fixBundleWithEmptyDocs(id, "test", "url")
	bundleWithEmptyDocsURL.Bundle.Docs[0].Template.Description = "new description"

	c := fake.NewFakeClient(cdt)
	docsProvider := bundle.NewDocsProvider(c)

	// when
	err = docsProvider.EnsureClusterDocsTopic(bundleWithEmptyDocsURL.Bundle)
	require.NoError(t, err)

	// then
	err = c.Get(context.Background(), client.ObjectKey{Namespace: cdt.Namespace, Name: cdt.Name}, cdt)
	require.NoError(t, err)
	assert.Equal(t, bundleWithEmptyDocsURL.Bundle.Docs[0].Template, cdt.Spec)
}

func TestDocsProvider_EnsureClusterDocsTopicRemoved(t *testing.T) {
	// given
	err := v1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	const id = "123"
	cdt := fixClusterDocsTopic(id)
	c := fake.NewFakeClient(cdt)
	docsProvider := bundle.NewDocsProvider(c)

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
	docsProvider := bundle.NewDocsProvider(c)

	// when
	err = docsProvider.EnsureClusterDocsTopicRemoved(id)
	require.NoError(t, err)

	// then
	err = c.Get(context.Background(), client.ObjectKey{Namespace: cdt.Namespace, Name: cdt.Name}, cdt)
	assert.True(t, errors.IsNotFound(err))
}

func fixClusterDocsTopic(id string) *v1alpha1.ClusterDocsTopic {
	return &v1alpha1.ClusterDocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name: id,
		},
	}
}
