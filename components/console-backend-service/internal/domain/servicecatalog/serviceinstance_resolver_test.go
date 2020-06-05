package servicecatalog_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceInstanceResolver_ServiceInstanceQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ServiceInstance{
			Name: "Test",
		}
		name := "name"
		namespace := "namespace"
		resource := &v1beta1.ServiceInstance{}
		resourceGetter := servicecatalog.NewMockServiceInstanceService()
		resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := servicecatalog.NewMockServiceInstanceConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceInstanceResolver(resourceGetter, nil, nil, nil, nil)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ServiceInstanceQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		namespace := "namespace"
		resourceGetter := servicecatalog.NewMockServiceInstanceService()
		resourceGetter.On("Find", name, namespace).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceInstanceResolver(resourceGetter, nil, nil, nil, nil)

		result, err := resolver.ServiceInstanceQuery(nil, name, namespace)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")
		name := "name"
		namespace := "namespace"
		resource := &v1beta1.ServiceInstance{}
		resourceGetter := servicecatalog.NewMockServiceInstanceService()
		resourceGetter.On("Find", name, namespace).Return(resource, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewServiceInstanceResolver(resourceGetter, nil, nil, nil, nil)

		result, err := resolver.ServiceInstanceQuery(nil, name, namespace)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestServiceInstanceResolver_ServiceInstancesQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "namespace"
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

		resourceGetter := servicecatalog.NewMockServiceInstanceService()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := servicecatalog.NewMockServiceInstanceConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceInstanceResolver(resourceGetter, nil, nil, nil, nil)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ServiceInstancesQuery(nil, namespace, nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "namespace"
		var resources []*v1beta1.ServiceInstance

		resourceGetter := servicecatalog.NewMockServiceInstanceService()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewServiceInstanceResolver(resourceGetter, nil, nil, nil, nil)
		var expected []gqlschema.ServiceInstance

		result, err := resolver.ServiceInstancesQuery(nil, namespace, nil, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		namespace := "namespace"
		var resources []*v1beta1.ServiceInstance

		resourceGetter := servicecatalog.NewMockServiceInstanceService()
		resourceGetter.On("List", namespace, pager.PagingParams{}).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewServiceInstanceResolver(resourceGetter, nil, nil, nil, nil)

		_, err := resolver.ServiceInstancesQuery(nil, namespace, nil, nil, nil)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceInstanceResolver_ServiceInstancesWithStatusQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "namespace"
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

		resourceGetter := servicecatalog.NewMockServiceInstanceService()
		resourceGetter.On("ListForStatus", namespace, pager.PagingParams{}, &statusType).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := servicecatalog.NewMockServiceInstanceConverter()
		converter.On("ToGQLs", resources).Return(expected, nil)
		converter.On("GQLStatusTypeToServiceStatusType", gqlStatusType).Return(statusType)
		defer converter.AssertExpectations(t)

		resolver := servicecatalog.NewServiceInstanceResolver(resourceGetter, nil, nil, nil, nil)
		resolver.SetInstanceConverter(converter)

		result, err := resolver.ServiceInstancesQuery(nil, namespace, nil, nil, &gqlStatusType)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "namespace"
		statusType := status.ServiceInstanceStatusTypeRunning
		gqlStatusType := gqlschema.InstanceStatusTypeRunning
		var resources []*v1beta1.ServiceInstance

		resourceGetter := servicecatalog.NewMockServiceInstanceService()
		resourceGetter.On("ListForStatus", namespace, pager.PagingParams{}, &statusType).Return(resources, nil).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewServiceInstanceResolver(resourceGetter, nil, nil, nil, nil)
		var expected []gqlschema.ServiceInstance

		result, err := resolver.ServiceInstancesQuery(nil, namespace, nil, nil, &gqlStatusType)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error", func(t *testing.T) {
		expected := errors.New("Test")

		namespace := "namespace"
		statusType := status.ServiceInstanceStatusTypeRunning
		gqlStatusType := gqlschema.InstanceStatusTypeRunning
		var resources []*v1beta1.ServiceInstance

		resourceGetter := servicecatalog.NewMockServiceInstanceService()
		resourceGetter.On("ListForStatus", namespace, pager.PagingParams{}, &statusType).Return(resources, expected).Once()
		defer resourceGetter.AssertExpectations(t)
		resolver := servicecatalog.NewServiceInstanceResolver(resourceGetter, nil, nil, nil, nil)

		_, err := resolver.ServiceInstancesQuery(nil, namespace, nil, nil, &gqlStatusType)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
	})
}

