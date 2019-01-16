package status

import (
	"fmt"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type PodExtractor struct{}

func (ext *PodExtractor) ContainerStatusesToGQLContainerStates(in []v1.ContainerStatus) []gqlschema.ContainerState {
	containerStates := []gqlschema.ContainerState{}

	for _, containerStatus := range in {
		if containerStatus.State.Waiting != nil {
			containerStates = append(containerStates, ext.getWaitingContainerState(containerStatus.State.Waiting))
		} else if containerStatus.State.Terminated != nil {
			containerStates = append(containerStates, ext.getTerminatedContainerState(containerStatus.State.Terminated))
		} else if containerStatus.State.Running != nil {
			containerStates = append(containerStates, ext.getRunningContainerState())
		} else {
			containerStates = append(containerStates, ext.getDefaultContainerState())
		}
	}

	return containerStates
}

func (ext *PodExtractor) getWaitingContainerState(in *v1.ContainerStateWaiting) gqlschema.ContainerState {
	if in == nil {
		return gqlschema.ContainerState{
			State: gqlschema.ContainerStateTypeWaiting,
		}
	}

	return gqlschema.ContainerState{
		State:   gqlschema.ContainerStateTypeWaiting,
		Reason:  in.Reason,
		Message: in.Message,
	}
}

func (ext *PodExtractor) getRunningContainerState() gqlschema.ContainerState {
	return gqlschema.ContainerState{
		State:   gqlschema.ContainerStateTypeRunning,
		Reason:  "",
		Message: "",
	}
}

func (ext *PodExtractor) getTerminatedContainerState(in *v1.ContainerStateTerminated) gqlschema.ContainerState {
	if in == nil {
		return gqlschema.ContainerState{
			State: gqlschema.ContainerStateTypeTerminated,
		}
	}

	reason := ""
	if in.Reason != "" {
		reason = in.Reason
	} else if in.Signal != 0 {
		reason = fmt.Sprintf("Signal: %d", in.Signal)
	} else {
		reason = fmt.Sprintf("Exit code: %d", in.ExitCode)
	}

	return gqlschema.ContainerState{
		State:   gqlschema.ContainerStateTypeTerminated,
		Reason:  reason,
		Message: in.Message,
	}
}

func (ext *PodExtractor) getDefaultContainerState() gqlschema.ContainerState {
	return ext.getWaitingContainerState(nil)
}
