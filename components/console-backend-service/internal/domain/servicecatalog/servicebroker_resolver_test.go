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

func TestServiceBrokerResolver_ServiceBrokerQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "name"
		ns := "ns"
		expected := &gqlschema.ServiceBroker{
			Name: "Test",
		}
		resource := &v1beta1.ServiceBroker{}

		svc := automock.NewServiceBrokerService()
		svc.On("Find", name, ns).
			Return(resource, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLServiceBrokerConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceBrokerResolver(svc)
		resolver.SetBrokerConverter(converter)

		result, err := resolver.ServiceBrokerQuery(nil, name, ns)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		ns := "ns"
		svc := automock.NewServiceBrokerService()
		svc.On("Find", name, ns).Return(nil, nil).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalog.NewServiceBrokerResolver(svc)

		result, err := resolver.ServiceBrokerQuery(nil, name, ns)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		name := "name"
		ns := "ns"
		expected := errors.New("Test")

		resource := &v1beta1.ServiceBroker{}

		svc := automock.NewServiceBrokerService()
		svc.On("Find", name, ns).Return(resource, expected).Once()
		defer svc.AssertExpectations(t)

		resolver := servicecatalog.NewServiceBrokerResolver(svc)

		result, err := resolver.ServiceBrokerQuery(nil, name, ns)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestServiceBrokerResolver_ServiceBrokersQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ns := "ns"
		resource :=
			&v1beta1.ServiceBroker{
				ObjectMeta: v1.ObjectMeta{
					Name: "test",
				},
			}
		resources := []*v1beta1.ServiceBroker{
			resource, resource,
		}
		expected := []*gqlschema.ServiceBroker{
			{
				Name: "Test",
			}, {
				Name: "Test",
			},
		}

		svc := automock.NewServiceBrokerService()
		svc.On("List", ns, pager.PagingParams{}).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)

		converter := automock.NewGQLServiceBrokerConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceBrokerResolver(svc)
		resolver.SetBrokerConverter(converter)

		result, err := resolver.ServiceBrokersQuery(nil, ns, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		ns := "ns"
		var resources []*v1beta1.ServiceBroker

		svc := automock.NewServiceBrokerService()
		svc.On("List", ns, pager.PagingParams{}).Return(resources, nil).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBrokerResolver(svc)
		var expected []*gqlschema.ServiceBroker

		result, err := resolver.ServiceBrokersQuery(nil, ns, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		ns := "ns"
		expected := errors.New("Test")

		var resources []*v1beta1.ServiceBroker

		svc := automock.NewServiceBrokerService()
		svc.On("List", ns, pager.PagingParams{}).Return(resources, expected).Once()
		defer svc.AssertExpectations(t)
		resolver := servicecatalog.NewServiceBrokerResolver(svc)

		_, err := resolver.ServiceBrokersQuery(nil, ns, nil, nil)

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
