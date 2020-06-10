package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReplicaSetConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := &replicaSetConverter{}
		in := fixExampleReplicaSet("exampleKind", "exampleName", "exampleNamespace", apps.ReplicaSetStatus{
			Replicas:      2,
			ReadyReplicas: 1,
		})
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
			Images: []string{"exampleImage_1", "exampleImage_2"},
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
			Pods:   "0/0",
			JSON:   emptyReplicaSetJSON,
			Labels: gqlschema.Labels{},
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
		expectedMap := fixExampleReplicaSetMap("exampleKind", "exampleName", "exampleNamespace",
			map[string]interface{}{
				"replicas":      float64(2),
				"readyReplicas": float64(1),
			})
		in := fixExampleReplicaSet("exampleKind", "exampleName", "exampleNamespace", apps.ReplicaSetStatus{
			Replicas:      2,
			ReadyReplicas: 1,
		})
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
		inMap := fixExampleReplicaSetMap("exampleKind", "exampleName", "exampleNamespace",
			map[string]interface{}{
				"replicas":      2,
				"readyReplicas": 1,
			})
		expected := fixExampleReplicaSet("exampleKind", "exampleName", "exampleNamespace", apps.ReplicaSetStatus{
			Replicas:      2,
			ReadyReplicas: 1,
		})
		inJSON := new(gqlschema.JSON)
		err := inJSON.UnmarshalGQL(inMap)
		require.NoError(t, err)

		result, err := converter.GQLJSONToReplicaSet(*inJSON)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func fixExampleReplicaSet(kind string, name string, namespcace string, status apps.ReplicaSetStatus) apps.ReplicaSet {
	return apps.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind: kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespcace,
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
		Status: status,
	}
}

func fixExampleReplicaSetMap(kind string, name string, namespace string, status map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"kind": kind,
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
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
		"status": status,
	}
}
