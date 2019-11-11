package rafter_test

import (
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestClusterAssetGroupConverter_ToGQL(t *testing.T) {
	converter := rafter.NewClusterAssetGroupConverter()

	t.Run("All properties are given", func(t *testing.T) {
		item := fixClusterAssetGroup()
		expected := gqlschema.ClusterAssetGroup{
			Name:        "ExampleName",
			DisplayName: "DisplayName",
			Description: "Description",
			GroupName:   "exampleGroupName",
			Status: gqlschema.AssetGroupStatus{
				Phase:   gqlschema.AssetGroupPhaseTypeReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		}

		result, err := converter.ToGQL(item)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		_, err := converter.ToGQL(&v1beta1.ClusterAssetGroup{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestClusterDocsTopicConverter_ToGQLs(t *testing.T) {
	converter := rafter.NewClusterAssetGroupConverter()

	t.Run("Success", func(t *testing.T) {
		clusterAssetGroups := []*v1beta1.ClusterAssetGroup{
			fixClusterAssetGroup(),
			fixClusterAssetGroup(),
		}

		result, err := converter.ToGQLs(clusterAssetGroups)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ExampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var clusterAssetGroups []*v1beta1.ClusterAssetGroup

		result, err := converter.ToGQLs(clusterAssetGroups)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		clusterAssetGroups := []*v1beta1.ClusterAssetGroup{
			nil,
			fixClusterAssetGroup(),
			nil,
		}

		result, err := converter.ToGQLs(clusterAssetGroups)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "ExampleName", result[0].Name)
	})
}

func fixClusterAssetGroup() *v1beta1.ClusterAssetGroup {
	return &v1beta1.ClusterAssetGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ExampleName",
			Namespace: "ExampleNamespace",
			Labels: map[string]string{
				rafter.GroupNameLabel: "exampleGroupName",
			},
		},
		Spec: v1beta1.ClusterAssetGroupSpec{
			CommonAssetGroupSpec: v1beta1.CommonAssetGroupSpec{
				DisplayName: "DisplayName",
				Description: "Description",
			},
		},
		Status: v1beta1.ClusterAssetGroupStatus{
			CommonAssetGroupStatus: v1beta1.CommonAssetGroupStatus{
				Phase:   v1beta1.AssetGroupReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		},
	}
}
