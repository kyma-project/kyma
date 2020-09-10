package rafter

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/clusterassetgroup"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/mocks"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestUpsertClusterAssetGroup(t *testing.T) {
	t.Run("Should create ClusterAssetGroup", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)

		clusterAssetGroupEntry := createTestClusterAssetGroupEntry()

		resourceInterfaceMock.On("Get", context.Background(), "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{}, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		resourceInterfaceMock.On("Create", context.Background(), mock.MatchedBy(createMatcherFunction(clusterAssetGroupEntry, "")), metav1.CreateOptions{}).
			Return(&unstructured.Unstructured{}, nil)

		// when
		err := repository.Upsert(clusterAssetGroupEntry)

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should update ClusterAssetGroup", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)

		dt := createK8sClusterAssetGroup()

		object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&dt)
		require.NoError(t, err)

		resourceInterfaceMock.On("Get", context.Background(), "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		resourceInterfaceMock.On("Get", context.Background(), "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		clusterAssetGroupEntry := createTestClusterAssetGroupEntry()
		resourceInterfaceMock.On("Update", context.Background(), mock.MatchedBy(createMatcherFunction(clusterAssetGroupEntry, "1")), metav1.UpdateOptions{}).Return(&unstructured.Unstructured{}, nil)

		// when
		err = repository.Upsert(clusterAssetGroupEntry)

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Create", 0)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("Should fail if k8s client returned error on Get", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Get", context.Background(), mock.Anything, metav1.GetOptions{}).
			Return(&unstructured.Unstructured{}, errors.New("some error"))

		// when
		err := repository.Upsert(createTestClusterAssetGroupEntry())

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should fail if k8s client returned error on Create", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Get", context.Background(), "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{}, k8serrors.NewNotFound(schema.GroupResource{}, ""))
		resourceInterfaceMock.On("Create", context.Background(), mock.Anything, metav1.CreateOptions{}).
			Return(&unstructured.Unstructured{}, errors.New("some error"))

		// when
		err := repository.Upsert(createTestClusterAssetGroupEntry())

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should fail if k8s client returned error on Update", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)

		dt := createK8sClusterAssetGroup()
		object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&dt)
		require.NoError(t, err)

		resourceInterfaceMock.On("Get", context.Background(), "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		resourceInterfaceMock.On("Get", context.Background(), "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		resourceInterfaceMock.On("Update", context.Background(), mock.Anything, metav1.UpdateOptions{}).Return(&unstructured.Unstructured{}, errors.New("some error"))

		// when
		err = repository.Upsert(createTestClusterAssetGroupEntry())

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Get", 2)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Create", 0)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Update", 1)
	})
}

func TestGetClusterAssetGroup(t *testing.T) {
	t.Run("Should get ClusterAssetGroup", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)
		{
			dt := createK8sClusterAssetGroup()

			object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&dt)
			require.NoError(t, err)

			resourceInterfaceMock.On("Get", context.Background(), "id1", metav1.GetOptions{}).
				Return(&unstructured.Unstructured{Object: object}, nil)
		}

		// when
		clusterAssetGroup, err := repository.Get("id1")
		require.NoError(t, err)

		// then
		assert.Equal(t, "Some display name", clusterAssetGroup.DisplayName)
		assert.Equal(t, "Some description", clusterAssetGroup.Description)
		assert.Equal(t, "id1", clusterAssetGroup.Id)
		assert.Equal(t, len(clusterAssetGroup.Urls), 1)
		assert.Equal(t, clusterAssetGroup.Urls["api"], "www.somestorage.com/api")
		assert.Equal(t, len(clusterAssetGroup.Labels), 1)
		assert.Equal(t, "value", clusterAssetGroup.Labels["key"])
	})

	t.Run("Should fail with Not Found if ClusterAssetGroup doesn't exist", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Get", context.Background(), mock.Anything, metav1.GetOptions{}).
			Return(&unstructured.Unstructured{}, k8serrors.NewNotFound(schema.GroupResource{}, ""))

		// when
		_, err := repository.Get("id1")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func createK8sClusterAssetGroup() v1beta1.ClusterAssetGroup {
	return v1beta1.ClusterAssetGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "id1",
			Namespace: "kyma-integration",
			Labels: map[string]string{
				"key": "value",
			},
			ResourceVersion: "1",
		},
		Spec: v1beta1.ClusterAssetGroupSpec{
			CommonAssetGroupSpec: v1beta1.CommonAssetGroupSpec{
				DisplayName: "Some display name",
				Description: "Some description",
				Sources: []v1beta1.Source{
					{
						URL:  "www.somestorage.com/api",
						Mode: v1beta1.AssetGroupSingle,
						Type: "api",
					},
				},
			},
		}}
}

func TestDeleteClusterAssetGroup(t *testing.T) {
	t.Run("Should delete ClusterAssetGroup", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Delete", context.Background(), "id1", metav1.DeleteOptions{}).Return(nil)

		// when
		err := repository.Delete("id1")

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should fail if k8s client returned error", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Delete", context.Background(), "id1", metav1.DeleteOptions{}).Return(errors.New("some error"))

		// when
		err := repository.Delete("id1")

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should not fail if Docs Topic doesn't exist", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewClusterAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Delete", context.Background(), "id1", metav1.DeleteOptions{}).Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))

		// when
		err := repository.Delete("id1")

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})
}

func createTestClusterAssetGroupEntry() clusterassetgroup.Entry {
	return clusterassetgroup.Entry{
		Id:          "id1",
		DisplayName: "Some display name",
		Description: "Some description",
		Urls: map[string]string{
			clusterassetgroup.KeyOpenApiSpec: "www.somestorage.com/api",
		},
		Labels: map[string]string{
			"key": "value",
		},
	}
}

func createMatcherFunction(clusterAssetGroupEntry clusterassetgroup.Entry, expectedResourceVersion string) func(*unstructured.Unstructured) bool {
	findSource := func(sources []v1beta1.Source, key string) (v1beta1.Source, bool) {
		for _, source := range sources {
			if string(source.Type) == key && string(source.Name) == fmt.Sprintf(ClusterAssetGroupNameFormat, key, clusterAssetGroupEntry.Id) {
				return source, true
			}
		}

		return v1beta1.Source{}, false
	}

	checkUrls := func(urls map[string]string, sources []v1beta1.Source) bool {
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
		dt := v1beta1.ClusterAssetGroup{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &dt)
		if err != nil {
			return false
		}

		resourceVersionMatch := dt.ResourceVersion == expectedResourceVersion
		objectMetadataMatch := dt.Name == clusterAssetGroupEntry.Id

		specBasicDataMatch := dt.Spec.DisplayName == clusterAssetGroupEntry.DisplayName &&
			dt.Spec.Description == clusterAssetGroupEntry.Description

		urlsMatch := checkUrls(clusterAssetGroupEntry.Urls, dt.Spec.Sources)
		labelsMatch := reflect.DeepEqual(dt.Labels, clusterAssetGroupEntry.Labels)

		return resourceVersionMatch && objectMetadataMatch &&
			specBasicDataMatch && urlsMatch && labelsMatch
	}
}
