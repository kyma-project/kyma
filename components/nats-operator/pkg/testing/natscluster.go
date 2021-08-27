// Package testing provides support for automated testing of nats-operator-doctor.
package testing

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nats-io/nats-operator/pkg/apis/nats/v1alpha2"
)

// GetNatsCluster returns a new NatsCluster instance.
func GetNatsCluster(name, namespace string, size int) *v1alpha2.NatsCluster {
	return &v1alpha2.NatsCluster{
		TypeMeta: getNatsClusterTypeMeta(),
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha2.ClusterSpec{
			Size: size,
		},
	}
}

// GetNatsClusterIfNil returns a new NatsCluster instance if the given one is nil.
func GetNatsClusterIfNil(natsCluster *v1alpha2.NatsCluster) *v1alpha2.NatsCluster {
	if natsCluster != nil {
		return natsCluster
	}
	return &v1alpha2.NatsCluster{}
}

// GetNatsClusterList returns a new NatsClusterList instance.
func GetNatsClusterList(natsClusters ...v1alpha2.NatsCluster) *v1alpha2.NatsClusterList {
	return &v1alpha2.NatsClusterList{
		TypeMeta: getNatsClusterTypeMeta(),
		Items:    natsClusters[:],
	}
}

// getNatsClusterTypeMeta returns a new TypeMeta instance for the NatsCluster type.
func getNatsClusterTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       v1alpha2.CRDResourceKind,
		APIVersion: v1alpha2.SchemeGroupVersion.String(),
	}
}
