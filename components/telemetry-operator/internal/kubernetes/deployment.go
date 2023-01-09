package kubernetes

import (
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type DeploymentProber struct {
	client.Client
}

func (dp *DeploymentProber) IsReady(ctx context.Context, name types.NamespacedName) (bool, error) {
	log := logf.FromContext(ctx)

	var d v1.Deployment
	if err := dp.Get(ctx, name, &d); err != nil {
		return false, fmt.Errorf("failed to get %s/%s Deployment: %v", name.Namespace, name.Name, err)
	}

	generation := d.Generation
	observedGeneration := d.Status.ObservedGeneration
	updated := d.Status.UpdatedReplicas
	desired := d.Status.Replicas
	ready := d.Status.ReadyReplicas

	log.V(1).Info(fmt.Sprintf("Checking Deployment: updated: %d, desired: %d, ready: %d, generation: %d, observed generation: %d",
		updated, desired, ready, generation, observedGeneration), "name", name.Name)

	return observedGeneration == generation && updated == desired && ready >= desired, nil
}
