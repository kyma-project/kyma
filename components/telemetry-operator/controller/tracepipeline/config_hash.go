package tracepipeline

import (
	"crypto/sha256"
	"fmt"
	"hash"

	corev1 "k8s.io/api/core/v1"
)

type ConfigHash struct {
	hash hash.Hash
}

func NewConfigHash(configMaps []corev1.ConfigMap, secrets []corev1.Secret) *ConfigHash {
	configHash := ConfigHash{
		hash: sha256.New(),
	}

	for _, cm := range configMaps {
		configHash.addStringMap(cm.Data)
	}

	for _, secret := range secrets {
		configHash.addByteMap(secret.Data)
	}

	return &configHash
}

func (c *ConfigHash) addStringMap(m map[string]string) {
	for k, v := range m {
		c.hash.Write([]byte(k))
		c.hash.Write([]byte(v))
	}
}

func (c *ConfigHash) addByteMap(m map[string][]byte) {
	for k, v := range m {
		c.hash.Write([]byte(k))
		c.hash.Write(v)
	}
}

func (c *ConfigHash) Build() string {
	return fmt.Sprintf("%x", c.hash.Sum(nil))
}
