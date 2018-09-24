package gqlschema

import "time"

type ServiceInstance struct {
	Name              string                      `json:"name"`
	Environment       string                      `json:"environment"`
	ClassReference    *ServiceInstanceResourceRef `json:"classReference"`
	PlanReference     *ServiceInstanceResourceRef `json:"planReference"`
	PlanSpec          *JSON                       `json:"planSpec"`
	CreationTimestamp time.Time                   `json:"creationTimestamp"`
	Labels            []string                    `json:"labels"`
	Status            ServiceInstanceStatus       `json:"status"`
}
