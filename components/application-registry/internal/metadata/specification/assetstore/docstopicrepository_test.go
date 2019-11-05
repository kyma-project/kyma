package assetstore

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/mocks"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestUpsertDocsTopic(t *testing.T) {
	t.Run("Should create DocsTopic", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)

		docsTopicEntry := createTestDocsTopicEntry()

		resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{}, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		resourceInterfaceMock.On("Create", mock.MatchedBy(createMatcherFunction(docsTopicEntry, "")), metav1.CreateOptions{}).
			Return(&unstructured.Unstructured{}, nil)

		// when
		err := repository.Upsert(docsTopicEntry)

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should update DocsTopic", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)

		dt := createK8sDocsTopic()

		object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&dt)
		require.NoError(t, err)

		resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		docsTopicEntry := createTestDocsTopicEntry()
		resourceInterfaceMock.On("Update", mock.MatchedBy(createMatcherFunction(docsTopicEntry, "1")), metav1.UpdateOptions{}).Return(&unstructured.Unstructured{}, nil)

		// when
		err = repository.Upsert(docsTopicEntry)

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Create", 0)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("Should fail if k8s client returned error on Get", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Get", mock.Anything, metav1.GetOptions{}).
			Return(&unstructured.Unstructured{}, errors.New("some error"))

		// when
		err := repository.Upsert(createTestDocsTopicEntry())

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should fail if k8s client returned error on Create", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{}, k8serrors.NewNotFound(schema.GroupResource{}, ""))
		resourceInterfaceMock.On("Create", mock.Anything, metav1.CreateOptions{}).
			Return(&unstructured.Unstructured{}, errors.New("some error"))

		// when
		err := repository.Upsert(createTestDocsTopicEntry())

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should fail if k8s client returned error on Update", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)

		dt := createK8sDocsTopic()
		object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&dt)
		require.NoError(t, err)

		resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		resourceInterfaceMock.On("Update", mock.Anything, metav1.UpdateOptions{}).Return(&unstructured.Unstructured{}, errors.New("some error"))

		// when
		err = repository.Upsert(createTestDocsTopicEntry())

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Get", 2)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Create", 0)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Update", 1)
	})
}

func TestGetDocsTopic(t *testing.T) {
	t.Run("Should get DocsTopic", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)
		{
			dt := createK8sDocsTopic()

			object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&dt)
			require.NoError(t, err)

			resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
				Return(&unstructured.Unstructured{Object: object}, nil)
		}

		// when
		docsTopic, err := repository.Get("id1")
		require.NoError(t, err)

		// then
		assert.Equal(t, "Some display name", docsTopic.DisplayName)
		assert.Equal(t, "Some description", docsTopic.Description)
		assert.Equal(t, "id1", docsTopic.Id)
		assert.Equal(t, len(docsTopic.Urls), 1)
		assert.Equal(t, docsTopic.Urls["api"], "www.somestorage.com/api")
		assert.Equal(t, len(docsTopic.Labels), 1)
		assert.Equal(t, "value", docsTopic.Labels["key"])
	})

	t.Run("Should fail with Not Found if DocsTopic doesn't exist", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Get", mock.Anything, metav1.GetOptions{}).
			Return(&unstructured.Unstructured{}, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		// when
		_, err := repository.Get("id1")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func createK8sDocsTopic() v1alpha1.ClusterDocsTopic {
	return v1alpha1.ClusterDocsTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "id1",
			Namespace: "kyma-integration",
			Labels: map[string]string{
				"key": "value",
			},
			ResourceVersion: "1",
		},
		Spec: v1alpha1.ClusterDocsTopicSpec{
			CommonDocsTopicSpec: v1alpha1.CommonDocsTopicSpec{
				DisplayName: "Some display name",
				Description: "Some description",
				Sources: []v1alpha1.Source{
					{
						URL:  "www.somestorage.com/api",
						Mode: v1alpha1.DocsTopicSingle,
						Type: "api",
					},
				},
			},
		}}
}

func TestDeleteDocsTopic(t *testing.T) {
	t.Run("Should delete DocsTopic", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Delete", "id1", &metav1.DeleteOptions{}).Return(nil)

		// when
		err := repository.Delete("id1")

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should fail if k8s client returned error", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Delete", "id1", &metav1.DeleteOptions{}).Return(errors.New("some error"))

		// when
		err := repository.Delete("id1")

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should not fail if Docs Topic doesn't exist", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewDocsTopicRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Delete", "id1", &metav1.DeleteOptions{}).Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))

		// when
		err := repository.Delete("id1")

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})
}

func createTestDocsTopicEntry() docstopic.Entry {
	return docstopic.Entry{
		Id:          "id1",
		DisplayName: "Some display name",
		Description: "Some description",
		Urls: map[string]string{
			docstopic.KeyOpenApiSpec: "www.somestorage.com/api",
		},
		Labels: map[string]string{
			"key": "value",
		},
	}
}

func createMatcherFunction(docsTopicEntry docstopic.Entry, expectedResourceVersion string) func(*unstructured.Unstructured) bool {
	findSource := func(sources []v1alpha1.Source, key string) (v1alpha1.Source, bool) {
		for _, source := range sources {
			if source.Type == key && source.Name == fmt.Sprintf(DocsTopicNameFormat, key, docsTopicEntry.Id) {
				return source, true
			}
		}

		return v1alpha1.Source{}, false
	}

	checkUrls := func(urls map[string]string, sources []v1alpha1.Source) bool {
		if len(urls) != len(sources) {
			return false
		}

		for key, value := range urls {
			source, found := findSource(sources, key)
			if !found || value != source.URL {
				return false
			}
		}

		return true
	}

	return func(u *unstructured.Unstructured) bool {
		dt := v1alpha1.ClusterDocsTopic{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &dt)
		if err != nil {
			return false
		}

		resourceVersionMatch := dt.ResourceVersion == expectedResourceVersion
		objectMetadataMatch := dt.Name == docsTopicEntry.Id

		specBasicDataMatch := dt.Spec.DisplayName == docsTopicEntry.DisplayName &&
			dt.Spec.Description == docsTopicEntry.Description

		urlsMatch := checkUrls(docsTopicEntry.Urls, dt.Spec.Sources)
		labelsMatch := reflect.DeepEqual(dt.Labels, docsTopicEntry.Labels)

		return resourceVersionMatch && objectMetadataMatch &&
			specBasicDataMatch && urlsMatch && labelsMatch
	}
}
