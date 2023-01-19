package kubernetes

import (
	"context"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestConfigMapProber(t *testing.T) {
	conf := `
global:
  logLevel: info
tracing:
  paused: true
`
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "kyma-system"},
		Data:       map[string]string{"override-config": conf},
	}
	fakeClient := fake.NewClientBuilder().WithObjects(configMap).Build()
	sut := ConfigmapProber{fakeClient}
	cm, err := sut.ReadConfigMapOrEmpty(context.Background(), types.NamespacedName{Name: "foo", Namespace: "kyma-system"})
	require.NoError(t, err)
	require.Equal(t, conf, cm)
}

func TestConfigMapNotExist(t *testing.T) {
	fakeClient := fake.NewClientBuilder().Build()
	sut := ConfigmapProber{fakeClient}
	cm, err := sut.ReadConfigMapOrEmpty(context.Background(), types.NamespacedName{Name: "foo", Namespace: "kyma-system"})
	require.NoError(t, err)
	require.Equal(t, "", cm)
}
