package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := &podConverter{}
		in := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				Namespace:         "exampleNamespace",
				CreationTimestamp: metav1.Time{},
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
			},
			Spec: v1.PodSpec{
				NodeName: "exampleNodeName",
			},
			Status: v1.PodStatus{
				Phase: v1.PodPending,
				ContainerStatuses: []v1.ContainerStatus{
					fixContainerStatus("exampleStatus", 123, &v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{
							Reason:  "exampleReason",
							Message: "exampleMessage",
						},
					}),
				},
			},
		}
		expectedJSON, err := converter.podToGQLJSON(&in)
		require.NoError(t, err)
		expected := gqlschema.Pod{
			Name:              "exampleName",
			NodeName:          "exampleNodeName",
			Namespace:         "exampleNamespace",
			RestartCount:      123,
			CreationTimestamp: time.Time{},
			Labels: map[string]string{
				"exampleKey":  "exampleValue",
				"exampleKey2": "exampleValue2",
			},
			Status: gqlschema.PodStatusTypePending,
			ContainerStates: []*gqlschema.ContainerState{
				fixGQLContainerState(gqlschema.ContainerStateTypeWaiting, "exampleReason", "exampleMessage"),
			},
			JSON: expectedJSON,
		}

		result, err := converter.ToGQL(&in)

		require.NoError(t, err)
		assert.Equal(t, &expected, result)

	})

	t.Run("Empty", func(t *testing.T) {
		converter := &podConverter{}
		emptyPodJSON, err := converter.podToGQLJSON(&v1.Pod{})
		require.NoError(t, err)
		expected := &gqlschema.Pod{
			Status:          gqlschema.PodStatusTypeUnknown,
			ContainerStates: []*gqlschema.ContainerState{},
			JSON:            emptyPodJSON,
			Labels:          gqlschema.Labels{},
		}

		result, err := converter.ToGQL(&v1.Pod{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &podConverter{}

		result, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestPodConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := podConverter{}
		expectedName := "exampleName"
		in := []*v1.Pod{
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
		converter := podConverter{}
		var in []*v1.Pod

		result, err := converter.ToGQLs(in)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		converter := podConverter{}
		expectedName := "exampleName"
		in := []*v1.Pod{
			nil,
			&v1.Pod{
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

func TestPodConverter_PodToGQLJSON(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := podConverter{}
		exampleContainerStateWaiting := v1.ContainerStateWaiting{
			Reason:  "exampleReason",
			Message: "exampleMessage",
		}
		expectedMap := map[string]interface{}{
			"kind": "exampleKind",
			"metadata": map[string]interface{}{
				"name": "exampleName",
				"labels": map[string]interface{}{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				"creationTimestamp": nil,
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": "exampleApiVersion",
						"kind":       "exampleKind",
						"name":       "exampleName",
						"uid":        "exampleUID",
					},
				},
			},
			"spec": map[string]interface{}{
				"nodeName":   "exampleNodeName",
				"containers": nil,
			},
			"status": map[string]interface{}{
				"phase": "examplePhase",
				"conditions": []interface{}{
					map[string]interface{}{
						"reason":             "exampleReason",
						"type":               "exampleType",
						"status":             "exampleStatus",
						"message":            "exampleMessage",
						"lastProbeTime":      nil,
						"lastTransitionTime": nil,
					},
					map[string]interface{}{
						"reason":             "exampleReason",
						"type":               "exampleType",
						"status":             "exampleStatus",
						"message":            "exampleMessage",
						"lastProbeTime":      nil,
						"lastTransitionTime": nil,
					},
				},
				"containerStatuses": []interface{}{
					map[string]interface{}{
						"name":         "exampleName",
						"restartCount": float64(5),
						"ready":        true,
						"state": map[string]interface{}{
							"waiting": map[string]interface{}{
								"reason":  "exampleReason",
								"message": "exampleMessage",
							},
						},
						"image":     "",
						"imageID":   "",
						"lastState": map[string]interface{}{},
					},
				},
			},
		}
		in := v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind: "exampleKind",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "exampleName",
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				CreationTimestamp: metav1.Time{},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "exampleApiVersion",
						Name:       "exampleName",
						UID:        "exampleUID",
						Kind:       "exampleKind",
					},
				},
			},
			Spec: v1.PodSpec{
				NodeName: "exampleNodeName",
			},
			Status: v1.PodStatus{
				Phase: "examplePhase",
				Conditions: []v1.PodCondition{
					{
						Type:               "exampleType",
						Status:             "exampleStatus",
						LastProbeTime:      metav1.Time{},
						LastTransitionTime: metav1.Time{},
						Reason:             "exampleReason",
						Message:            "exampleMessage",
					},
					{
						Type:               "exampleType",
						Status:             "exampleStatus",
						LastProbeTime:      metav1.Time{},
						LastTransitionTime: metav1.Time{},
						Reason:             "exampleReason",
						Message:            "exampleMessage",
					},
				},
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "exampleName",
						State: v1.ContainerState{
							Waiting: &exampleContainerStateWaiting,
						},
						LastTerminationState: v1.ContainerState{},
						Ready:                true,
						RestartCount:         5,
						Image:                "",
						ImageID:              "",
						ContainerID:          "",
					},
				},
			},
		}

		expectedJSON := new(gqlschema.JSON)
		err := expectedJSON.UnmarshalGQL(expectedMap)
		require.NoError(t, err)

		result, err := converter.podToGQLJSON(&in)

		require.NoError(t, err)
		assert.Equal(t, *expectedJSON, result)
	})

	t.Run("NilPassed", func(t *testing.T) {
		converter := podConverter{}

		result, err := converter.podToGQLJSON(nil)

		require.Nil(t, result)
		require.NoError(t, err)
	})
}

