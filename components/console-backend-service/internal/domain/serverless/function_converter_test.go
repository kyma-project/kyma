package serverless

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestFunctionConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedName := "expectedName"
		expectedNamespace := "expectedNamespace"
		expectedUID := "expectedUID"
		expectedSource := "expectedSource"
		expectedDependencies := "expectedDependencies"
		expectedLabels := map[string]string{"foo": "bar"}

		function := fixFunction(expectedName, expectedNamespace, expectedUID, expectedSource, expectedDependencies, expectedLabels)
		gqlFunction := fixGQLFunction(expectedName, expectedNamespace, expectedUID, expectedSource, expectedDependencies, expectedLabels)

		converter := newFunctionConverter()
		result, err := converter.ToGQL(function)

		require.NoError(t, err)
		assert.Equal(t, &gqlFunction, result)
	})

	t.Run("Empty", func(t *testing.T) {
		function := &v1alpha1.Function{}

		expected := gqlschema.Function{
			Status: gqlschema.FunctionStatus{
				Phase: gqlschema.FunctionPhaseTypeInitializing,
			},
		}

		converter := newFunctionConverter()
		result, err := converter.ToGQL(function)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := newFunctionConverter()
		result, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFunctionConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedName := "expectedName"
		expectedNamespace := "expectedNamespace"
		expectedUID := "expectedUID"
		expectedSource := "expectedSource"
		expectedDependencies := "expectedDependencies"
		expectedLabels := map[string]string{"foo": "bar"}

		function1 := fixFunction(expectedName, expectedNamespace, expectedUID, expectedSource, expectedDependencies, expectedLabels)
		function2 := fixFunction(expectedName, expectedNamespace, expectedUID, expectedSource, expectedDependencies, expectedLabels)
		functions := []*v1alpha1.Function{function1, function2}

		gqlFunction1 := fixGQLFunction(expectedName, expectedNamespace, expectedUID, expectedSource, expectedDependencies, expectedLabels)
		gqlFunction2 := fixGQLFunction(expectedName, expectedNamespace, expectedUID, expectedSource, expectedDependencies, expectedLabels)
		gqlFunctions := []gqlschema.Function{gqlFunction1, gqlFunction2}

		converter := newFunctionConverter()
		result, err := converter.ToGQLs(functions)

		require.NoError(t, err)
		assert.Equal(t, gqlFunctions, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := newFunctionConverter()
		result, err := converter.ToGQLs([]*v1alpha1.Function{})

		require.NoError(t, err)
		assert.Equal(t, []gqlschema.Function(nil), result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := newFunctionConverter()
		result, err := converter.ToGQLs(nil)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestFunctionConverter_ToFunction(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedName := "expectedName"
		expectedNamespace := "expectedNamespace"
		expectedSource := "expectedSource"
		expectedDependencies := "expectedDependencies"
		expectedLabels := map[string]string{"foo": "bar"}

		function := fixFunction(expectedName, expectedNamespace, "", expectedSource, expectedDependencies, expectedLabels)
		gqlMutationInput := fixGQLMutationInput(expectedSource, expectedDependencies, expectedLabels)
		gqlMutationInput.Env = []gqlschema.FunctionEnvInput{
			{
				Name:  "foo",
				Value: "bar",
			},
		}

		converter := newFunctionConverter()
		result, err := converter.ToFunction(expectedName, expectedNamespace, gqlMutationInput)

		require.NoError(t, err)
		assert.Equal(t, function, result)
	})
}

func TestFunctionConverter_Env(t *testing.T) {
	converter := newFunctionConverter()
	kubernetesEnv := []v1.EnvVar{
		{
			Name:  "foo",
			Value: "bar",
		},
		{
			Name:  "bar",
			Value: "foo",
		},
		{
			Name: "configMap",
			ValueFrom: &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "configMap",
					},
					Key:      "key",
					Optional: new(bool),
				},
			},
		},
		{
			Name: "secret",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "secret",
					},
					Key:      "key",
					Optional: nil,
				},
			},
		},
	}
	gqlEnv := []gqlschema.FunctionEnv{
		{
			Name:  "foo",
			Value: "bar",
		},
		{
			Name:  "bar",
			Value: "foo",
		},
		{
			Name: "configMap",
			ValueFrom: &gqlschema.FunctionEnvValueFrom{
				Type:     gqlschema.FunctionEnvValueFromTypeConfigMap,
				Name:     "configMap",
				Key:      "key",
				Optional: new(bool),
			},
		},
		{
			Name: "secret",
			ValueFrom: &gqlschema.FunctionEnvValueFrom{
				Type: gqlschema.FunctionEnvValueFromTypeSecret,
				Name: "secret",
				Key:  "key",
			},
		},
	}
	gqlEnvInput := []gqlschema.FunctionEnvInput{
		{
			Name:  "foo",
			Value: "bar",
		},
		{
			Name:  "bar",
			Value: "foo",
		},
		{
			Name: "configMap",
			ValueFrom: &gqlschema.FunctionEnvValueFromInput{
				Type:     gqlschema.FunctionEnvValueFromTypeConfigMap,
				Name:     "configMap",
				Key:      "key",
				Optional: new(bool),
			},
		},
		{
			Name: "secret",
			ValueFrom: &gqlschema.FunctionEnvValueFromInput{
				Type: gqlschema.FunctionEnvValueFromTypeSecret,
				Name: "secret",
				Key:  "key",
			},
		},
	}

	t.Run("Success - toGQLEnv", func(t *testing.T) {
		result := converter.toGQLEnv(kubernetesEnv)
		assert.ElementsMatch(t, gqlEnv, result)
	})

	t.Run("Empty - toGQLEnv", func(t *testing.T) {
		result := converter.toGQLEnv([]v1.EnvVar{})
		assert.ElementsMatch(t, []gqlschema.FunctionEnv{}, result)
	})

	t.Run("Success - fromGQLEnv", func(t *testing.T) {
		result := converter.fromGQLEnv(gqlEnvInput)
		assert.ElementsMatch(t, kubernetesEnv, result)
	})

	t.Run("Empty - fromGQLEnv", func(t *testing.T) {
		result := converter.fromGQLEnv([]gqlschema.FunctionEnvInput{})
		assert.ElementsMatch(t, []v1.EnvVar{}, result)
	})
}

