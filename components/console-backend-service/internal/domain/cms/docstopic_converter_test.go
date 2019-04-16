package cms

import (
	"testing"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDocsTopicConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := docsTopicConverter{}

		item := fixDocsTopic()
		expected := gqlschema.DocsTopic{
			Name:        "ExampleName",
			Namespace:   "ExampleNamespace",
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
		converter := &docsTopicConverter{}
		_, err := converter.ToGQL(&v1alpha1.DocsTopic{})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &docsTopicConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestDocsTopicConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		docsTopics := []*v1alpha1.DocsTopic{
			fixDocsTopic(),
			fixDocsTopic(),
		}

		converter := docsTopicConverter{}
		result, err := converter.ToGQLs(docsTopics)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ExampleName", result[0].Name)
		assert.Equal(t, "ExampleNamespace", result[0].Namespace)
	})

	t.Run("Empty", func(t *testing.T) {
		var docsTopics []*v1alpha1.DocsTopic

		converter := docsTopicConverter{}
		result, err := converter.ToGQLs(docsTopics)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		docsTopics := []*v1alpha1.DocsTopic{
			nil,
			fixDocsTopic(),
			nil,
		}

		converter := docsTopicConverter{}
		result, err := converter.ToGQLs(docsTopics)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "ExampleName", result[0].Name)
		assert.Equal(t, "ExampleNamespace", result[0].Namespace)
	})
}

func fixDocsTopic() *v1alpha1.DocsTopic {
	return &v1alpha1.DocsTopic{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ExampleName",
			Namespace: "ExampleNamespace",
			Labels: map[string]string{
				GroupNameLabel: "exampleGroupName",
			},
		},
		Spec: v1alpha1.DocsTopicSpec{
			CommonDocsTopicSpec: v1alpha1.CommonDocsTopicSpec{
				DisplayName: "DisplayName",
				Description: "Description",
			},
		},
		Status: v1alpha1.DocsTopicStatus{
			CommonDocsTopicStatus: v1alpha1.CommonDocsTopicStatus{
				Phase:   v1alpha1.DocsTopicReady,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		},
	}
}
