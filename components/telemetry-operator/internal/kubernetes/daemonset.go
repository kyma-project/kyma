package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type DaemonSetHelper struct {
	client        client.Client
	daemonSet     types.NamespacedName
	restartsTotal prometheus.Counter
}

func NewDaemonSetHelper(client client.Client, restartsTotal prometheus.Counter) *DaemonSetHelper {
	return &DaemonSetHelper{
		client:        client,
		restartsTotal: restartsTotal,
	}
}

// Restart deletes all Fluent Bit pods to apply the new configuration
func (f *DaemonSetHelper) Restart(ctx context.Context, daemonSet types.NamespacedName) error {
	log := logf.FromContext(ctx)

	var ds appsv1.DaemonSet
	if err := f.client.Get(ctx, daemonSet, &ds); err != nil {
		log.Error(err, "Failed to get DaemonSet", "name", f.daemonSet)
		return err
	}

	patchedDS := *ds.DeepCopy()
	if patchedDS.Spec.Template.ObjectMeta.Annotations == nil {
		patchedDS.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	patchedDS.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	if err := f.client.Patch(ctx, &patchedDS, client.MergeFrom(&ds)); err != nil {
		log.Error(err, "Failed to patch DaemonSet to trigger rolling update", "name", f.daemonSet)
		return err
	}
	f.restartsTotal.Inc()
	return nil
}

func (f *DaemonSetHelper) IsReady(ctx context.Context, daemonSet types.NamespacedName) (bool, error) {
	log := logf.FromContext(ctx)

	var ds appsv1.DaemonSet
	if err := f.client.Get(ctx, daemonSet, &ds); err != nil {
		log.Error(err, "Failed getting DaemonSet", "name", f.daemonSet)
		return false, err
	}

	generation := ds.Generation
	observedGeneration := ds.Status.ObservedGeneration
	updated := ds.Status.UpdatedNumberScheduled
	desired := ds.Status.DesiredNumberScheduled
	ready := ds.Status.NumberReady

	log.V(1).Info(fmt.Sprintf("Checking DaemonSet: updated: %d, desired: %d, ready: %d, generation: %d, observed generation: %d",
		updated, desired, ready, generation, observedGeneration), "name", f.daemonSet)

	return observedGeneration == generation && updated == desired && ready >= desired, nil
}
