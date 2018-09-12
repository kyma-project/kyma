package servicecatalog

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestClassConverter_ToGQL(t *testing.T) {
	var mockTimeStamp metav1.Time
	var zeroTimeStamp time.Time
	t.Run("All properties are given", func(t *testing.T) {
		converter := classConverter{}
		maps := map[string]string{
			"displayName":         "exampleDisplayName",
			"providerDisplayName": "exampleProviderName",
			"imageUrl":            "exampleImageUrl",
			"documentationUrl":    "exampleDocumentationUrl",
			"longDescription":     "exampleLongDescription",
			"supportUrl":          "exampleSupportUrl",
		}

		byteMaps, err := json.Marshal(maps)
		item := v1beta1.ClusterServiceClass{
			Spec: v1beta1.ClusterServiceClassSpec{
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

		providerDisplayName := "exampleProviderName"
		imageUrl := "exampleImageUrl"
		documentationUrl := "exampleDocumentationUrl"
		supportUrl := "exampleSupportUrl"
		displayName := "exampleDisplayName"
		longDescription := "exampleLongDescription"
		expected := gqlschema.ServiceClass{
			Name:                "exampleName",
			ExternalName:        "ExampleExternalName",
			DisplayName:         &displayName,
			Description:         "ExampleDescription",
			LongDescription:     &longDescription,
			ProviderDisplayName: &providerDisplayName,
			ImageUrl:            &imageUrl,
			DocumentationUrl:    &documentationUrl,
			SupportUrl:          &supportUrl,
			CreationTimestamp:   zeroTimeStamp,
			Tags:                []string{"tag1", "tag2"},
		}

		result, err := converter.ToGQL(&item)

		assert.Equal(t, err, nil)
		assert.Equal(t, &expected, result)
	})

	t.Run("Invalid externalMetadata (not equals to maps[string]string)", func(t *testing.T) {
		converter := classConverter{}
		maps := "randomString"
		byteMaps, err := json.Marshal(maps)
		item := v1beta1.ClusterServiceClass{
			Spec: v1beta1.ClusterServiceClassSpec{
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
		converter := &classConverter{}
		_, err := converter.ToGQL(&v1beta1.ClusterServiceClass{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &classConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestClassConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		classes := []*v1beta1.ClusterServiceClass{
			fixServiceClass(t),
			fixServiceClass(t),
		}

		converter := classConverter{}
		result, err := converter.ToGQLs(classes)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "exampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var classes []*v1beta1.ClusterServiceClass

		converter := classConverter{}
		result, err := converter.ToGQLs(classes)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		classes := []*v1beta1.ClusterServiceClass{
			nil,
			fixServiceClass(t),
			nil,
		}

		converter := classConverter{}
		result, err := converter.ToGQLs(classes)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "exampleName", result[0].Name)
	})
}

func fixServiceClass(t require.TestingT) *v1beta1.ClusterServiceClass {
	var mockTimeStamp metav1.Time
	maps := map[string]string{
		"displayName":         "exampleDisplayName",
		"providerDisplayName": "exampleProviderName",
		"imageUrl":            "exampleImageUrl",
		"documentationUrl":    "exampleDocumentationUrl",
		"longDescription":     "exampleLongDescription",
		"supportUrl":          "exampleSupportUrl",
	}

	byteMaps, err := json.Marshal(maps)
	require.NoError(t, err)

	return &v1beta1.ClusterServiceClass{
		Spec: v1beta1.ClusterServiceClassSpec{
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
}
