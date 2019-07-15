package testkit

import (
	coreApi "k8s.io/api/core/v1"
)

type PodHelper struct {
}

func (PodHelper) IsPodReady(pod coreApi.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == coreApi.PodReady {
			return condition.Status == coreApi.ConditionTrue
		}
	}
	return false
}
