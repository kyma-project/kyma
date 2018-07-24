package servicecatalog

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestInstanceConverter_ToGQL(t *testing.T) {
	var mockTimeStamp metav1.Time
	var zeroTimeStamp time.Time
	t.Run("All properties are given", func(t *testing.T) {
		converter := instanceConverter{}

		in := v1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				Namespace:         "Environment",
				UID:               types.UID("uid"),
				CreationTimestamp: mockTimeStamp,
				Annotations: map[string]string{
					"tags": "test1,test2",
				},
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
			Name:              "exampleName",
			Environment:       "Environment",
			ServiceClassName:  &testClassName,
			ServicePlanName:   &testPlanName,
			Labels:            []string{"test1", "test2"},
			CreationTimestamp: zeroTimeStamp,
			Status: gqlschema.ServiceInstanceStatus{
				Type:    gqlschema.InstanceStatusTypeRunning,
				Reason:  "Testing",
				Message: "Working",
			},
		}

		result := converter.ToGQL(&in)

		assert.Equal(t, &expected, result)
	})

	t.Run("Parameters not provided", func(t *testing.T) {
		converter := instanceConverter{}

		in := v1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				Namespace:         "Environment",
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
			Name:              "exampleName",
			Environment:       "Environment",
			ServiceClassName:  &testClassName,
			ServicePlanName:   &testPlanName,
			Labels:            []string{},
			CreationTimestamp: zeroTimeStamp,
			Status: gqlschema.ServiceInstanceStatus{
				Type:    gqlschema.InstanceStatusTypeRunning,
				Reason:  "Testing",
				Message: "Working",
			},
		}

		result := converter.ToGQL(&in)

		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &instanceConverter{}
		converter.ToGQL(&v1beta1.ServiceInstance{})
	})

	t.Run("Empty properties", func(t *testing.T) {
		converter := &instanceConverter{}
		converter.ToGQL(&v1beta1.ServiceInstance{
			Status: v1beta1.ServiceInstanceStatus{
				InProgressProperties: &v1beta1.ServiceInstancePropertiesState{},
				ExternalProperties:   &v1beta1.ServiceInstancePropertiesState{},
			},
		})
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &instanceConverter{}
		item := converter.ToGQL(nil)
		assert.Nil(t, item)
	})
}

func TestInstanceConverter_GQLCreateInputToInstanceCreateParameters(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		JSON := gqlschema.JSON{
			"key1": "val1",
			"key2": "val2",
		}
		input := gqlschema.ServiceInstanceCreateInput{
			Name:                     "name",
			Environment:              "environment",
			Labels:                   []string{"test", "label"},
			ParameterSchema:          &JSON,
			ExternalServiceClassName: "className",
			ExternalPlanName:         "planName",
		}
		expected := &instanceCreateParameters{
			Name:      "name",
			Namespace: "environment",
			Labels:    []string{"test", "label"},
			Schema:    JSON,
			ExternalServiceClassName: "className",
			ExternalServicePlanName:  "planName",
		}
		converter := instanceConverter{}

		result := converter.GQLCreateInputToInstanceCreateParameters(&input)

		assert.Equal(t, expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := instanceConverter{}
		result := converter.GQLCreateInputToInstanceCreateParameters(&gqlschema.ServiceInstanceCreateInput{})

		assert.Empty(t, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := instanceConverter{}
		result := converter.GQLCreateInputToInstanceCreateParameters(nil)

		assert.Nil(t, result)
	})
}

func TestInstanceConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		instances := []*v1beta1.ServiceInstance{
			fixServiceInstance(t),
			fixServiceInstance(t),
		}

		converter := instanceConverter{}
		result := converter.ToGQLs(instances)

		assert.Len(t, result, 2)
		assert.Equal(t, "exampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var instances []*v1beta1.ServiceInstance

		converter := instanceConverter{}
		result := converter.ToGQLs(instances)

		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		instances := []*v1beta1.ServiceInstance{
			nil,
			fixServiceInstance(t),
			nil,
		}

		converter := instanceConverter{}
		result := converter.ToGQLs(instances)

		assert.Len(t, result, 1)
		assert.Equal(t, "exampleName", result[0].Name)
	})
}

func TestInstanceConverter_ServiceStatusToGQLStatus(t *testing.T) {
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

		converter := instanceConverter{}
		result := converter.ServiceStatusToGQLStatus(&s)

		assert.Equal(t, &expected, result)
	})
}

func TestInstanceConverter_GQLStatusToServiceStatus(t *testing.T) {
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

		converter := instanceConverter{}
		result := converter.GQLStatusToServiceStatus(&s)

		assert.Equal(t, &expected, result)
	})
}

func TestInstanceConverter_ServiceStatusToGQLStatusWithConvert(t *testing.T) {
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

		converter := instanceConverter{}
		result := converter.ServiceStatusToGQLStatus(converter.extractor.Status(&instance))

		assert.Equal(t, &expected, result)
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
			Namespace:         "Environment",
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
