package gqlschema

import "time"

type ServiceInstance struct {
	Name                    string
	Environment             string
	ServiceClassName        *string
	ServiceClassDisplayName string
	ServicePlanName         *string
	ServicePlanDisplayName  string
	ServicePlanSpec         *JSON
	CreationTimestamp       time.Time
	Labels                  []string
	Status                  ServiceInstanceStatus
}
