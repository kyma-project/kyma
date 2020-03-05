package rafter

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/mocks"
)

func TestCreateClusterAssetGroup(t *testing.T) {
	t.Run("Should create ClusterAssetGroup", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewAssetGroupRepository(resourceInterfaceMock)

		clusterAssetGroupEntry := createTestClusterAssetGroupEntry()

		resourceInterfaceMock.On("Create", mock.MatchedBy(createMatcherFunction(clusterAssetGroupEntry, "")), metav1.CreateOptions{}).
			Return(&unstructured.Unstructured{}, nil)

		// when
		err := repository.Create(clusterAssetGroupEntry)

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should fail if k8s client returned error on Create", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Create", mock.Anything, metav1.CreateOptions{}).
			Return(&unstructured.Unstructured{}, errors.New("some error"))

		// when
		err := repository.Create(createTestClusterAssetGroupEntry())

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})
}

func TestUpdateClusterAssetGroup(t *testing.T) {
	t.Run("Should update ClusterAssetGroup", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewAssetGroupRepository(resourceInterfaceMock)

		ag := createK8sClusterAssetGroup()

		object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ag)
		require.NoError(t, err)

		resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		clusterAssetGroupEntry := createTestClusterAssetGroupEntry()
		resourceInterfaceMock.On("Update", mock.MatchedBy(createMatcherFunction(clusterAssetGroupEntry, "1")), metav1.UpdateOptions{}).Return(&unstructured.Unstructured{}, nil)

		// when
		err = repository.Update(clusterAssetGroupEntry)

		// then
		require.NoError(t, err)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Update", 1)
	})

	t.Run("Should fail if k8s client returned error on Get", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Get", mock.Anything, metav1.GetOptions{}).
			Return(&unstructured.Unstructured{}, errors.New("some error"))

		// when
		err := repository.Update(createTestClusterAssetGroupEntry())

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertExpectations(t)
	})

	t.Run("Should fail if k8s client returned error on Update", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewAssetGroupRepository(resourceInterfaceMock)

		ag := createK8sClusterAssetGroup()
		object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ag)
		require.NoError(t, err)

		resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
			Return(&unstructured.Unstructured{Object: object}, nil)

		resourceInterfaceMock.On("Update", mock.Anything, metav1.UpdateOptions{}).Return(&unstructured.Unstructured{}, errors.New("some error"))

		// when
		err = repository.Update(createTestClusterAssetGroupEntry())

		// then
		require.Error(t, err)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Get", 1)
		resourceInterfaceMock.AssertNumberOfCalls(t, "Update", 1)
	})
}

func TestGetClusterAssetGroup(t *testing.T) {
	t.Run("Should get ClusterAssetGroup", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewAssetGroupRepository(resourceInterfaceMock)
		{
			ag := createK8sClusterAssetGroup()

			object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ag)
			require.NoError(t, err)

			resourceInterfaceMock.On("Get", "id1", metav1.GetOptions{}).
				Return(&unstructured.Unstructured{Object: object}, nil)
		}

		// when
		clusterAssetGroup, err := repository.Get("id1")
		require.NoError(t, err)

		// then
		assert.Equal(t, "Some display name", clusterAssetGroup.DisplayName)
		assert.Equal(t, "Some description", clusterAssetGroup.Description)
		assert.Equal(t, "id1", clusterAssetGroup.Id)
		assert.Equal(t, len(clusterAssetGroup.Assets), 1)
		assert.Equal(t, clusterAssetGroup.Assets[0].Url, "www.somestorage.com/api")
		assert.Equal(t, len(clusterAssetGroup.Labels), 1)
		assert.Equal(t, "value", clusterAssetGroup.Labels["key"])
	})

	t.Run("Should fail with Not Found if ClusterAssetGroup doesn't exist", func(t *testing.T) {
		// given
		resourceInterfaceMock := &mocks.ResourceInterface{}
		repository := NewAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Get", mock.Anything, metav1.GetOptions{}).
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
		repository := NewAssetGroupRepository(resourceInterfaceMock)

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
		repository := NewAssetGroupRepository(resourceInterfaceMock)

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
		repository := NewAssetGroupRepository(resourceInterfaceMock)

		resourceInterfaceMock.On("Delete", "id1", &metav1.DeleteOptions{}).Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))

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
		Labels: map[string]string{
			"key": "value",
		},
		Assets: []clusterassetgroup.Asset{{
			Name:     "id1",
			Type:     clusterassetgroup.OpenApiType,
			Format:   clusterassetgroup.SpecFormatYAML,
			Url:      "www.somestorage.com/api",
			SpecHash: "39faae9f5e6e58d758bce2c88a247a45",
		},
		},
	}
}

func createMatcherFunction(clusterAssetGroupEntry clusterassetgroup.Entry, expectedResourceVersion string) func(*unstructured.Unstructured) bool {
	findSource := func(sources []v1beta1.Source, assetName string, assetType clusterassetgroup.ApiType) (v1beta1.Source, bool) {
		for _, source := range sources {
			if source.Type == v1beta1.AssetGroupSourceType(assetType) && source.Name == v1beta1.AssetGroupSourceName(assetName) {
				return source, true
			}
		}

		return v1beta1.Source{}, false
	}

	checkAssets := func(assets []clusterassetgroup.Asset, sources []v1beta1.Source) bool {
		if len(assets) != len(sources) {
			return false
		}

		for _, asset := range assets {
			source, found := findSource(sources, asset.Name, asset.Type)
			if !found || asset.Url != source.URL {
				return false
			}
		}

		return true
	}

	return func(u *unstructured.Unstructured) bool {
		ag := v1beta1.ClusterAssetGroup{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ag)
		if err != nil {
			return false
		}

		resourceVersionMatch := ag.ResourceVersion == expectedResourceVersion
		objectMetadataMatch := ag.Name == clusterAssetGroupEntry.Id

		specBasicDataMatch := ag.Spec.DisplayName == clusterAssetGroupEntry.DisplayName &&
			ag.Spec.Description == clusterAssetGroupEntry.Description

		urlsMatch := checkAssets(clusterAssetGroupEntry.Assets, ag.Spec.Sources)
		labelsMatch := reflect.DeepEqual(ag.Labels, clusterAssetGroupEntry.Labels)

		return resourceVersionMatch && objectMetadataMatch &&
			specBasicDataMatch && urlsMatch && labelsMatch
	}
}
