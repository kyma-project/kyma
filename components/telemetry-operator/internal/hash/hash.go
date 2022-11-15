package hash

import (
	"crypto/sha256"
	"fmt"
	"hash"

	corev1 "k8s.io/api/core/v1"
)

func Calculate(configMaps []corev1.ConfigMap, secrets []corev1.Secret) string {
	h := sha256.New()

	for _, cm := range configMaps {
		addStringMap(h, cm.Data)
	}

	for _, secret := range secrets {
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