func TestServiceInstanceResolver_ServiceInstanceClusterServicePlanField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		planName := "name"
		expected := &gqlschema.ClusterServicePlan{
			Name: planName,
		}
		resource := &v1beta1.ClusterServicePlan{}
		resourceGetter := automock.NewClusterServicePlanGetter()
		resourceGetter.On("Find", planName).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLClusterServicePlanConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name: "test",
			PlanReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        planName,
				ClusterWide: true,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, resourceGetter, nil, nil, nil)
		resolver.SetClusterServicePlanConverter(converter)

		result, err := resolver.ServiceInstanceClusterServicePlanField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := automock.NewClusterServicePlanGetter()
		resourceGetter.On("Find", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name: "test",
			PlanReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: true,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, resourceGetter, nil, nil, nil)

		result, err := resolver.ServiceInstanceClusterServicePlanField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ServicePlanName not provided", func(t *testing.T) {
		parentObj := gqlschema.ServiceInstance{
			Name:          "test",
			PlanReference: nil,
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, nil, nil)

		result, err := resolver.ServiceInstanceClusterServicePlanField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := automock.NewClusterServicePlanGetter()
		resourceGetter.On("Find", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name: "test",
			PlanReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: true,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, resourceGetter, nil, nil, nil)

		result, err := resolver.ServiceInstanceClusterServicePlanField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestServiceInstanceResolver_ServiceInstanceClusterServiceClassField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ClusterServiceClass{
			Name: "Test",
		}
		name := "name"
		resource := &v1beta1.ClusterServiceClass{}
		resourceGetter := automock.NewClusterServiceClassListGetter()
		resourceGetter.On("Find", name).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLClusterServiceClassConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name: "test",
			ClassReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: true,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, resourceGetter, nil, nil)
		resolver.SetClusterServiceClassConverter(converter)

		result, err := resolver.ServiceInstanceClusterServiceClassField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		name := "name"
		resourceGetter := automock.NewClusterServiceClassListGetter()
		resourceGetter.On("Find", name).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name: "test",
			ClassReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: true,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, resourceGetter, nil, nil)

		result, err := resolver.ServiceInstanceClusterServiceClassField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ServiceClassName not provided", func(t *testing.T) {
		parentObj := gqlschema.ServiceInstance{
			Name:           "test",
			ClassReference: nil,
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, nil, nil)

		result, err := resolver.ServiceInstanceClusterServiceClassField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := automock.NewClusterServiceClassListGetter()
		resourceGetter.On("Find", name).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name: "test",
			ClassReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: true,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, resourceGetter, nil, nil)

		result, err := resolver.ServiceInstanceClusterServiceClassField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestServiceInstanceResolver_ServiceInstanceServicePlanField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ns := "ns"
		planName := "name"
		expected := &gqlschema.ServicePlan{
			Name:      planName,
			Namespace: ns,
		}
		resource := &v1beta1.ServicePlan{}
		resourceGetter := automock.NewServicePlanGetter()
		resourceGetter.On("Find", planName, ns).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLServicePlanConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:      "test",
			Namespace: ns,
			PlanReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        planName,
				ClusterWide: false,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, resourceGetter, nil)
		resolver.SetServicePlanConverter(converter)

		result, err := resolver.ServiceInstanceServicePlanField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		ns := "ns"
		name := "name"
		resourceGetter := automock.NewServicePlanGetter()
		resourceGetter.On("Find", name, ns).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:      "test",
			Namespace: ns,
			PlanReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: false,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceInstanceServicePlanField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ServicePlanName not provided", func(t *testing.T) {
		ns := "ns"
		parentObj := gqlschema.ServiceInstance{
			Name:          "test",
			Namespace:     ns,
			PlanReference: nil,
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, nil, nil)

		result, err := resolver.ServiceInstanceServicePlanField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		expectedErr := errors.New("Test")
		ns := "ns"
		name := "name"
		resourceGetter := automock.NewServicePlanGetter()
		resourceGetter.On("Find", name, ns).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:      "test",
			Namespace: ns,
			PlanReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: false,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, resourceGetter, nil)

		result, err := resolver.ServiceInstanceServicePlanField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestServiceInstanceResolver_ServiceInstanceServiceClassField(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expected := &gqlschema.ServiceClass{
			Name:      "Test",
			Namespace: "ns",
		}

		ns := "ns"
		name := "name"
		resource := &v1beta1.ServiceClass{}
		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("Find", name, ns).Return(resource, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		converter := automock.NewGQLServiceClassConverter()
		converter.On("ToGQL", resource).Return(expected, nil).Once()
		defer converter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:      "test",
			Namespace: ns,
			ClassReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: false,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, nil, resourceGetter)
		resolver.SetServiceClassConverter(converter)

		result, err := resolver.ServiceInstanceServiceClassField(nil, &parentObj)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		ns := "ns"
		name := "name"
		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("Find", name, ns).Return(nil, nil).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:      "test",
			Namespace: ns,
			ClassReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: false,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, nil, resourceGetter)

		result, err := resolver.ServiceInstanceServiceClassField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("ServiceClassName not provided", func(t *testing.T) {
		ns := "ns"
		parentObj := gqlschema.ServiceInstance{
			Name:           "test",
			Namespace:      ns,
			ClassReference: nil,
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, nil, nil)

		result, err := resolver.ServiceInstanceServiceClassField(nil, &parentObj)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		ns := "ns"
		expectedErr := errors.New("Test")
		name := "name"
		resourceGetter := automock.NewServiceClassListGetter()
		resourceGetter.On("Find", name, ns).Return(nil, expectedErr).Once()
		defer resourceGetter.AssertExpectations(t)

		parentObj := gqlschema.ServiceInstance{
			Name:      "test",
			Namespace: ns,
			ClassReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        name,
				ClusterWide: false,
			},
		}

		resolver := servicecatalog.NewServiceInstanceResolver(nil, nil, nil, nil, resourceGetter)

		result, err := resolver.ServiceInstanceServiceClassField(nil, &parentObj)

		assert.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}

