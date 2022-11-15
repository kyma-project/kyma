package hash

import (
	"crypto/sha256"
	"fmt"
	"hash"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"

	corev1 "k8s.io/api/core/v1"
)

type objectMetaGetter interface {
	GetObjectMeta() metav1.Object
}

func Calculate(configMaps []corev1.ConfigMap, secrets []corev1.Secret) string {
	h := sha256.New()

	for _, cm := range sortConfigMaps(configMaps) {
		addStringMap(h, cm.Data)
	}

	for _, secret := range sortSecrets(secrets) {
		addByteMap(h, secret.Data)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func addStringMap(h hash.Hash, m map[string]string) {
	for k, v := range m {
		h.Write([]byte(k))
		h.Write([]byte(v))
	}
}

func addByteMap(h hash.Hash, m map[string][]byte) {
	for k, v := range m {
		h.Write([]byte(k))
		h.Write(v)
	}
}

func sortConfigMaps(unsorted []corev1.ConfigMap) []corev1.ConfigMap {
	sorted := make([]corev1.ConfigMap, len(unsorted))
	copy(sorted, unsorted)
	sort.Slice(sorted, func(i, j int) bool {
		return less(&sorted[i].ObjectMeta, &sorted[j].ObjectMeta)
	})
	return sorted
}

func sortSecrets(unsorted []corev1.Secret) []corev1.Secret {
	sorted := make([]corev1.Secret, len(unsorted))
	copy(sorted, unsorted)
	sort.Slice(sorted, func(i, j int) bool {
		return less(&sorted[i].ObjectMeta, &sorted[j].ObjectMeta)
	})
	return sorted
}

func less(x, y objectMetaGetter) bool {
	xNamespace, yNamespace := x.GetObjectMeta().GetNamespace(), y.GetObjectMeta().GetNamespace()
	if xNamespace != yNamespace {
		return xNamespace < yNamespace
	}
	return x.GetObjectMeta().GetName() < y.GetObjectMeta().GetName()
}
