package application_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/automock"
	rafterMock "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared/automock"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter/spec"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
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

		resolver := application.NewEventActivationResolver(svc, nil)
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

		resolver := application.NewEventActivationResolver(svc, nil)
		result, err := resolver.EventActivationsQuery(nil, "test")

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		svc := automock.NewEventActivationLister()
		svc.On("List", "test").Return(nil, errors.New("trol"))
		defer svc.AssertExpectations(t)

		resolver := application.NewEventActivationResolver(svc, nil)
		_, err := resolver.EventActivationsQuery(nil, "test")

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestEventActivationResolver_EventActivationEventsField(t *testing.T) {
	asyncApiBaseUrl := "example.com"
	asyncApiFileName := "asyncApiSpec.json"
	clusterAssets := []*v1beta1.ClusterAsset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "exampleName",
			},
			Status: v1beta1.ClusterAssetStatus{
				CommonAssetStatus: v1beta1.CommonAssetStatus{
					Phase: v1beta1.AssetReady,
					AssetRef: v1beta1.AssetStatusRef{
						BaseURL: asyncApiBaseUrl,
						Files: []v1beta1.AssetFile{
							{
								Name: asyncApiFileName,
							},
						},
					},
				},
			},
		},
	}
	types := []string{"asyncapi", "asyncApi", "asyncapispec", "asyncApiSpec", "events", "async-api"}

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

		clusterAssetGetter := new(rafterMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForClusterAssetGroupByType", "test", types).Return(clusterAssets, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		specificationGetter := new(rafterMock.SpecificationGetter)
		specificationGetter.On("AsyncAPI", asyncApiBaseUrl, asyncApiFileName).Return(asyncApiSpec, nil)
		defer specificationGetter.AssertExpectations(t)

		retriever := new(rafterMock.RafterRetriever)
		retriever.On("ClusterAsset").Return(clusterAssetGetter)
		retriever.On("Specification").Return(specificationGetter)

		resolver := application.NewEventActivationResolver(nil, retriever)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Contains(t, result, *fixGQLEventActivationEvent("sell", "v1", "desc"))
		assert.Contains(t, result, *fixGQLEventActivationEvent("sell", "v2", "desc"))
	})

	t.Run("Not found", func(t *testing.T) {
		clusterAssetGetter := new(rafterMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForClusterAssetGroupByType", "test", types).Return(nil, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		assetStoreRetriever := new(rafterMock.RafterRetriever)
		assetStoreRetriever.On("ClusterAsset").Return(clusterAssetGetter)

		resolver := application.NewEventActivationResolver(nil, assetStoreRetriever)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Not ready", func(t *testing.T) {
		asset := v1beta1.ClusterAsset{}
		asset.Status.Phase = v1beta1.AssetFailed

		clusterAssetGetter := new(rafterMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForClusterAssetGroupByType", "test", types).Return([]*v1beta1.ClusterAsset{&asset}, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		assetStoreRetriever := new(rafterMock.RafterRetriever)
		assetStoreRetriever.On("ClusterAsset").Return(clusterAssetGetter)

		resolver := application.NewEventActivationResolver(nil, assetStoreRetriever)
		result, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("No files", func(t *testing.T) {
		asset := v1beta1.ClusterAsset{}
		asset.Status.Phase = v1beta1.AssetReady

		clusterAssetGetter := new(rafterMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForClusterAssetGroupByType", "test", types).Return([]*v1beta1.ClusterAsset{&asset}, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		assetStoreRetriever := new(rafterMock.RafterRetriever)
		assetStoreRetriever.On("ClusterAsset").Return(clusterAssetGetter)

		resolver := application.NewEventActivationResolver(nil, assetStoreRetriever)
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

		clusterAssetGetter := new(rafterMock.ClusterAssetGetter)
		clusterAssetGetter.On("ListForClusterAssetGroupByType", "test", types).Return(clusterAssets, nil)
		defer clusterAssetGetter.AssertExpectations(t)

		specificationGetter := new(rafterMock.SpecificationGetter)
		specificationGetter.On("AsyncAPI", asyncApiBaseUrl, asyncApiFileName).Return(asyncApiSpec, nil)
		defer specificationGetter.AssertExpectations(t)

		retriever := new(rafterMock.RafterRetriever)
		retriever.On("ClusterAsset").Return(clusterAssetGetter)
		retriever.On("Specification").Return(specificationGetter)

		resolver := application.NewEventActivationResolver(nil, retriever)
		_, err := resolver.EventActivationEventsField(nil, fixGQLEventActivation("test"))

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Nil", func(t *testing.T) {
		getter := new(rafterMock.ClusterAssetGetter)

		retriever := new(rafterMock.RafterRetriever)
		retriever.On("ClusterAsset").Return(getter)

		resolver := application.NewEventActivationResolver(nil, retriever)
		_, err := resolver.EventActivationEventsField(nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})

	t.Run("Error", func(t *testing.T) {
		getter := new(rafterMock.ClusterAssetGetter)
		getter.On("ListForClusterAssetGroupByType", "test", types).Return(nil, errors.New("nope"))
		defer getter.AssertExpectations(t)

		retriever := new(rafterMock.RafterRetriever)
		retriever.On("ClusterAsset").Return(getter)

		resolver := application.NewEventActivationResolver(nil, retriever)
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
