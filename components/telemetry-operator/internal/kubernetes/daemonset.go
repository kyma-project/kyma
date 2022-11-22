package kubernetes

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type DaemonSetProber struct {
	client.Client
}

func (dsp *DaemonSetProber) IsReady(ctx context.Context, name types.NamespacedName) (bool, error) {
	log := logf.FromContext(ctx)

	var ds appsv1.DaemonSet
	if err := dsp.Get(ctx, name, &ds); err != nil {
		return false, fmt.Errorf("failed to get %s/%s DaemonSet: %v", name.Namespace, name.Name, err)
	}

	generation := ds.Generation
	observedGeneration := ds.Status.ObservedGeneration
	updated := ds.Status.UpdatedNumberScheduled
	desired := ds.Status.DesiredNumberScheduled
	ready := ds.Status.NumberReady

	log.V(1).Info(fmt.Sprintf("Checking DaemonSet: updated: %d, desired: %d, ready: %d, generation: %d, observed generation: %d",
		updated, desired, ready, generation, observedGeneration), "name", name.Name)

	return observedGeneration == generation && updated == desired && ready >= desired, nil
}

type DaemonSetAnnotator struct {
	client.Client
}

func (dsa *DaemonSetAnnotator) SetAnnotation(ctx context.Context, name types.NamespacedName, key, value string) (patched bool, err error) {
	var ds appsv1.DaemonSet
	if err = dsa.Get(ctx, name, &ds); err != nil {
		return false, fmt.Errorf("failed to get %s/%s DaemonSet: %v", name.Namespace, name.Name, err)
	}

	patchedDS := *ds.DeepCopy()
	if patchedDS.Spec.Template.ObjectMeta.Annotations == nil {
		patchedDS.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	} else if patchedDS.Spec.Template.ObjectMeta.Annotations[key] == value {
		return false, nil
	}

	patchedDS.Spec.Template.ObjectMeta.Annotations[key] = value

	if err = dsa.Patch(ctx, &patchedDS, client.MergeFrom(&ds)); err != nil {
		return false, fmt.Errorf("failed to patch %s/%s DaemonSet: %v", name.Namespace, name.Name, err)
	}

	return true, nil
}
