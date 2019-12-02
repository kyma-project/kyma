package application_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/automock"
	assetstoreMock "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/assetstore/spec"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEventActivationResolver_EventActivationsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		eventActivation1 := fixEventActivation("test", "event1")
		eventActivation2 := fixEventActivation("test", "event2")

		svc := automock.NewEventActivationLister()
		svc.On("List", "test").Return([]*v1alpha1.EventActivation{
			eventActivation1,
			eventActivation2,
		}, nil)
		defer svc.AssertExpectations(t)

		resolver := application.NewEventActivationResolver(svc, nil, nil)
		result, err := resolver.EventActivationsQuery(nil, "test")

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, *fixGQLEventActivation("event1"))
		assert.Contains(t, result, *fixGQLEventActivation("event2"))
	})

	t.Run("Not found", func(t *testing.T) {
		svc := automock.NewEventActivationLister()
		svc.On("List", "test").Return([]*v1alpha1.EventActivation{}, nil)
		defer svc.AssertExpectations(t)

		resolver := application.NewEventActivationResolver(svc, nil, nil)
		result, err := resolver.EventActivationsQuery(nil, "test")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewEventActivationLister()
		svc.On("List", "test").Return(nil, errors.New("trol"))
		defer svc.AssertExpectations(t)

		resolver := application.NewEventActivationResolver(svc, nil, nil)
		_, err := resolver.EventActivationsQuery(nil, "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestEventActivationResolver_EventActivationEventsField(t *testing.T) {
	asyncApiBaseUrl := "example.com"
	asyncApiFileName := "asyncApiSpec.json"

	clusterAssets := []*v1alpha2.ClusterAsset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "exampleName",
			},
			Status: v1alpha2.ClusterAssetStatus{
				CommonAssetStatus: v1alpha2.CommonAssetStatus{
					Phase: v1alpha2.AssetReady,
					AssetRef: v1alpha2.AssetStatusRef{
						BaseURL: asyncApiBaseUrl,
						Files: []v1alpha2.AssetFile{
							{
								Name: asyncApiFileName,
							},
						},
					},
				},
			},
		},
	}

	types := []string{"asyncapi", "asyncApi", "asyncapispec", "asyncApiSpec", "events"}

	t.Run("Success", func(t *testing.T) {
		asyncApiSpec := &spec.AsyncAPISpec{
			Data: spec.AsyncAPISpecData{
				AsyncAPI: "2.0.0",
				Channels: map[string]interface{}{
					"sell.v1": map[string]interface{}{
						"subscribe": map[string]interface{}{
							"message": map[string]interface{}{
								"summary": "desc",
							},
						},
					},
					"sell.v2": map[string]interface{}{
						"subscribe": map[string]interface{}{
							"message": map[string]interface{}{
								"summary": "desc",
							},
						},
					},
				},
			},
		}

		clusterAssetGetter := new(assetstoreMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForDocsTopicByType", "test", types).Return(clusterAssets, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		specificationGetter := new(assetstoreMock.SpecificationGetter)
		specificationGetter.On("AsyncAPI", asyncApiBaseUrl, asyncApiFileName).Return(asyncApiSpec, nil)
		defer specificationGetter.AssertExpectations(t)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("ClusterAsset").Return(clusterAssetGetter)
		retriever.On("Specification").Return(specificationGetter)

		resolver := application.NewEventActivationResolver(nil, retriever, nil)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, *fixGQLEventActivationEvent("sell", "v1", "desc"))
		assert.Contains(t, result, *fixGQLEventActivationEvent("sell", "v2", "desc"))
	})

	t.Run("Not found", func(t *testing.T) {
		clusterAssetGetter := new(assetstoreMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForDocsTopicByType", "test", types).Return(nil, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		assetStoreRetriever := new(assetstoreMock.AssetStoreRetriever)
		assetStoreRetriever.On("ClusterAsset").Return(clusterAssetGetter)

		resolver := application.NewEventActivationResolver(nil, assetStoreRetriever, nil)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Not ready", func(t *testing.T) {
		asset := v1alpha2.ClusterAsset{}
		asset.Status.Phase = v1alpha2.AssetFailed

		clusterAssetGetter := new(assetstoreMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForDocsTopicByType", "test", types).Return([]*v1alpha2.ClusterAsset{&asset}, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		assetStoreRetriever := new(assetstoreMock.AssetStoreRetriever)
		assetStoreRetriever.On("ClusterAsset").Return(clusterAssetGetter)

		resolver := application.NewEventActivationResolver(nil, assetStoreRetriever, nil)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("No files", func(t *testing.T) {
		asset := v1alpha2.ClusterAsset{}
		asset.Status.Phase = v1alpha2.AssetReady

		clusterAssetGetter := new(assetstoreMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForDocsTopicByType", "test", types).Return([]*v1alpha2.ClusterAsset{&asset}, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		assetStoreRetriever := new(assetstoreMock.AssetStoreRetriever)
		assetStoreRetriever.On("ClusterAsset").Return(clusterAssetGetter)

		resolver := application.NewEventActivationResolver(nil, assetStoreRetriever, nil)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Invalid version", func(t *testing.T) {
		asyncApiSpec := &spec.AsyncAPISpec{
			Data: spec.AsyncAPISpecData{
				AsyncAPI: "1.0.1",
			},
		}

		clusterAssetGetter := new(assetstoreMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForDocsTopicByType", "test", types).Return(clusterAssets, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		specificationGetter := new(assetstoreMock.SpecificationGetter)
		specificationGetter.On("AsyncAPI", asyncApiBaseUrl, asyncApiFileName).Return(asyncApiSpec, nil)
		defer specificationGetter.AssertExpectations(t)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("ClusterAsset").Return(clusterAssetGetter)
		retriever.On("Specification").Return(specificationGetter)

		resolver := application.NewEventActivationResolver(nil, retriever, nil)
		_, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Nil", func(t *testing.T) {
		getter := new(assetstoreMock.ClusterAssetGetter)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("ClusterAsset").Return(getter)

		resolver := application.NewEventActivationResolver(nil, retriever, nil)
		_, err := resolver.EventActivationEventsField(nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Error", func(t *testing.T) {
		getter := new(assetstoreMock.ClusterAssetGetter)
		getter.On("ListForDocsTopicByType", "test", types).Return(nil, errors.New("nope"))
		defer getter.AssertExpectations(t)

		retriever := new(assetstoreMock.AssetStoreRetriever)
		retriever.On("ClusterAsset").Return(getter)

		resolver := application.NewEventActivationResolver(nil, retriever, nil)
		_, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func fixGQLEventActivation(name string) *gqlschema.EventActivation {
	return &gqlschema.EventActivation{
		Name:        name,
		DisplayName: "aha!",
		SourceID:    "picco-bello",
	}
}

func fixGQLEventActivationEvent(eventType, version, desc string) *gqlschema.EventActivationEvent {
	return &gqlschema.EventActivationEvent{
		EventType:   eventType,
		Version:     version,
		Description: desc,
	}
}