func TestPodConverter_GQLJSONToPod(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := podConverter{}
		exampleContainerStateWaiting := v1.ContainerStateWaiting{
			Reason:  "exampleReason",
			Message: "exampleMessage",
		}
		inMap := map[string]interface{}{
			"kind": "exampleKind",
			"metadata": map[string]interface{}{
				"name": "exampleName",
				"labels": map[string]interface{}{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				"creationTimestamp": nil,
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": "exampleApiVersion",
						"kind":       "exampleKind",
						"name":       "exampleName",
						"uid":        "exampleUID",
					},
				},
			},
			"spec": map[string]interface{}{
				"nodeName":   "exampleNodeName",
				"containers": nil,
			},
			"status": map[string]interface{}{
				"phase": "examplePhase",
				"conditions": []interface{}{
					map[string]interface{}{
						"reason":             "exampleReason",
						"type":               "exampleType",
						"status":             "exampleStatus",
						"message":            "exampleMessage",
						"lastProbeTime":      nil,
						"lastTransitionTime": nil,
					},
					map[string]interface{}{
						"reason":             "exampleReason",
						"type":               "exampleType",
						"status":             "exampleStatus",
						"message":            "exampleMessage",
						"lastProbeTime":      nil,
						"lastTransitionTime": nil,
					},
				},
				"containerStatuses": []interface{}{
					map[string]interface{}{
						"name":         "exampleName",
						"restartCount": float64(5),
						"ready":        true,
						"state": map[string]interface{}{
							"waiting": map[string]interface{}{
								"reason":  "exampleReason",
								"message": "exampleMessage",
							},
						},
						"image":     "",
						"imageID":   "",
						"lastState": map[string]interface{}{},
					},
				},
			},
		}
		expected := v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind: "exampleKind",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "exampleName",
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
				CreationTimestamp: metav1.Time{},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "exampleApiVersion",
						Name:       "exampleName",
						UID:        "exampleUID",
						Kind:       "exampleKind",
					},
				},
			},
			Spec: v1.PodSpec{
				NodeName: "exampleNodeName",
			},
			Status: v1.PodStatus{
				Phase: "examplePhase",
				Conditions: []v1.PodCondition{
					{
						Type:               "exampleType",
						Status:             "exampleStatus",
						LastProbeTime:      metav1.Time{},
						LastTransitionTime: metav1.Time{},
						Reason:             "exampleReason",
						Message:            "exampleMessage",
					},
					{
						Type:               "exampleType",
						Status:             "exampleStatus",
						LastProbeTime:      metav1.Time{},
						LastTransitionTime: metav1.Time{},
						Reason:             "exampleReason",
						Message:            "exampleMessage",
					},
				},
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "exampleName",
						State: v1.ContainerState{
							Waiting: &exampleContainerStateWaiting,
						},
						LastTerminationState: v1.ContainerState{},
						Ready:                true,
						RestartCount:         5,
						Image:                "",
						ImageID:              "",
						ContainerID:          "",
					},
				},
			},
		}

		inJSON := new(gqlschema.JSON)
		err := inJSON.UnmarshalGQL(inMap)
		require.NoError(t, err)

		result, err := converter.GQLJSONToPod(*inJSON)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func TestPodConverter_PodStatusPhaseToGQLStatusType(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		converter := podConverter{}

		types := map[v1.PodPhase]gqlschema.PodStatusType{
			v1.PodPending:   gqlschema.PodStatusTypePending,
			v1.PodRunning:   gqlschema.PodStatusTypeRunning,
			v1.PodSucceeded: gqlschema.PodStatusTypeSucceeded,
			v1.PodFailed:    gqlschema.PodStatusTypeFailed,
			v1.PodUnknown:   gqlschema.PodStatusTypeUnknown,
		}

		for k, v := range types {
			result := converter.podStatusPhaseToGQLStatusType(k)
			assert.Equal(t, v, result)
		}
	})
}

func TestPodConverter_GetRestartCount(t *testing.T) {
	restarts := []int{22, 0, 15}

	t.Run("Empty", func(t *testing.T) {
		converter := podConverter{}
		in := []v1.ContainerStatus{}
		expected := 0

		result := converter.getRestartCount(in)

		assert.Equal(t, expected, result)
	})

	t.Run("Single", func(t *testing.T) {
		converter := podConverter{}
		in := []v1.ContainerStatus{
			fixContainerStatus("test1", restarts[0], nil),
		}
		expected := restarts[0]

		result := converter.getRestartCount(in)

		assert.Equal(t, expected, result)
	})

	t.Run("Multiple", func(t *testing.T) {
		converter := podConverter{}
		in := []v1.ContainerStatus{
			fixContainerStatus("test1", restarts[0], nil),
			fixContainerStatus("test2", restarts[1], nil),
			fixContainerStatus("test3", restarts[2], nil),
		}
		expected := restarts[0] + restarts[1] + restarts[2]

		result := converter.getRestartCount(in)

		assert.Equal(t, expected, result)
	})
}

func fixContainerStatus(name string, restartCount int, state *v1.ContainerState) v1.ContainerStatus {
	if state == nil {
		state = &v1.ContainerState{}
	}

	return v1.ContainerStatus{
		Name:         name,
		State:        *state,
		RestartCount: int32(restartCount),
	}
}

func fixGQLContainerState(state gqlschema.ContainerStateType, reason string, message string) *gqlschema.ContainerState {
	return &gqlschema.ContainerState{
		State:   state,
		Reason:  reason,
		Message: message,
	}
}
