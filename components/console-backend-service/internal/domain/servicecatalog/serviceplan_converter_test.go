package servicecatalog

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestServicePlanConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := servicePlanConverter{}
		metadata := map[string]string{
			"displayName": "ExampleDisplayName",
		}

		metadataBytes, err := json.Marshal(metadata)
		assert.Nil(t, err)

		parameterSchema := map[string]interface{}{
			"properties": map[string]interface{}{
				"field": "value",
			},
		}

		parameterSchemaBytes, err := json.Marshal(parameterSchema)
		encodedParameterSchemaBytes := make([]byte, base64.StdEncoding.EncodedLen(len(parameterSchemaBytes)))
		base64.StdEncoding.Encode(encodedParameterSchemaBytes, parameterSchemaBytes)
		assert.Nil(t, err)

		parameterSchemaJSON := new(gqlschema.JSON)
		err = parameterSchemaJSON.UnmarshalGQL(parameterSchema)
		assert.Nil(t, err)

		clusterServicePlan := v1beta1.ServicePlan{
			Spec: v1beta1.ServicePlanSpec{
				CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
					ExternalMetadata: &runtime.RawExtension{Raw: metadataBytes},
					ExternalName:     "ExampleExternalName",
					InstanceCreateParameterSchema: &runtime.RawExtension{
						Raw: encodedParameterSchemaBytes,
					},
					ServiceBindingCreateParameterSchema: &runtime.RawExtension{
						Raw: encodedParameterSchemaBytes,
					},
				},
				ServiceClassRef: v1beta1.LocalObjectReference{
					Name: "serviceClassRef",
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "ExampleName",
				UID:  types.UID("uid"),
			},
		}
		displayName := "ExampleDisplayName"
		expected := gqlschema.ServicePlan{
			Name:                          "ExampleName",
			RelatedServiceClassName:       "serviceClassRef",
			DisplayName:                   &displayName,
			ExternalName:                  "ExampleExternalName",
			InstanceCreateParameterSchema: parameterSchemaJSON,
			BindingCreateParameterSchema:  parameterSchemaJSON,
		}

		result, err := converter.ToGQL(&clusterServicePlan)
		assert.Nil(t, err)

		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &servicePlanConverter{}
		_, err := converter.ToGQL(&v1beta1.ServicePlan{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &servicePlanConverter{}
		item, err := converter.ToGQL(nil)
		require.NoError(t, err)
		assert.Nil(t, item)
	})

	t.Run("CreateParameterSchema with properties", func(t *testing.T) {
		converter := &servicePlanConverter{}
		parameterSchema := map[string]interface{}{
			"additionalProperties": false,
			"properties": map[string]interface{}{
				"field": "value",
			},
		}

		parameterSchemaJSON := new(gqlschema.JSON)
		err := parameterSchemaJSON.UnmarshalGQL(parameterSchema)
		assert.Nil(t, err)

		clusterServicePlan := fixServicePlan(t, parameterSchema)
		displayName := "ExampleDisplayName"
		expected := gqlschema.ServicePlan{
			Name:                          "ExampleName",
			RelatedServiceClassName:       "serviceClassRef",
			DisplayName:                   &displayName,
			ExternalName:                  "ExampleExternalName",
			InstanceCreateParameterSchema: parameterSchemaJSON,
			BindingCreateParameterSchema:  parameterSchemaJSON,
		}

		result, err := converter.ToGQL(clusterServicePlan)
		assert.Nil(t, err)

		assert.Equal(t, &expected, result)
	})

	t.Run("CreateParameterSchema with ref", func(t *testing.T) {
		converter := &servicePlanConverter{}
		parameterSchema := map[string]interface{}{
			"additionalProperties": false,
			"$ref":                 "reference",
		}

		parameterSchemaJSON := new(gqlschema.JSON)
		err := parameterSchemaJSON.UnmarshalGQL(parameterSchema)
		assert.Nil(t, err)

		clusterServicePlan := fixServicePlan(t, parameterSchema)
		displayName := "ExampleDisplayName"
		expected := gqlschema.ServicePlan{
			Name:                          "ExampleName",
			RelatedServiceClassName:       "serviceClassRef",
			DisplayName:                   &displayName,
			ExternalName:                  "ExampleExternalName",
			InstanceCreateParameterSchema: parameterSchemaJSON,
			BindingCreateParameterSchema:  parameterSchemaJSON,
		}

		result, err := converter.ToGQL(clusterServicePlan)
		assert.Nil(t, err)

		assert.Equal(t, &expected, result)
	})

	t.Run("CreateParameterSchema with empty properties", func(t *testing.T) {
		converter := &servicePlanConverter{}
		parameterSchema := map[string]interface{}{
			"additionalProperties": false,
			"properties":           map[string]interface{}{},
		}

		parameterSchemaJSON := new(gqlschema.JSON)
		err := parameterSchemaJSON.UnmarshalGQL(parameterSchema)
		assert.Nil(t, err)

		clusterServicePlan := fixServicePlan(t, parameterSchema)
		displayName := "ExampleDisplayName"
		expected := gqlschema.ServicePlan{
			Name:                          "ExampleName",
			RelatedServiceClassName:       "serviceClassRef",
			DisplayName:                   &displayName,
			ExternalName:                  "ExampleExternalName",
			InstanceCreateParameterSchema: nil,
			BindingCreateParameterSchema:  nil,
		}

		result, err := converter.ToGQL(clusterServicePlan)
		assert.Nil(t, err)

		assert.Equal(t, &expected, result)
	})

	t.Run("CreateParameterSchema with empty ref", func(t *testing.T) {
		converter := &servicePlanConverter{}
		parameterSchema := map[string]interface{}{
			"additionalProperties": false,
			"$ref":                 "",
		}

		parameterSchemaJSON := new(gqlschema.JSON)
		err := parameterSchemaJSON.UnmarshalGQL(parameterSchema)
		assert.Nil(t, err)

		clusterServicePlan := fixServicePlan(t, parameterSchema)
		displayName := "ExampleDisplayName"
		expected := gqlschema.ServicePlan{
			Name:                          "ExampleName",
			RelatedServiceClassName:       "serviceClassRef",
			DisplayName:                   &displayName,
			ExternalName:                  "ExampleExternalName",
			InstanceCreateParameterSchema: nil,
			BindingCreateParameterSchema:  nil,
		}

		result, err := converter.ToGQL(clusterServicePlan)
		assert.Nil(t, err)

		assert.Equal(t, &expected, result)
	})

	t.Run("CreateParameterSchema with empty properties and ref", func(t *testing.T) {
		converter := &servicePlanConverter{}
		parameterSchema := map[string]interface{}{
			"additionalProperties": false,
			"properties":           map[string]interface{}{},
			"$ref":                 "",
		}

		parameterSchemaJSON := new(gqlschema.JSON)
		err := parameterSchemaJSON.UnmarshalGQL(parameterSchema)
		assert.Nil(t, err)

		clusterServicePlan := fixServicePlan(t, parameterSchema)
		displayName := "ExampleDisplayName"
		expected := gqlschema.ServicePlan{
			Name:                          "ExampleName",
			RelatedServiceClassName:       "serviceClassRef",
			DisplayName:                   &displayName,
			ExternalName:                  "ExampleExternalName",
			InstanceCreateParameterSchema: nil,
			BindingCreateParameterSchema:  nil,
		}

		result, err := converter.ToGQL(clusterServicePlan)
		assert.Nil(t, err)

		assert.Equal(t, &expected, result)
	})
}

func TestServicePlanConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		parameterSchema := map[string]interface{}{
			"properties": map[string]interface{}{
				"field": "value",
			},
		}

		plans := []*v1beta1.ServicePlan{
			fixServicePlan(t, parameterSchema),
			fixServicePlan(t, parameterSchema),
		}

		converter := servicePlanConverter{}
		result, err := converter.ToGQLs(plans)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ExampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var plans []*v1beta1.ServicePlan

		converter := servicePlanConverter{}
		result, err := converter.ToGQLs(plans)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		parameterSchema := map[string]interface{}{
			"first": "1",
			"second": map[string]interface{}{
				"value": "2",
			},
		}

		plans := []*v1beta1.ServicePlan{
			nil,
			fixServicePlan(t, parameterSchema),
			nil,
		}

		converter := servicePlanConverter{}
		result, err := converter.ToGQLs(plans)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "ExampleName", result[0].Name)
	})
}

func fixServicePlan(t require.TestingT, parameterSchema map[string]interface{}) *v1beta1.ServicePlan {
	metadata := map[string]string{
		"displayName": "ExampleDisplayName",
	}

	metadataBytes, err := json.Marshal(metadata)
	require.NoError(t, err)

	parameterSchemaBytes, err := json.Marshal(parameterSchema)
	encodedParameterSchemaBytes := make([]byte, base64.StdEncoding.EncodedLen(len(parameterSchemaBytes)))
	base64.StdEncoding.Encode(encodedParameterSchemaBytes, parameterSchemaBytes)
	require.NoError(t, err)

	return &v1beta1.ServicePlan{
		Spec: v1beta1.ServicePlanSpec{
			CommonServicePlanSpec: v1beta1.CommonServicePlanSpec{
				ExternalMetadata: &runtime.RawExtension{Raw: metadataBytes},
				ExternalName:     "ExampleExternalName",
				InstanceCreateParameterSchema: &runtime.RawExtension{
					Raw: encodedParameterSchemaBytes,
				},
				ServiceBindingCreateParameterSchema: &runtime.RawExtension{
					Raw: encodedParameterSchemaBytes,
				},
			},
			ServiceClassRef: v1beta1.LocalObjectReference{
				Name: "serviceClassRef",
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ExampleName",
			UID:  types.UID("uid"),
		},
	}
}