func TestServiceInstanceResolver_ServiceInstanceBindableField(t *testing.T) {
	testErr := errors.New("Test")
	className := "className"
	planName := "planName"

	t.Run("ClusterWide", func(t *testing.T) {
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

			planGetter := automock.NewClusterServicePlanGetter()
			planGetter.On("Find", planName).Return(plan, tc.planErr).Once()
			classGetter := automock.NewClusterServiceClassListGetter()
			classGetter.On("Find", className).Return(class, tc.classErr).Once()
			instanceSvc := servicecatalog.NewMockServiceInstanceService()
			instanceSvc.On("IsBindableWithClusterRefs", class, plan).Return(tc.instanceBindable).Once()

			resolver := servicecatalog.NewServiceInstanceResolver(instanceSvc, planGetter, classGetter, nil, nil)

			var classReference *gqlschema.ServiceInstanceResourceRef
			if tc.className != nil {
				classReference = &gqlschema.ServiceInstanceResourceRef{
					Name:        *tc.className,
					ClusterWide: true,
				}
			}

			var planReference *gqlschema.ServiceInstanceResourceRef
			if tc.planName != nil {
				planReference = &gqlschema.ServiceInstanceResourceRef{
					Name:        *tc.planName,
					ClusterWide: true,
				}
			}

			parentObj := &gqlschema.ServiceInstance{
				Name:           "test",
				ClassReference: classReference,
				PlanReference:  planReference,
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
	})

	t.Run("Local", func(t *testing.T) {
		ns := "ns"

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
			class := &v1beta1.ServiceClass{}
			plan := &v1beta1.ServicePlan{}

			planGetter := automock.NewServicePlanGetter()
			planGetter.On("Find", planName, ns).Return(plan, tc.planErr).Once()
			classGetter := automock.NewServiceClassListGetter()
			classGetter.On("Find", className, ns).Return(class, tc.classErr).Once()
			instanceSvc := servicecatalog.NewMockServiceInstanceService()
			instanceSvc.On("IsBindableWithLocalRefs", class, plan).Return(tc.instanceBindable).Once()

			resolver := servicecatalog.NewServiceInstanceResolver(instanceSvc, nil, nil, planGetter, classGetter)

			var classReference *gqlschema.ServiceInstanceResourceRef
			if tc.className != nil {
				classReference = &gqlschema.ServiceInstanceResourceRef{
					Name:        *tc.className,
					ClusterWide: false,
				}
			}

			var planReference *gqlschema.ServiceInstanceResourceRef
			if tc.planName != nil {
				planReference = &gqlschema.ServiceInstanceResourceRef{
					Name:        *tc.planName,
					ClusterWide: false,
				}
			}

			parentObj := &gqlschema.ServiceInstance{
				Name:           "test",
				Namespace:      ns,
				ClassReference: classReference,
				PlanReference:  planReference,
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
	})
}

func TestServiceInstanceResolver_ServiceInstanceEventSubscription(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := servicecatalog.NewMockServiceInstanceService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalog.NewServiceInstanceResolver(svc, nil, nil, nil, nil)

		_, err := resolver.ServiceInstanceEventSubscription(ctx, "test")

		require.NoError(t, err)
		svc.AssertCalled(t, "Subscribe", mock.Anything)
	})

	t.Run("Unsubscribe after connection close", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), (-24 * time.Hour))
		cancel()

		svc := servicecatalog.NewMockServiceInstanceService()
		svc.On("Subscribe", mock.Anything).Once()
		svc.On("Unsubscribe", mock.Anything).Once()
		resolver := servicecatalog.NewServiceInstanceResolver(svc, nil, nil, nil, nil)

		channel, err := resolver.ServiceInstanceEventSubscription(ctx, "test")
		<-channel

		require.NoError(t, err)
		svc.AssertCalled(t, "Unsubscribe", mock.Anything)
	})
}
