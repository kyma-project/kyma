package tracepipeline

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

var (
	secret1 = corev1.Secret{
		Data: map[string][]byte{
			"1": []byte("2"),
			"3": []byte("4"),
		},
	}
	secret2 = corev1.Secret{
		Data: map[string][]byte{
			"2": []byte("1"),
			"4": []byte("3"),
		},
	}
	configMap1 = corev1.ConfigMap{
		Data: map[string]string{
			"a": "b",
			"c": "d",
		},
	}
	configMap2 = corev1.ConfigMap{
		Data: map[string]string{
			"b": "a",
			"d": "c",
		},
	}
	emptySecret = corev1.Secret{
		Data: map[string][]byte{},
	}
	emptyConfigMap = corev1.ConfigMap{
		Data: map[string]string{},
	}
)

func TestEqualConfig(t *testing.T) {
	hash1 := NewConfigHash([]corev1.ConfigMap{configMap1}, []corev1.Secret{secret1}).Build()
	hash2 := NewConfigHash([]corev1.ConfigMap{configMap1}, []corev1.Secret{secret1}).Build()
	require.Equal(t, hash1, hash2)
}

func TestUnequalConfig(t *testing.T) {
	hash1 := NewConfigHash([]corev1.ConfigMap{configMap1}, []corev1.Secret{secret1}).Build()
	hash2 := NewConfigHash([]corev1.ConfigMap{configMap2}, []corev1.Secret{secret2}).Build()
	require.NotEqual(t, hash1, hash2)
}

func TestEmptyConfig(t *testing.T) {
	hash := NewConfigHash([]corev1.ConfigMap{emptyConfigMap}, []corev1.Secret{emptySecret}).Build()
	require.NotEmpty(t, hash)
}
