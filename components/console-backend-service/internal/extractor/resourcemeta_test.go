package extractor

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakedisc "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	name         = "test"
	namespace    = "testspace"
	groupVersion = "apps/v1"
	resource     = "deployments"
	kind         = "Deployment"
)

var (
	fakeResources = []*v1.APIResourceList{
		&v1.APIResourceList{
			TypeMeta:     v1.TypeMeta{},
			GroupVersion: groupVersion,
			APIResources: []v1.APIResource{
				v1.APIResource{
					Name: resource,
					Kind: kind,
				},
			},
		},
	}
	resourceJSON = gqlschema.JSON{
		"apiVersion": groupVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	}
	resourceMeta = ResourceMeta{
		Name:       name,
		Namespace:  namespace,
		Kind:       kind,
		APIVersion: groupVersion,
	}
)

func TestExtractResourceMeta(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		rm, err := ExtractResourceMeta(resourceJSON)

		require.NoError(t, err)
		assert.Equal(t, resourceMeta, rm)
	})

	t.Run("Missing metadata fields", func(t *testing.T) {
		rm, err := ExtractResourceMeta(gqlschema.JSON{
			"metadata": map[string]interface{}{},
		})

		require.Error(t, err)
		assert.Empty(t, rm)
	})

	t.Run("Empty input", func(t *testing.T) {
		rm, err := ExtractResourceMeta(gqlschema.JSON{})

		require.Error(t, err)
		assert.Empty(t, rm)
	})
}

func TestGetPluralNameFromKind(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		fakeDiscovery, ok := clientset.Discovery().(*fakedisc.FakeDiscovery)
		if !ok {
			t.Fatalf("couldn't convert Discovery() to *FakeDiscovery")
		}
		fakeDiscovery.Fake.Resources = fakeResources

		pluralName, err := GetPluralNameFromKind(kind, groupVersion, fakeDiscovery)

		require.NoError(t, err)
		assert.Equal(t, resource, pluralName)
	})

	t.Run("APIGroup not found", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()

		pluralName, err := GetPluralNameFromKind(kind, groupVersion, clientset.Discovery())

		require.Error(t, err)
		assert.Empty(t, pluralName)
	})
}
