package cms

import (
	"testing"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterDocsTopicConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := clusterDocsTopicConverter{}

		item := fixClusterDocsTopic()
		expected := gqlschema.ClusterDocsTopic{
			Name:        "ExampleName",
			DisplayName: "DisplayName",
			Description: "Description",
			GroupName:   "exampleGroupName",
			Status: gqlschema.DocsTopicStatus{
				Phase:   gqlschema.DocsTopicPhaseTypeReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		}

		result, err := converter.ToGQL(item)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &clusterDocsTopicConverter{}
		_, err := converter.ToGQL(&v1alpha1.ClusterDocsTopic{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &clusterDocsTopicConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestClusterDocsTopicConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		clusterDocsTopics := []*v1alpha1.ClusterDocsTopic{
			fixClusterDocsTopic(),
			fixClusterDocsTopic(),
		}

		converter := clusterDocsTopicConverter{}
		result, err := converter.ToGQLs(clusterDocsTopics)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ExampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var clusterDocsTopics []*v1alpha1.ClusterDocsTopic

		converter := clusterDocsTopicConverter{}
		result, err := converter.ToGQLs(clusterDocsTopics)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		clusterDocsTopics := []*v1alpha1.ClusterDocsTopic{
			nil,
			fixClusterDocsTopic(),
			nil,
		}

		converter := clusterDocsTopicConverter{}
		result, err := converter.ToGQLs(clusterDocsTopics)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "ExampleName", result[0].Name)
	})
}

func fixClusterDocsTopic() *v1alpha1.ClusterDocsTopic {
	return &v1alpha1.ClusterDocsTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ExampleName",
			Labels: map[string]string{
				GroupNameLabel: "exampleGroupName",
			},
		},
		Spec: v1alpha1.ClusterDocsTopicSpec{
			CommonDocsTopicSpec: v1alpha1.CommonDocsTopicSpec{
				DisplayName: "DisplayName",
				Description: "Description",
			},
		},
		Status: v1alpha1.ClusterDocsTopicStatus{
			CommonDocsTopicStatus: v1alpha1.CommonDocsTopicStatus{
				Phase:   v1alpha1.DocsTopicReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		},
	}
}
