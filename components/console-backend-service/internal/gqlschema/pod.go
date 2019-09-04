package gqlschema

import "time"

type Pod struct {
	Name              string           `json:"name"`
	NodeName          string           `json:"nodeName"`
	Namespace         string           `json:"namespace"`
	RestartCount      int              `json:"restartCount"`
	CreationTimestamp time.Time        `json:"creationTimestamp"`
	Labels            Labels           `json:"labels"`
	Status            PodStatusType    `json:"status"`
	ContainerStates   []ContainerState `json:"containerStates"`
	JSON              JSON             `json:"json"`
}