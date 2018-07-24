package gqlschema

import "time"

type Deployment struct {
	Name              string
	Environment       string
	CreationTimestamp time.Time
	Status            DeploymentStatus
	Labels            JSON
	Containers        []Container
}
