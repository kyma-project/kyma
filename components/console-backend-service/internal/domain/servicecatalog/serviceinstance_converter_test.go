package servicecatalog

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestServiceInstanceConverter_ToGQL(t *testing.T) {
	var mockTimeStamp metav1.Time
	var zeroTimeStamp time.Time
	t.Run("All properties are given", func(t *testing.T) {
		converter := serviceInstanceConverter{}
		parameterSchema := map[string]interface{}{
			"first": "1",
			"second": map[string]interface{}{
				"value": "2",
			},
		}

		parameterSchemaBytes, err := json.Marshal(parameterSchema)
		require.NoError(t, err)

		parameterSchemaJSON := new(gqlschema.JSON)
		err = parameterSchemaJSON.UnmarshalGQL(parameterSchema)
		require.NoError(t, err)

		testClassName := "testClass"
		testPlanName := "testPlan"

		in := v1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				Namespace:         "Namespace",
				UID:               types.UID("uid"),
				CreationTimestamp: mockTimeStamp,
				Annotations: map[string]string{
					"tags": "test1,test2",
				},
			},
			Spec: v1beta1.ServiceInstanceSpec{
				ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
					Name: testClassName,
				},
				ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
					Name: testPlanName,
				},
				Parameters: &runtime.RawExtension{Raw: parameterSchemaBytes},
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

		expected := gqlschema.ServiceInstance{
			Name:      "exampleName",
			Namespace: "Namespace",
			ClassReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        testClassName,
				ClusterWide: true,
			},
			PlanReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        testPlanName,
				ClusterWide: true,
			},
			PlanSpec:          parameterSchemaJSON,
			Labels:            []string{"test1", "test2"},
			CreationTimestamp: zeroTimeStamp,
			Status: gqlschema.ServiceInstanceStatus{
				Type:    gqlschema.InstanceStatusTypeRunning,
				Reason:  "Testing",
				Message: "Working",
			},
		}

		result, err := converter.ToGQL(&in)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Parameters not provided", func(t *testing.T) {
		converter := serviceInstanceConverter{}

		in := v1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				Namespace:         "Namespace",
				UID:               types.UID("uid"),
				CreationTimestamp: mockTimeStamp,
			},
			Spec: v1beta1.ServiceInstanceSpec{
				ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
					Name: "testClass",
				},
				ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
					Name: "testPlan",
				},
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

		testClassName := "testClass"
		testPlanName := "testPlan"
		expected := gqlschema.ServiceInstance{
			Name:      "exampleName",
			Namespace: "Namespace",
			ClassReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        testClassName,
				ClusterWide: true,
			},
			PlanReference: &gqlschema.ServiceInstanceResourceRef{
				Name:        testPlanName,
				ClusterWide: true,
			},
			Labels:            []string{},
			CreationTimestamp: zeroTimeStamp,
			Status: gqlschema.ServiceInstanceStatus{
				Type:    gqlschema.InstanceStatusTypeRunning,
				Reason:  "Testing",
				Message: "Working",
			},
		}

		result, err := converter.ToGQL(&in)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &serviceInstanceConverter{}
		item, err := converter.ToGQL(&v1beta1.ServiceInstance{})

		expected := &gqlschema.ServiceInstance{
			Labels: []string{},
			Status: gqlschema.ServiceInstanceStatus{
				Type: gqlschema.InstanceStatusTypePending,
			},
		}

		require.NoError(t, err)
		assert.Equal(t, expected, item)
	})

	t.Run("Empty properties", func(t *testing.T) {
		converter := &serviceInstanceConverter{}
		item, err := converter.ToGQL(&v1beta1.ServiceInstance{
			Status: v1beta1.ServiceInstanceStatus{
				InProgressProperties: &v1beta1.ServiceInstancePropertiesState{},
				ExternalProperties:   &v1beta1.ServiceInstancePropertiesState{},
			},
		})

		expected := &gqlschema.ServiceInstance{
			Labels: []string{},
			Status: gqlschema.ServiceInstanceStatus{
				Type: gqlschema.InstanceStatusTypePending,
			},
		}

		require.NoError(t, err)
		assert.Equal(t, expected, item)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &serviceInstanceConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestServiceInstanceConverter_GQLCreateInputToInstanceCreateParameters(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		JSON := gqlschema.JSON{
			"key1": "val1",
			"key2": "val2",
		}
		input := gqlschema.ServiceInstanceCreateInput{
			Name:            "name",
			Labels:          []string{"test", "label"},
			ParameterSchema: &JSON,
			ClassRef: gqlschema.ServiceInstanceCreateInputResourceRef{
				ExternalName: "className",
				ClusterWide:  true,
			},
			PlanRef: gqlschema.ServiceInstanceCreateInputResourceRef{
				ExternalName: "planName",
				ClusterWide:  true,
			},
		}
		expected := &serviceInstanceCreateParameters{
			Name:      "name",
			Namespace: "ns",
			Labels:    []string{"test", "label"},
			Schema:    JSON,
			ClassRef: instanceCreateResourceRef{
				ExternalName: "className",
				ClusterWide:  true,
			},
			PlanRef: instanceCreateResourceRef{
				ExternalName: "planName",
				ClusterWide:  true,
			},
		}
		converter := serviceInstanceConverter{}

		result := converter.GQLCreateInputToInstanceCreateParameters(&input, "ns")

		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := serviceInstanceConverter{}
		result := converter.GQLCreateInputToInstanceCreateParameters(&gqlschema.ServiceInstanceCreateInput{}, "")

		assert.Empty(t, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := serviceInstanceConverter{}
		result := converter.GQLCreateInputToInstanceCreateParameters(nil, "")

		assert.Nil(t, result)
	})
}

func TestServiceInstanceConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instances := []*v1beta1.ServiceInstance{
			fixServiceInstance(t),
			fixServiceInstance(t),
		}

		converter := serviceInstanceConverter{}
		result, err := converter.ToGQLs(instances)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "exampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var instances []*v1beta1.ServiceInstance

		converter := serviceInstanceConverter{}
		result, err := converter.ToGQLs(instances)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		instances := []*v1beta1.ServiceInstance{
			nil,
			fixServiceInstance(t),
			nil,
		}

		converter := serviceInstanceConverter{}
		result, err := converter.ToGQLs(instances)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "exampleName", result[0].Name)
	})
}

