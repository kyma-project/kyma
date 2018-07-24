package gqlschema

type DeploymentStatus struct {
	Replicas          int
	UpdatedReplicas   int
	ReadyReplicas     int
	AvailableReplicas int
	Conditions        []DeploymentCondition
}
