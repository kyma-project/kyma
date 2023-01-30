package kubernetes

import (
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

type DeploymentProber struct {
	client.Client
}

func (dp *DeploymentProber) IsReady(ctx context.Context, name types.NamespacedName) (bool, error) {
	var d v1.Deployment
	if err := dp.Get(ctx, name, &d); err != nil {
		return false, fmt.Errorf("failed to get %s/%s Deployment: %v", name.Namespace, name.Name, err)
	}

	desiredReplicas := *d.Spec.Replicas
	var allReplicaSets v1.ReplicaSetList

	listOps := &client.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(d.Spec.Selector.MatchLabels),
		Namespace:     d.Namespace,
	}
	if err := dp.List(ctx, &allReplicaSets, listOps); err != nil {
		return false, fmt.Errorf("failed to list ReplicaSets: %v", err)
	}

	if err := dp.Get(ctx, name, &d); err != nil {
		return false, fmt.Errorf("failed to get %s/%s ReplicaSet for deployment: %v", name.Namespace, name.Name, err)
	}

	replicaSet, err := getLatestReplicaSet(&d, &allReplicaSets)
	if err != nil || replicaSet == nil {
		return false, fmt.Errorf("failed to get latest ReplicaSet: %v", err)
	}

	isReady := replicaSet.Status.ReadyReplicas >= desiredReplicas
	return isReady, nil
}

func getLatestReplicaSet(deployment *v1.Deployment, allReplicaSets *v1.ReplicaSetList) (*v1.ReplicaSet, error) {
	var ownedReplicaSets []*v1.ReplicaSet
	for i := range allReplicaSets.Items {
		if metav1.IsControlledBy(&allReplicaSets.Items[i], deployment) {
			ownedReplicaSets = append(ownedReplicaSets, &allReplicaSets.Items[i])
		}
	}

	if len(ownedReplicaSets) == 0 {
		return nil, nil
	}

	return findNewReplicaSet(deployment, ownedReplicaSets), nil
}

// findNewReplicaSet returns the new RS this given deployment targets (the one with the same pod template).
func findNewReplicaSet(deployment *v1.Deployment, rsList []*v1.ReplicaSet) *v1.ReplicaSet {
	sort.Sort(replicaSetsByCreationTimestamp(rsList))
	for i := range rsList {
		if equalIgnoreHash(&rsList[i].Spec.Template, &deployment.Spec.Template) {
			// In rare cases, such as after cluster upgrades, Deployment may end up with
			// having more than one new ReplicaSets that have the same template as its template,
			// see https://github.com/kubernetes/kubernetes/issues/40415
			// We deterministically choose the oldest new ReplicaSet.
			return rsList[i]
		}
	}
	// new ReplicaSet does not exist.
	return nil
}

func equalIgnoreHash(template1, template2 *corev1.PodTemplateSpec) bool {
	t1Copy := template1.DeepCopy()
	t2Copy := template2.DeepCopy()
	delete(t1Copy.Labels, v1.DefaultDeploymentUniqueLabelKey)
	delete(t2Copy.Labels, v1.DefaultDeploymentUniqueLabelKey)
	return apiequality.Semantic.DeepEqual(t1Copy, t2Copy)
}

type replicaSetsByCreationTimestamp []*v1.ReplicaSet

func (o replicaSetsByCreationTimestamp) Len() int      { return len(o) }
func (o replicaSetsByCreationTimestamp) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o replicaSetsByCreationTimestamp) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(&o[j].CreationTimestamp)
}