func TestServiceInstanceConverter_ServiceStatusToGQLStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := status.ServiceInstanceStatus{
			Type:    status.ServiceInstanceStatusTypeRunning,
			Reason:  "Testing",
			Message: "Working",
		}
		expected := gqlschema.ServiceInstanceStatus{
			Type:    gqlschema.InstanceStatusTypeRunning,
			Reason:  "Testing",
			Message: "Working",
		}

		converter := serviceInstanceConverter{}
		result := converter.ServiceStatusToGQLStatus(s)

		assert.Equal(t, expected, result)
	})
}

func TestServiceInstanceConverter_GQLStatusToServiceStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s := gqlschema.ServiceInstanceStatus{
			Type:    gqlschema.InstanceStatusTypeRunning,
			Reason:  "Testing",
			Message: "Working",
		}
		expected := status.ServiceInstanceStatus{
			Type:    status.ServiceInstanceStatusTypeRunning,
			Reason:  "Testing",
			Message: "Working",
		}

		converter := serviceInstanceConverter{}
		result := converter.GQLStatusToServiceStatus(&s)

		assert.Equal(t, &expected, result)
	})
}

func TestServiceInstanceConverter_ServiceStatusToGQLStatusWithConvert(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instance := v1beta1.ServiceInstance{
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
		expected := gqlschema.ServiceInstanceStatus{
			Type:    gqlschema.InstanceStatusTypeRunning,
			Reason:  "Testing",
			Message: "Working",
		}

		converter := serviceInstanceConverter{}
		result := converter.ServiceStatusToGQLStatus(converter.extractor.Status(instance))

		assert.Equal(t, expected, result)
	})
}

func fixServiceInstance(t require.TestingT) *v1beta1.ServiceInstance {
	var mockTimeStamp metav1.Time
	rawMap := map[string]interface{}{
		"labels": []string{"test1", "test2"},
	}
	raw, err := json.Marshal(rawMap)
	require.NoError(t, err)

	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "exampleName",
			Namespace:         "Namespace",
			UID:               types.UID("uid"),
			CreationTimestamp: mockTimeStamp,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			ClusterServiceClassRef: &v1beta1.ClusterObjectReference{
				Name: "testClass",
			},
			ClusterServicePlanRef: &v1beta1.ClusterObjectReference{
				Name: "testPlan",
			},
			Parameters: &runtime.RawExtension{Raw: raw},
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
}
