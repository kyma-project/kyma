// Package testing provides support for automated testing of nats-operator-doctor.
package testing

import (
	"github.com/nats-io/nats-operator/pkg/apis/nats/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNatsCluster TODO ...
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

// GetNatsClusterIfNil TODO ...
func GetNatsClusterIfNil(natsCluster *v1alpha2.NatsCluster) *v1alpha2.NatsCluster {
	if natsCluster != nil {
		return natsCluster
	}
	return &v1alpha2.NatsCluster{}
}

// GetNatsClusterList TODO ...
func GetNatsClusterList(natsClusters ...v1alpha2.NatsCluster) *v1alpha2.NatsClusterList {
	return &v1alpha2.NatsClusterList{
		TypeMeta: getNatsClusterTypeMeta(),
		Items:    natsClusters[:],
	}
}

// getNatsClusterTypeMeta TODO ...
func getNatsClusterTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       v1alpha2.CRDResourceKind,
		APIVersion: v1alpha2.SchemeGroupVersion.String(),
	}
}
