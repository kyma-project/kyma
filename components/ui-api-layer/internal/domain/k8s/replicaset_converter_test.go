package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReplicaSetConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := &replicaSetConverter{}
		in := apps.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				Namespace:         "exampleNamespace",
				CreationTimestamp: metav1.Time{},
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
			},
			Spec: apps.ReplicaSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "image_1",
							},
							{
								Image: "image_2",
							},
						},
					},
				},
			},
			Status: apps.ReplicaSetStatus{
				Replicas:      2,
				ReadyReplicas: 1,
			},
		}
		expectedJSON, err := converter.replicaSetToGQLJSON(&in)
		require.NoError(t, err)
		expected := gqlschema.ReplicaSet{
			Name:              "exampleName",
			Namespace:         "exampleNamespace",
			Pods:              "1/2",
			CreationTimestamp: time.Time{},
			Labels: map[string]string{
				"exampleKey":  "exampleValue",
				"exampleKey2": "exampleValue2",
			},
			Images: []string{"image_1", "image_2"},
			JSON:   expectedJSON,
		}

		result, err := converter.ToGQL(&in)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)

	})

	t.Run("Empty", func(t *testing.T) {
		converter := &replicaSetConverter{}
		emptyReplicaSetJSON, err := converter.replicaSetToGQLJSON(&apps.ReplicaSet{})
		require.NoError(t, err)
		expected := &gqlschema.ReplicaSet{
			Pods: "0/0",
			JSON: emptyReplicaSetJSON,
		}

		result, err := converter.ToGQL(&apps.ReplicaSet{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &replicaSetConverter{}

		result, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestReplicaSetConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := replicaSetConverter{}
		expectedName := "exampleName"
		in := []*apps.ReplicaSet{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: expectedName,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "exampleName2",
				},
			},
		}

		result, err := converter.ToGQLs(in)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedName, result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := replicaSetConverter{}
		var in []*apps.ReplicaSet

		result, err := converter.ToGQLs(in)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		converter := replicaSetConverter{}
		expectedName := "exampleName"
		in := []*apps.ReplicaSet{
			nil,
			&apps.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name: expectedName,
				},
			},
			nil,
		}

		result, err := converter.ToGQLs(in)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, expectedName, result[0].Name)
	})
}

func TestReplicaSetConverter_ReplicaSetToGQLJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := replicaSetConverter{}
		expectedMap := map[string]interface{}{
			"kind": "exampleKind",
			"metadata": map[string]interface{}{
				"name":      "exampleName",
				"namespace": "exampleNamespace",
				"labels": map[string]interface{}{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				"creationTimestamp": nil,
			},
			"spec": map[string]interface{}{
				"selector": interface{}(nil),
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"creationTimestamp": interface{}(nil),
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"resources": map[string]interface{}{},
								"name":      "",
								"image":     "exampleImage_1",
							},
							map[string]interface{}{
								"resources": map[string]interface{}{},
								"name":      "",
								"image":     "exampleImage_2",
							},
						},
					},
				},
			},
			"status": map[string]interface{}{
				"replicas":      float64(2),
				"readyReplicas": float64(1),
			},
		}
		in := apps.ReplicaSet{
			TypeMeta: metav1.TypeMeta{
				Kind: "exampleKind",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "exampleName",
				Namespace: "exampleNamespace",
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				CreationTimestamp: metav1.Time{},
			},
			Spec: apps.ReplicaSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "exampleImage_1",
							},
							{
								Image: "exampleImage_2",
							},
						},
					},
				},
			},
			Status: apps.ReplicaSetStatus{
				Replicas:      2,
				ReadyReplicas: 1,
			},
		}

		expectedJSON := new(gqlschema.JSON)
		err := expectedJSON.UnmarshalGQL(expectedMap)
		require.NoError(t, err)

		result, err := converter.replicaSetToGQLJSON(&in)

		require.NoError(t, err)
		assert.Equal(t, *expectedJSON, result)
	})

	t.Run("NilPassed", func(t *testing.T) {
		converter := replicaSetConverter{}

		result, err := converter.replicaSetToGQLJSON(nil)

		require.Nil(t, result)
		require.NoError(t, err)
	})
}

func TestReplicaSetConverter_GQLJSONToReplicaSet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := replicaSetConverter{}
		inMap := map[string]interface{}{
			"kind": "exampleKind",
			"metadata": map[string]interface{}{
				"name":      "exampleName",
				"namespace": "exampleNamespace",
				"labels": map[string]interface{}{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				"creationTimestamp": nil,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"image": "exampleImage_1",
							},
							map[string]interface{}{
								"image": "exampleImage_2",
							},
						},
					},
				},
			},
			"status": map[string]interface{}{
				"replicas":      2,
				"readyReplicas": 1,
			},
		}
		expected := apps.ReplicaSet{
			TypeMeta: metav1.TypeMeta{
				Kind: "exampleKind",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "exampleName",
				Namespace: "exampleNamespace",
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				CreationTimestamp: metav1.Time{},
			},
			Spec: apps.ReplicaSetSpec{
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Image: "exampleImage_1",
							},
							{
								Image: "exampleImage_2",
							},
						},
					},
				},
			},
			Status: apps.ReplicaSetStatus{
				Replicas:      2,
				ReadyReplicas: 1,
			},
		}

		inJSON := new(gqlschema.JSON)
		err := inJSON.UnmarshalGQL(inMap)
		require.NoError(t, err)

		result, err := converter.GQLJSONToReplicaSet(*inJSON)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}
