package kubernetes

import (
	"context"
	"fmt"
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

type ChecksumParams struct {
	ConfigMapNames   []string
	SecretNames      []string
	AnnotationSuffix string
}

func NewDaemonSetHelper(client client.Client, restartsTotal prometheus.Counter) *DaemonSetHelper {
	return &DaemonSetHelper{
		client:        client,
		restartsTotal: restartsTotal,
	}
}

// UpdateConfigChecksum deletes all Fluent Bit pods to apply the new configuration
func (f *DaemonSetHelper) UpdateConfigChecksum(ctx context.Context, daemonSet types.NamespacedName, params *ChecksumParams) error {
	var ds appsv1.DaemonSet
	if err := f.client.Get(ctx, daemonSet, &ds); err != nil {
		return fmt.Errorf("failed to get %s/%s DaemonSet: %v", daemonSet.Namespace, daemonSet.Name, err)
	}

	patchedDS := *ds.DeepCopy()
	if patchedDS.Spec.Template.ObjectMeta.Annotations == nil {
		patchedDS.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	checksum, err := f.calculateConfigChecksum(params)
	if err != nil {
		return fmt.Errorf("failed to calculate config checksum: %v", err)
	}

	annotation := fmt.Sprintf("checksum/%s", params.AnnotationSuffix)
	patchedDS.Spec.Template.ObjectMeta.Annotations[annotation] = checksum

	if err := f.client.Patch(ctx, &patchedDS, client.MergeFrom(&ds)); err != nil {
		return fmt.Errorf("failed to patch %s/%s DaemonSet: %v", daemonSet.Namespace, daemonSet.Name, err)
	}
	f.restartsTotal.Inc()
	return nil
}

func (f *DaemonSetHelper) calculateConfigChecksum(params *ChecksumParams) (string, error) {
	return "", nil
}

func (f *DaemonSetHelper) IsReady(ctx context.Context, daemonSet types.NamespacedName) (bool, error) {
	log := logf.FromContext(ctx)

	var ds appsv1.DaemonSet
	if err := f.client.Get(ctx, daemonSet, &ds); err != nil {
		return false, fmt.Errorf("failed to get %s/%s DaemonSet: %v", daemonSet.Namespace, daemonSet.Name, err)
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
