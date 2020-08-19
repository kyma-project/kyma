package servicecatalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterServiceBrokerResolver_ClusterServiceBrokerQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ClusterServiceBroker{
			Name: "Test",
		}
		resource := &v1beta1.ClusterServiceBroker{}

		svc := automock.NewClusterServiceBrokerService()
		svc.On("Find", "broker").
			Return(resource, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLClusterServiceBrokerConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewClusterServiceBrokerResolver(svc)
		resolver.SetBrokerConverter(converter)

		result, err := resolver.ClusterServiceBrokerQuery(nil, "broker")

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		svc := automock.NewClusterServiceBrokerService()
		svc.On("Find", name).Return(nil, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalog.NewClusterServiceBrokerResolver(svc)

		result, err := resolver.ClusterServiceBrokerQuery(nil, name)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"

		resource := &v1beta1.ClusterServiceBroker{}

		svc := automock.NewClusterServiceBrokerService()
		svc.On("Find", name).Return(resource, expected).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalog.NewClusterServiceBrokerResolver(svc)

		result, err := resolver.ClusterServiceBrokerQuery(nil, name)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestClusterServiceBrokerResolver_ClusterServiceBrokersQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		resource :=
			&v1beta1.ClusterServiceBroker{
				ObjectMeta: v1.ObjectMeta{
					Name: "test",
				},
			}
		resources := []*v1beta1.ClusterServiceBroker{
			resource, resource,
		}
		expected := []*gqlschema.ClusterServiceBroker{
			{
				Name: "Test",
			}, {
				Name: "Test",
			},
		}

		svc := automock.NewClusterServiceBrokerService()
		svc.On("List", pager.PagingParams{}).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLClusterServiceBrokerConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewClusterServiceBrokerResolver(svc)
		resolver.SetBrokerConverter(converter)

		result, err := resolver.ClusterServiceBrokersQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		var resources []*v1beta1.ClusterServiceBroker

		svc := automock.NewClusterServiceBrokerService()
		svc.On("List", pager.PagingParams{}).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewClusterServiceBrokerResolver(svc)
		var expected []*gqlschema.ClusterServiceBroker

		result, err := resolver.ClusterServiceBrokersQuery(nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		var resources []*v1beta1.ClusterServiceBroker

		svc := automock.NewClusterServiceBrokerService()
		svc.On("List", pager.PagingParams{}).Return(resources, expected).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewClusterServiceBrokerResolver(svc)

		_, err := resolver.ClusterServiceBrokersQuery(nil, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestClusterServiceBrokerResolver_ClusterServiceBrokerEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewClusterServiceBrokerService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalog.NewClusterServiceBrokerResolver(svc)

		_, err := resolver.ClusterServiceBrokerEventSubscription(ctx)

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewClusterServiceBrokerService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalog.NewClusterServiceBrokerResolver(svc)

		channel, err := resolver.ClusterServiceBrokerEventSubscription(ctx)
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
