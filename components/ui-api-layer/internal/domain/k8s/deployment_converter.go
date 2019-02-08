package k8s

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	api "k8s.io/api/apps/v1beta2"
)

type deploymentConverter struct{}

func (c *deploymentConverter) ToGQL(in *api.Deployment) *gqlschema.Deployment {
	if in == nil {
		return nil
	}

	return &gqlschema.Deployment{
		Name:              in.Name,
		Namespace:         in.Namespace,
		CreationTimestamp: in.CreationTimestamp.Time,
		Labels:            in.Labels,
		Status:            c.toGQLStatus(*in),
		Containers:        c.toGQLContainers(*in),
	}
}

func (c *deploymentConverter) ToGQLs(in []*api.Deployment) []gqlschema.Deployment {
	var result []gqlschema.Deployment
	for _, item := range in {
		converted := c.ToGQL(item)

		if converted != nil {
			result = append(result, *converted)
		}
	}

	return result
}

func (c *deploymentConverter) toGQLStatus(in api.Deployment) gqlschema.DeploymentStatus {
	var conditions []gqlschema.DeploymentCondition
	for _, condition := range in.Status.Conditions {
		conditions = append(conditions, c.toGQLCondition(condition))
	}

	return gqlschema.DeploymentStatus{
		AvailableReplicas: int(in.Status.AvailableReplicas),
		ReadyReplicas:     int(in.Status.ReadyReplicas),
		Replicas:          int(in.Status.Replicas),
		UpdatedReplicas:   int(in.Status.UpdatedReplicas),
		Conditions:        conditions,
	}
}

func (c *deploymentConverter) toGQLCondition(in api.DeploymentCondition) gqlschema.DeploymentCondition {
	return gqlschema.DeploymentCondition{
		Reason:                  in.Reason,
		Message:                 in.Message,
		LastUpdateTimestamp:     in.LastUpdateTime.Time,
		LastTransitionTimestamp: in.LastTransitionTime.Time,
		Type:                    string(in.Type),
		Status:                  string(in.Status),
	}
}

func (c *deploymentConverter) toGQLContainers(in api.Deployment) []gqlschema.Container {
	var containers []gqlschema.Container
	for _, container := range in.Spec.Template.Spec.Containers {
		gqlContainer := gqlschema.Container{
			Name:  container.Name,
			Image: container.Image,
		}

		containers = append(containers, gqlContainer)
	}

	return containers
}
