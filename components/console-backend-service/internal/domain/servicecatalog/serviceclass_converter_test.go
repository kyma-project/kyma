package servicecatalog

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestServiceClassConverter_ToGQL(t *testing.T) {
	var mockTimeStamp metav1.Time
	var zeroTimeStamp time.Time
	t.Run("All properties are given", func(t *testing.T) {
		converter := serviceClassConverter{}

		labels := gqlschema.Labels{
			"connected-app": "exampleConnectedApp",
			"local":         "true",
			"showcase":      "false",
		}
		maps := map[string]interface{}{
			"displayName":         "exampleDisplayName",
			"providerDisplayName": "exampleProviderName",
			"imageUrl":            "exampleImageURL",
			"documentationUrl":    "exampleDocumentationURL",
			"longDescription":     "exampleLongDescription",
			"supportUrl":          "exampleSupportURL",
			"labels":              labels,
		}

		byteMaps, err := json.Marshal(maps)
		item := v1beta1.ServiceClass{
			Spec: v1beta1.ServiceClassSpec{
				CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
					ExternalMetadata: &runtime.RawExtension{Raw: byteMaps},
					ExternalName:     "ExampleExternalName",
					Tags:             []string{"tag1", "tag2"},
					Description:      "ExampleDescription",
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				Namespace:         "exampleNamespace",
				UID:               types.UID("exampleUid"),
				CreationTimestamp: mockTimeStamp,
				ResourceVersion:   "exampleVersion",
			},
		}

		providerDisplayName := "exampleProviderName"
		imageURL := "exampleImageURL"
		documentationURL := "exampleDocumentationURL"
		supportURL := "exampleSupportURL"
		displayName := "exampleDisplayName"
		longDescription := "exampleLongDescription"
		expected := gqlschema.ServiceClass{
			Name:                "exampleName",
			Namespace:           "exampleNamespace",
			ExternalName:        "ExampleExternalName",
			DisplayName:         &displayName,
			Description:         "ExampleDescription",
			LongDescription:     &longDescription,
			ProviderDisplayName: &providerDisplayName,
			ImageURL:            &imageURL,
			DocumentationURL:    &documentationURL,
			SupportURL:          &supportURL,
			CreationTimestamp:   zeroTimeStamp,
			Tags:                []string{"tag1", "tag2"},
			Labels:              labels,
		}

		result, err := converter.ToGQL(&item)

		assert.Equal(t, err, nil)
		assert.Equal(t, &expected, result)
	})

	t.Run("Invalid externalMetadata (not equals to maps[string]string)", func(t *testing.T) {
		converter := serviceClassConverter{}
		maps := "randomString"
		byteMaps, err := json.Marshal(maps)
		item := v1beta1.ServiceClass{
			Spec: v1beta1.ServiceClassSpec{
				CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
					ExternalMetadata: &runtime.RawExtension{Raw: byteMaps},
					ExternalName:     "ExampleExternalName",
					Tags:             []string{"tag1", "tag2"},
					Description:      "ExampleDescription",
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				UID:               types.UID("exampleUid"),
				CreationTimestamp: mockTimeStamp,
				ResourceVersion:   "exampleVersion",
			},
		}

		_, err = converter.ToGQL(&item)

		assert.Error(t, err)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &serviceClassConverter{}
		_, err := converter.ToGQL(&v1beta1.ServiceClass{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &serviceClassConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestServiceClassConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		classes := []*v1beta1.ServiceClass{
			fixServiceClass(t),
			fixServiceClass(t),
		}

		converter := serviceClassConverter{}
		result, err := converter.ToGQLs(classes)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "exampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var classes []*v1beta1.ServiceClass

		converter := serviceClassConverter{}
		result, err := converter.ToGQLs(classes)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		classes := []*v1beta1.ServiceClass{
			nil,
			fixServiceClass(t),
			nil,
		}

		converter := serviceClassConverter{}
		result, err := converter.ToGQLs(classes)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "exampleName", result[0].Name)
	})
}

func fixServiceClass(t require.TestingT) *v1beta1.ServiceClass {
	var mockTimeStamp metav1.Time

	labels := gqlschema.Labels{
		"connected-app": "exampleConnectedApp",
		"local":         "true",
		"showcase":      "false",
	}
	maps := map[string]interface{}{
		"displayName":         "exampleDisplayName",
		"providerDisplayName": "exampleProviderName",
		"imageUrl":            "exampleImageURL",
		"documentationUrl":    "exampleDocumentationURL",
		"longDescription":     "exampleLongDescription",
		"supportUrl":          "exampleSupportURL",
		"labels":              labels,
	}

	byteMaps, err := json.Marshal(maps)
	require.NoError(t, err)

	return &v1beta1.ServiceClass{
		Spec: v1beta1.ServiceClassSpec{
			CommonServiceClassSpec: v1beta1.CommonServiceClassSpec{
				ExternalMetadata: &runtime.RawExtension{Raw: byteMaps},
				ExternalName:     "ExampleExternalName",
				Tags:             []string{"tag1", "tag2"},
				Description:      "ExampleDescription",
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "exampleName",
			Namespace:         "exampleNamespace",
			UID:               types.UID("exampleUid"),
			CreationTimestamp: mockTimeStamp,
			ResourceVersion:   "exampleVersion",
		},
	}
}
