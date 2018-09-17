package servicecatalog_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Create test suite to reduce boilerplate
func TestInstanceResolver_ServiceInstanceQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ServiceInstance{
			Name: "Test",
		}
		name := "name"
		environment := "environment"
		resource := &v1beta1.ServiceInstance{}
		resourceGetter := servicecatalog.NewMockInstanceService()
		resourceGetter.On("Find", name, environment).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := servicecatalog.NewMockInstanceConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewInstanceResolver(resourceGetter, nil, nil)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ServiceInstanceQuery(nil, name, environment)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		environment := "environment"
		resourceGetter := servicecatalog.NewMockInstanceService()
		resourceGetter.On("Find", name, environment).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewInstanceResolver(resourceGetter, nil, nil)

		result, err := resolver.ServiceInstanceQuery(nil, name, environment)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		environment := "environment"
		resource := &v1beta1.ServiceInstance{}
		resourceGetter := servicecatalog.NewMockInstanceService()
		resourceGetter.On("Find", name, environment).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewInstanceResolver(resourceGetter, nil, nil)

		result, err := resolver.ServiceInstanceQuery(nil, name, environment)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestInstanceResolver_ServiceInstancesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		environment := "environment"
		resource :=
			&v1beta1.ServiceInstance{
				ObjectMeta: v1.ObjectMeta{
					Name: "Test",
				},
			}
		resources := []*v1beta1.ServiceInstance{
			resource, resource,
		}
		expected := []gqlschema.ServiceInstance{
			{
				Name: "Test",
			},
			{
				Name: "Test",
			},
		}

		resourceGetter := servicecatalog.NewMockInstanceService()
		resourceGetter.On("List", environment, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := servicecatalog.NewMockInstanceConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewInstanceResolver(resourceGetter, nil, nil)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ServiceInstancesQuery(nil, environment, nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		environment := "environment"
		var resources []*v1beta1.ServiceInstance

		resourceGetter := servicecatalog.NewMockInstanceService()
		resourceGetter.On("List", environment, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewInstanceResolver(resourceGetter, nil, nil)
		var expected []gqlschema.ServiceInstance

		result, err := resolver.ServiceInstancesQuery(nil, environment, nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		environment := "environment"
		var resources []*v1beta1.ServiceInstance

		resourceGetter := servicecatalog.NewMockInstanceService()
		resourceGetter.On("List", environment, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewInstanceResolver(resourceGetter, nil, nil)

		_, err := resolver.ServiceInstancesQuery(nil, environment, nil, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestInstanceResolver_ServiceInstancesWithStatusQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		environment := "environment"
		resource :=
			&v1beta1.ServiceInstance{
				ObjectMeta: v1.ObjectMeta{
					Name: "Test",
				},
				Status: v1beta1.ServiceInstanceStatus{
					AsyncOpInProgress: false,
					Conditions: []v1beta1.ServiceInstanceCondition{
						{
							Type:    v1beta1.ServiceInstanceConditionReady,
							Status:  v1beta1.ConditionTrue,
							Message: "Working",
							Reason:  "Testing",
						},
					},
				},
			}
		resources := []*v1beta1.ServiceInstance{
			resource, resource,
		}
		expected := []gqlschema.ServiceInstance{
			{
				Name: "Test",
				Status: gqlschema.ServiceInstanceStatus{
					Type:    gqlschema.InstanceStatusTypeRunning,
					Message: "Working",
					Reason:  "Testing",
				},
			},
			{
				Name: "Test",
				Status: gqlschema.ServiceInstanceStatus{
					Type:    gqlschema.InstanceStatusTypeRunning,
					Message: "Working",
					Reason:  "Testing",
				},
			},
		}

		statusType := status.ServiceInstanceStatusTypeRunning
		gqlStatusType := gqlschema.InstanceStatusTypeRunning

		resourceGetter := servicecatalog.NewMockInstanceService()
		resourceGetter.On("ListForStatus", environment, pager.PagingParams{}, &statusType).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := servicecatalog.NewMockInstanceConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		converter.On("GQLStatusTypeToServiceStatusType", gqlStatusType).Return(statusType)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewInstanceResolver(resourceGetter, nil, nil)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ServiceInstancesQuery(nil, environment, nil, nil, &gqlStatusType)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		environment := "environment"
		statusType := status.ServiceInstanceStatusTypeRunning
		gqlStatusType := gqlschema.InstanceStatusTypeRunning
		var resources []*v1beta1.ServiceInstance

		resourceGetter := servicecatalog.NewMockInstanceService()
		resourceGetter.On("ListForStatus", environment, pager.PagingParams{}, &statusType).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewInstanceResolver(resourceGetter, nil, nil)
		var expected []gqlschema.ServiceInstance

		result, err := resolver.ServiceInstancesQuery(nil, environment, nil, nil, &gqlStatusType)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		environment := "environment"
		statusType := status.ServiceInstanceStatusTypeRunning
		gqlStatusType := gqlschema.InstanceStatusTypeRunning
		var resources []*v1beta1.ServiceInstance

		resourceGetter := servicecatalog.NewMockInstanceService()
		resourceGetter.On("ListForStatus", environment, pager.PagingParams{}, &statusType).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewInstanceResolver(resourceGetter, nil, nil)

		_, err := resolver.ServiceInstancesQuery(nil, environment, nil, nil, &gqlStatusType)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestInstanceResolver_ServiceInstanceServicePlanField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ServicePlan{
			Name: "Test",
		}
		name := "name"
		resource := &v1beta1.ClusterServicePlan{}
		resourceGetter := automock.NewPlanGetter()
		resourceGetter.On("Find", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLPlanConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:            "test",
			ServicePlanName: &name,
		}

		resolver := servicecatalog.NewInstanceResolver(nil, resourceGetter, nil)
		resolver.SetPlanConverter(converter)

		result, err := resolver.ServiceInstanceServicePlanField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := automock.NewPlanGetter()
		resourceGetter.On("Find", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:            "test",
			ServicePlanName: &name,
		}

		resolver := servicecatalog.NewInstanceResolver(nil, resourceGetter, nil)

		result, err := resolver.ServiceInstanceServicePlanField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ServicePlanName not provided", func(t *testing.T) {
		parentObj := gqlschema.ServiceInstance{
			Name:            "test",
			ServicePlanName: nil,
		}

		resolver := servicecatalog.NewInstanceResolver(nil, nil, nil)

		result, err := resolver.ServiceInstanceServicePlanField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := automock.NewPlanGetter()
		resourceGetter.On("Find", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:            "test",
			ServicePlanName: &name,
		}

		resolver := servicecatalog.NewInstanceResolver(nil, resourceGetter, nil)

		result, err := resolver.ServiceInstanceServicePlanField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestInstanceResolver_ServiceInstanceServiceClassField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ServiceClass{
			Name: "Test",
		}
		name := "name"
		resource := &v1beta1.ClusterServiceClass{}
		resourceGetter := automock.NewClassGetter()
		resourceGetter.On("Find", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLClassConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:             "test",
			ServiceClassName: &name,
		}

		resolver := servicecatalog.NewInstanceResolver(nil, nil, resourceGetter)
		resolver.SetClassConverter(converter)

		result, err := resolver.ServiceInstanceServiceClassField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := automock.NewClassGetter()
		resourceGetter.On("Find", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:             "test",
			ServiceClassName: &name,
		}

		resolver := servicecatalog.NewInstanceResolver(nil, nil, resourceGetter)

		result, err := resolver.ServiceInstanceServiceClassField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ServiceClassName not provided", func(t *testing.T) {
		parentObj := gqlschema.ServiceInstance{
			Name:             "test",
			ServiceClassName: nil,
		}

		resolver := servicecatalog.NewInstanceResolver(nil, nil, nil)

		result, err := resolver.ServiceInstanceServiceClassField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := automock.NewClassGetter()
		resourceGetter.On("Find", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:             "test",
			ServiceClassName: &name,
		}

		resolver := servicecatalog.NewInstanceResolver(nil, nil, resourceGetter)

		result, err := resolver.ServiceInstanceServiceClassField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestInstanceResolver_ServiceInstanceBindableField(t *testing.T) {
	testErr := errors.New("Test")
	className := "className"
	planName := "planName"

	for _, tc := range []struct {
		// Input
		className        *string
		planName         *string
		planErr          error
		classErr         error
		instanceBindable bool

		// Expected result
		expectedResult    bool
		shouldReturnError bool
	}{
		{className: &className, planName: &planName, planErr: nil, classErr: nil, instanceBindable: true, expectedResult: true, shouldReturnError: false},
		{className: &className, planName: &planName, planErr: nil, classErr: nil, instanceBindable: false, expectedResult: false, shouldReturnError: false},
		{className: &className, planName: &planName, planErr: testErr, classErr: nil, instanceBindable: false, expectedResult: false, shouldReturnError: true},
		{className: &className, planName: &planName, planErr: nil, classErr: testErr, instanceBindable: false, expectedResult: false, shouldReturnError: true},
		{className: &className, planName: nil, planErr: nil, classErr: nil, instanceBindable: false, expectedResult: false, shouldReturnError: false},
		{className: nil, planName: &planName, planErr: nil, classErr: nil, instanceBindable: false, expectedResult: false, shouldReturnError: false},
	} {
		class := &v1beta1.ClusterServiceClass{}
		plan := &v1beta1.ClusterServicePlan{}

		planGetter := automock.NewPlanGetter()
		planGetter.On("Find", planName).Return(plan, tc.planErr).Once()
		classGetter := automock.NewClassGetter()
		classGetter.On("Find", className).Return(class, tc.classErr).Once()
		instanceSvc := servicecatalog.NewMockInstanceService()
		instanceSvc.On("IsBindable", class, plan).Return(tc.instanceBindable).Once()

		resolver := servicecatalog.NewInstanceResolver(instanceSvc, planGetter, classGetter)

		parentObj := &gqlschema.ServiceInstance{
			Name:             "test",
			ServiceClassName: tc.className,
			ServicePlanName:  tc.planName,
		}

		result, err := resolver.ServiceInstanceBindableField(nil, parentObj)

		if tc.shouldReturnError {
			require.Error(t, err)
			assert.True(t, gqlerror.IsInternal(err))
		} else {
			require.NoError(t, err)
		}
		assert.Equal(t, tc.expectedResult, result)
	}
}

func TestInstanceResolver_ServiceInstanceEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := servicecatalog.NewMockInstanceService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalog.NewInstanceResolver(svc, nil, nil)

		_, err := resolver.ServiceInstanceEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := servicecatalog.NewMockInstanceService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalog.NewInstanceResolver(svc, nil, nil)

		channel, err := resolver.ServiceInstanceEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
