package k8s

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type podConverter struct{}

func (c *podConverter) ToGQL(in *v1.Pod) (*gqlschema.Pod, error) {
	if in == nil {
		return nil, nil
	}

	containerStates := c.containerStatusesToGQLContainerStates(in.Status.ContainerStatuses)

	gqlJSON, err := c.podToGQLJSON(in)
	if err != nil {
		return nil, err
	}

	return &gqlschema.Pod{
		Name:              in.Name,
		NodeName:          in.Spec.NodeName,
		Namespace:         in.Namespace,
		RestartCount:      c.getRestartCount(in.Status.ContainerStatuses),
		CreationTimestamp: in.CreationTimestamp.Time,
		Labels:            in.Labels,
		Status:            c.podStatusPhaseToGQLStatusType(in.Status.Phase),
		ContainerStates:   containerStates,
		JSON:              gqlJSON,
	}, nil
}

func (c *podConverter) ToGQLs(in []*v1.Pod) ([]gqlschema.Pod, error) {
	var result []gqlschema.Pod
	for _, u := range in {
		converted, err := c.ToGQL(u)
		if err != nil {
			return nil, err
		}

		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result, nil
}

func (c *podConverter) containerStatusesToGQLContainerStates(in []v1.ContainerStatus) []gqlschema.ContainerState {
	containerStates := []gqlschema.ContainerState{}

	for _, containerStatus := range in {
		if containerStatus.State.Waiting != nil {
			containerStates = append(containerStates, c.getWaitingContainerState(containerStatus.State.Waiting))
		} else if containerStatus.State.Terminated != nil {
			containerStates = append(containerStates, c.getTerminatedContainerState(containerStatus.State.Terminated))
		} else if containerStatus.State.Running != nil {
			containerStates = append(containerStates, c.getRunningContainerState())
		} else {
			containerStates = append(containerStates, c.getDefaultContainerState())
		}
	}

	return containerStates
}

func (c *podConverter) getWaitingContainerState(in *v1.ContainerStateWaiting) gqlschema.ContainerState {
	if in == nil {
		return gqlschema.ContainerState{
			State: gqlschema.ContainerStateTypeWaiting,
		}
	}

	var reason, message *string
	if in.Reason != "" {
		tmp := in.Reason
		reason = &tmp
	}
	if in.Message != "" {
		tmp := in.Message
		message = &tmp
	}

	return gqlschema.ContainerState{
		State:   gqlschema.ContainerStateTypeWaiting,
		Reason:  reason,
		Message: message,
	}
}

func (c *podConverter) getRunningContainerState() gqlschema.ContainerState {
	return gqlschema.ContainerState{
		State:   gqlschema.ContainerStateTypeRunning,
		Reason:  nil,
		Message: nil,
	}
}

func (c *podConverter) getTerminatedContainerState(in *v1.ContainerStateTerminated) gqlschema.ContainerState {
	if in == nil {
		return gqlschema.ContainerState{
			State: gqlschema.ContainerStateTypeTerminated,
		}
	}

	var reason, message *string
	if in.Reason != "" {
		tmp := in.Reason
		reason = &tmp
	} else if in.Signal != 0 {
		tmp := fmt.Sprintf("Signal: %d", in.Signal)
		reason = &tmp
	} else {
		tmp := fmt.Sprintf("Exit code: %d", in.ExitCode)
		reason = &tmp
	}
	if in.Message != "" {
		tmp := in.Message
		message = &tmp
	}

	return gqlschema.ContainerState{
		State:   gqlschema.ContainerStateTypeTerminated,
		Reason:  reason,
		Message: message,
	}
}

func (c *podConverter) getDefaultContainerState() gqlschema.ContainerState {
	return c.getWaitingContainerState(nil)
}

func (c *podConverter) podToGQLJSON(in *v1.Pod) (gqlschema.JSON, error) {
	if in == nil {
		return nil, nil
	}

	jsonByte, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonByte, &jsonMap)
	if err != nil {
		return nil, err
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(jsonMap)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *podConverter) podStatusPhaseToGQLStatusType(in v1.PodPhase) gqlschema.PodStatusType {
	switch in {
	case v1.PodPending:
		return gqlschema.PodStatusTypePending
	case v1.PodRunning:
		return gqlschema.PodStatusTypeRunning
	case v1.PodSucceeded:
		return gqlschema.PodStatusTypeSucceeded
	case v1.PodFailed:
		return gqlschema.PodStatusTypeFailed
	default:
		return gqlschema.PodStatusTypeUnknown
	}
}

func (c *podConverter) getRestartCount(in []v1.ContainerStatus) int {
	restartCount := 0
	for _, containerStatus := range in {
		restartCount += int(containerStatus.RestartCount)
	}
	return restartCount
}
