package gqlschema

import "time"

type ServiceInstance struct {
	Name                    string                `json:"name"`
	Environment             string                `json:"environment"`
	ServiceClassName        *string               `json:"serviceClassName"`
	ServiceClassDisplayName string                `json:"ServiceClassDisplayName"`
	ServicePlanName         *string               `json:"servicePlanName"`
	ServicePlanDisplayName  string                `json:"servicePlanDisplayName"`
	ServicePlanSpec         *JSON                 `json:"servicePlanSpec"`
	CreationTimestamp       time.Time             `json:"creationTimestamp"`
	Labels                  []string              `json:"labels"`
	Status                  ServiceInstanceStatus `json:"status"`
}