func TestFunctionConverter_Replicas(t *testing.T) {
	converter := newFunctionConverter()
	min := 5
	max := 7
	int32Min := int32(min)
	int32Max := int32(max)

	gqlReplicas := gqlschema.FunctionReplicas{
		Min: &min,
		Max: &max,
	}
	gqlReplicasInput := gqlschema.FunctionReplicasInput{
		Min: &min,
		Max: &max,
	}

	t.Run("Success - toGQLReplicas", func(t *testing.T) {
		result := converter.toGQLReplicas(&int32Min, &int32Max)
		assert.Equal(t, gqlReplicas, result)
	})

	t.Run("Empty - toGQLReplicas", func(t *testing.T) {
		result := converter.toGQLReplicas(nil, nil)
		assert.ElementsMatch(t, gqlschema.FunctionReplicas{}, result)
	})

	t.Run("Success - fromGQLReplicas", func(t *testing.T) {
		k8sMin, k8sMax := converter.fromGQLReplicas(gqlReplicasInput)
		assert.Equal(t, int32Min, *k8sMin)
		assert.Equal(t, int32Max, *k8sMax)
	})

	t.Run("Empty - fromGQLReplicas", func(t *testing.T) {
		k8sMin, k8sMax := converter.fromGQLReplicas(gqlschema.FunctionReplicasInput{})
		assert.Nil(t, k8sMin)
		assert.Nil(t, k8sMax)
	})
}

func TestFunctionConverter_Resources(t *testing.T) {
	converter := newFunctionConverter()

	t.Run("Success - toGQLResources", func(t *testing.T) {
		cpu := resource.NewQuantity(3006477108, resource.DecimalSI)
		resources := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: *cpu,
			},
		}
		gqlCPU := "3006477108"
		gqlResources := gqlschema.FunctionResources{
			Requests: gqlschema.ResourceValues{
				CPU: &gqlCPU,
			},
		}

		result := converter.toGQLResources(resources)
		assert.Equal(t, gqlResources, result)
	})

	t.Run("Empty - toGQLResources", func(t *testing.T) {
		result := converter.toGQLResources(v1.ResourceRequirements{})
		assert.ElementsMatch(t, gqlschema.FunctionResources{}, result)
	})

	t.Run("Success - fromGQLResources", func(t *testing.T) {
		gqlCPU := "3006477108"
		gqlResources := gqlschema.FunctionResourcesInput{
			Requests: gqlschema.ResourceValuesInput{
				CPU: &gqlCPU,
			},
		}

		cpuParsed, err := resource.ParseQuantity(gqlCPU)
		assert.NoError(t, err)
		resources := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: cpuParsed,
			},
		}

		result, errs := converter.fromGQLResources(gqlResources)
		assert.Len(t, errs, 0)
		assert.Equal(t, resources, result)
	})

	t.Run("Error - fromGQLResources", func(t *testing.T) {
		gqlCPU := "pico-bello"
		gqlResources := gqlschema.FunctionResourcesInput{
			Requests: gqlschema.ResourceValuesInput{
				CPU: &gqlCPU,
			},
		}

		result, errs := converter.fromGQLResources(gqlResources)
		assert.Len(t, errs, 1)
		assert.Equal(t, v1.ResourceRequirements{}, result)
	})

	t.Run("Empty - fromGQLResources", func(t *testing.T) {
		result, errs := converter.fromGQLResources(gqlschema.FunctionResourcesInput{})
		assert.Len(t, errs, 0)
		assert.Equal(t, v1.ResourceRequirements{}, result)
	})
}

