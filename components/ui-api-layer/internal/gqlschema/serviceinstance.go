package gqlschema

import "time"

type ServiceInstance struct {
	Name              string                      `json:"name"`
	Namespace         string                      `json:"namespace"`
	ClassReference    *ServiceInstanceResourceRef `json:"classReference"`
	PlanReference     *ServiceInstanceResourceRef `json:"planReference"`
	PlanSpec          *JSON                       `json:"planSpec"`
	CreationTimestamp time.Time                   `json:"creationTimestamp"`
	Labels            []string                    `json:"labels"`
	Status            ServiceInstanceStatus       `json:"status"`
}
