package servicecatalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceBrokerResolver_ServiceBrokerQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		env := "env"
		expected := &gqlschema.ServiceBroker{
			Name: "Test",
		}
		resource := &v1beta1.ServiceBroker{}

		svc := automock.NewServiceBrokerService()
		svc.On("Find", name, env).
			Return(resource, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLServiceBrokerConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceBrokerResolver(svc)
		resolver.SetBrokerConverter(converter)

		result, err := resolver.ServiceBrokerQuery(nil, name, env)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		env := "env"
		svc := automock.NewServiceBrokerService()
		svc.On("Find", name, env).Return(nil, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalog.NewServiceBrokerResolver(svc)

		result, err := resolver.ServiceBrokerQuery(nil, name, env)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		name := "name"
		env := "env"
		expected := errors.New("Test")

		resource := &v1beta1.ServiceBroker{}

		svc := automock.NewServiceBrokerService()
		svc.On("Find", name, env).Return(resource, expected).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalog.NewServiceBrokerResolver(svc)

		result, err := resolver.ServiceBrokerQuery(nil, name, env)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestServiceBrokerResolver_ServiceBrokersQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		env := "env"
		resource :=
			&v1beta1.ServiceBroker{
				ObjectMeta: v1.ObjectMeta{
					Name: "test",
				},
			}
		resources := []*v1beta1.ServiceBroker{
			resource, resource,
		}
		expected := []gqlschema.ServiceBroker{
			{
				Name: "Test",
			}, {
				Name: "Test",
			},
		}

		svc := automock.NewServiceBrokerService()
		svc.On("List", env, pager.PagingParams{}).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLServiceBrokerConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceBrokerResolver(svc)
		resolver.SetBrokerConverter(converter)

		result, err := resolver.ServiceBrokersQuery(nil, env, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		env := "env"
		var resources []*v1beta1.ServiceBroker

		svc := automock.NewServiceBrokerService()
		svc.On("List", env, pager.PagingParams{}).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBrokerResolver(svc)
		var expected []gqlschema.ServiceBroker

		result, err := resolver.ServiceBrokersQuery(nil, env, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		env := "env"
		expected := errors.New("Test")

		var resources []*v1beta1.ServiceBroker

		svc := automock.NewServiceBrokerService()
		svc.On("List", env, pager.PagingParams{}).Return(resources, expected).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBrokerResolver(svc)

		_, err := resolver.ServiceBrokersQuery(nil, env, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceBrokerResolver_ServiceBrokerEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewServiceBrokerService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalog.NewServiceBrokerResolver(svc)

		_, err := resolver.ServiceBrokerEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := automock.NewServiceBrokerService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalog.NewServiceBrokerResolver(svc)

		channel, err := resolver.ServiceBrokerEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