func TestFunctionConverter_Status(t *testing.T) {
	converter := newFunctionConverter()

	t.Run("Initializing", func(t *testing.T) {
		status := v1alpha1.FunctionStatus{Conditions: []v1alpha1.Condition{}}
		gqlStatus := converter.getStatus(status)

		assert.Equal(t, gqlschema.FunctionPhaseTypeInitializing, gqlStatus.Phase)
	})

	t.Run("Building", func(t *testing.T) {
		status := v1alpha1.FunctionStatus{Conditions: []v1alpha1.Condition{
			{
				Type:   v1alpha1.ConditionConfigurationReady,
				Status: v1.ConditionTrue,
			},
		}}
		gqlStatus := converter.getStatus(status)

		assert.Equal(t, gqlschema.FunctionPhaseTypeBuilding, gqlStatus.Phase)
	})

	t.Run("Deploying", func(t *testing.T) {
		status := v1alpha1.FunctionStatus{Conditions: []v1alpha1.Condition{
			{
				Type:   v1alpha1.ConditionConfigurationReady,
				Status: v1.ConditionTrue,
			},
			{
				Type:   v1alpha1.ConditionBuildReady,
				Status: v1.ConditionTrue,
			},
		}}
		gqlStatus := converter.getStatus(status)

		assert.Equal(t, gqlschema.FunctionPhaseTypeDeploying, gqlStatus.Phase)
	})

	t.Run("Running", func(t *testing.T) {
		status := v1alpha1.FunctionStatus{Conditions: []v1alpha1.Condition{
			{
				Type:   v1alpha1.ConditionConfigurationReady,
				Status: v1.ConditionTrue,
			},
			{
				Type:   v1alpha1.ConditionBuildReady,
				Status: v1.ConditionTrue,
			},
			{
				Type:   v1alpha1.ConditionRunning,
				Status: v1.ConditionTrue,
			},
		}}
		gqlStatus := converter.getStatus(status)

		assert.Equal(t, gqlschema.FunctionPhaseTypeRunning, gqlStatus.Phase)
	})

	t.Run("Failed", func(t *testing.T) {
		errorMessage := "Error"

		status := v1alpha1.FunctionStatus{Conditions: []v1alpha1.Condition{
			{
				Type:    v1alpha1.ConditionConfigurationReady,
				Status:  v1.ConditionFalse,
				Message: errorMessage,
			},
			{
				Type:    v1alpha1.ConditionBuildReady,
				Status:  v1.ConditionTrue,
				Message: errorMessage,
			},
			{
				Type:    v1alpha1.ConditionRunning,
				Status:  v1.ConditionUnknown,
				Message: errorMessage,
			},
		}}
		gqlStatus := converter.getStatus(status)

		// Config Failed
		reason := gqlschema.FunctionReasonTypeConfig
		expected := gqlschema.FunctionStatus{
			Phase:   gqlschema.FunctionPhaseTypeFailed,
			Reason:  &reason,
			Message: &errorMessage,
		}

		assert.Equal(t, expected, gqlStatus)

		// Job Failed
		status.Conditions[0].Status = v1.ConditionTrue
		status.Conditions[1].Status = v1.ConditionFalse

		reason = gqlschema.FunctionReasonTypeJob
		expected.Reason = &reason

		gqlStatus = converter.getStatus(status)
		assert.Equal(t, expected, gqlStatus)

		// Service Failed
		status.Conditions[1].Status = v1.ConditionTrue
		status.Conditions[2].Status = v1.ConditionFalse

		reason = gqlschema.FunctionReasonTypeService
		expected.Reason = &reason

		gqlStatus = converter.getStatus(status)
		assert.Equal(t, expected, gqlStatus)
	})

	t.Run("NewRevisionError", func(t *testing.T) {
		errorMessage := "Error"

		status := v1alpha1.FunctionStatus{Conditions: []v1alpha1.Condition{
			{
				Type:    v1alpha1.ConditionConfigurationReady,
				Status:  v1.ConditionFalse,
				Message: errorMessage,
			},
			{
				Type:    v1alpha1.ConditionBuildReady,
				Status:  v1.ConditionTrue,
				Message: errorMessage,
			},
			{
				Type:   v1alpha1.ConditionRunning,
				Status: v1.ConditionTrue,
			},
		}}
		gqlStatus := converter.getStatus(status)

		// Config Failed
		reason := gqlschema.FunctionReasonTypeConfig
		expected := gqlschema.FunctionStatus{
			Phase:   gqlschema.FunctionPhaseTypeNewRevisionError,
			Reason:  &reason,
			Message: &errorMessage,
		}

		assert.Equal(t, expected, gqlStatus)

		// Job Failed
		status.Conditions[0].Status = v1.ConditionTrue
		status.Conditions[1].Status = v1.ConditionFalse

		reason = gqlschema.FunctionReasonTypeJob
		expected.Reason = &reason

		gqlStatus = converter.getStatus(status)
		assert.Equal(t, expected, gqlStatus)
	})
}
