package assetstore

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/mocks"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"
)

func TestAddDocsTopic(t *testing.T) {
	t.Run("Should add DocsTopic", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock, "kyma-integration")

		docsTopicEntry := docstopic.Entry{
			Id:          "id1",
			DisplayName: "Some display name",
			Description: "Some description",
			ApiSpec: &docstopic.SpecEntry{
				Url: "www.somestorage.com/api",
				Key: "api",
			},
			EventsSpec: &docstopic.SpecEntry{
				Url: "www.somestorage.com/events",
				Key: "events",
			},
			Documentation: &docstopic.SpecEntry{
				Url: "www.somestorage.com/docs",
				Key: "documentation",
			},
			Labels: map[string]string{
				"key": "value",
			},
		}

		resourceInterfaceMock.On("Create", mock.MatchedBy(createMatcherFunction(docsTopicEntry, "kyma-integration"))).Return(&unstructured.Unstructured{}, nil)

		// when
		err := repository.Create(docsTopicEntry)

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should fail if k8s client returned error", func(t *testing.T) {

	})
}

func TestGetDocsTopic(t *testing.T) {
	t.Run("Should delete DocsTopic", func(t *testing.T) {

	})

	t.Run("Should fail if k8s client returned error", func(t *testing.T) {

	})
}

func TestUpdateDocsTopic(t *testing.T) {
	t.Run("Should update DocsTopic", func(t *testing.T) {

	})

	t.Run("Should fail if k8s client returned error", func(t *testing.T) {

	})
}

func TestDeleteDocsTopic(t *testing.T) {
	t.Run("Should delete DocsTopic", func(t *testing.T) {

	})

	t.Run("Should fail if k8s client returned error", func(t *testing.T) {

	})
}

func createMatcherFunction(docsTopicEntry docstopic.Entry, namespace string) func(*unstructured.Unstructured) bool {

	checkSpecEntry := func(entry docstopic.SpecEntry, sources map[string]v1alpha1.Source) bool {
		source, found := sources[entry.Key]
		if !found {
			return false
		}

		return source.URL == entry.Url && source.Mode == DocsTopicModeSingle
	}

	return func(u *unstructured.Unstructured) bool {
		dc := v1alpha1.DocsTopic{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &dc)
		if err != nil {
			return false
		}

		objectMetadataMatch := dc.Name == docsTopicEntry.Id && dc.Namespace == namespace

		specBasicDataMatch := dc.Spec.DisplayName == docsTopicEntry.DisplayName &&
			dc.Spec.Description == docsTopicEntry.Description

		apiSpecMatch := docsTopicEntry.ApiSpec != nil &&
			checkSpecEntry(*docsTopicEntry.ApiSpec, dc.Spec.Sources)

		eventSpecMatch := docsTopicEntry.EventsSpec != nil &&
			checkSpecEntry(*docsTopicEntry.EventsSpec, dc.Spec.Sources)

		documentationSpecMatch := docsTopicEntry.Documentation != nil &&
			checkSpecEntry(*docsTopicEntry.Documentation, dc.Spec.Sources)

		labelsMatch := reflect.DeepEqual(dc.Labels, docsTopicEntry.Labels)

		return objectMetadataMatch &&
			specBasicDataMatch &&
			apiSpecMatch &&
			eventSpecMatch &&
			documentationSpecMatch &&
			labelsMatch
	}
}
