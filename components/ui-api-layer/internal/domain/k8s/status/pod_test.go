package status

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodExtractor_ContainerStatusesToGQLContainerStates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		extractor := PodExtractor{}
		strings := []string{
			"WaitingReason",
			"WaitingMessage",
			"TerminatedReason",
			"TerminatedMessage",
		}
		in := []v1.ContainerStatus{
			fixContainerStatus("Waiting", 0, &v1.ContainerState{
				Waiting: &v1.ContainerStateWaiting{
					Reason:  strings[0],
					Message: strings[1],
				},
			}),
			fixContainerStatus("Running", 0, &v1.ContainerState{
				Running: &v1.ContainerStateRunning{
					StartedAt: metav1.Time{},
				},
			}),
			fixContainerStatus("Terminated", 0, &v1.ContainerState{
				Terminated: &v1.ContainerStateTerminated{
					ExitCode:    111,
					Signal:      222,
					Reason:      strings[2],
					Message:     strings[3],
					StartedAt:   metav1.Time{},
					FinishedAt:  metav1.Time{},
					ContainerID: "",
				},
			}),
			fixContainerStatus("EmptyState", 0, &v1.ContainerState{}),
		}
		expected := []gqlschema.ContainerState{
			{
				State:   gqlschema.ContainerStateTypeWaiting,
				Reason:  strings[0],
				Message: strings[1],
			},
			{
				State:   gqlschema.ContainerStateTypeRunning,
				Reason:  "",
				Message: "",
			},
			{
				State:   gqlschema.ContainerStateTypeTerminated,
				Reason:  strings[2],
				Message: strings[3],
			},
			{
				State:   gqlschema.ContainerStateTypeWaiting,
				Reason:  "",
				Message: "",
			},
		}

		result := extractor.ContainerStatusesToGQLContainerStates(in)

		assert.Equal(t, expected, result)
	})

	t.Run("EmptyPassed", func(t *testing.T) {
		extractor := PodExtractor{}
		in := []v1.ContainerStatus{}
		expected := []gqlschema.ContainerState{}

		result := extractor.ContainerStatusesToGQLContainerStates(in)

		assert.Equal(t, expected, result)
	})

	t.Run("NilPassed", func(t *testing.T) {
		extractor := PodExtractor{}
		var in []v1.ContainerStatus = nil
		expected := []gqlschema.ContainerState{}

		result := extractor.ContainerStatusesToGQLContainerStates(in)

		assert.Equal(t, expected, result)
	})
}

func TestPodConverter_GetTerminatedContainerState(t *testing.T) {
	t.Run("Reason", func(t *testing.T) {
		extractor := PodExtractor{}
		reason := "exampleReason"
		message := "exampleMessage"
		in := v1.ContainerStateTerminated{
			ExitCode: 111,
			Signal:   222,
			Reason:   reason,
			Message:  message,
		}
		expected := fixGQLContainerState(gqlschema.ContainerStateTypeTerminated, reason, message)

		result := extractor.getTerminatedContainerState(&in)

		assert.Equal(t, expected, result)
	})

	t.Run("Signal", func(t *testing.T) {
		extractor := PodExtractor{}
		signal := int32(123)
		message := "exampleMessage"
		in := v1.ContainerStateTerminated{
			ExitCode: 111,
			Signal:   signal,
			Message:  message,
		}
		expectedReason := fmt.Sprintf("Signal: %d", signal)
		expected := fixGQLContainerState(gqlschema.ContainerStateTypeTerminated, expectedReason, message)

		result := extractor.getTerminatedContainerState(&in)

		assert.Equal(t, expected, result)
	})

	t.Run("ExitCode", func(t *testing.T) {
		extractor := PodExtractor{}
		exitCode := int32(123)
		message := "exampleMessage"
		in := v1.ContainerStateTerminated{
			ExitCode: exitCode,
			Message:  message,
		}
		expectedReason := fmt.Sprintf("Exit code: %d", exitCode)
		expected := fixGQLContainerState(gqlschema.ContainerStateTypeTerminated, expectedReason, message)

		result := extractor.getTerminatedContainerState(&in)

		assert.Equal(t, expected, result)
	})

	t.Run("Nil", func(t *testing.T) {
		extractor := PodExtractor{}
		expected := gqlschema.ContainerState{
			State: gqlschema.ContainerStateTypeTerminated,
		}

		result := extractor.getTerminatedContainerState(nil)

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

func fixGQLContainerState(state gqlschema.ContainerStateType, reason string, message string) gqlschema.ContainerState {
	return gqlschema.ContainerState{
		State:   state,
		Reason:  reason,
		Message: message,
	}
}
