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
	expectedCM := make(map[string]interface{})
	tracingCM := make(map[string]interface{})
	globalCM := make(map[string]interface{})
	globalCM["logLevel"] = "info"
	tracingCM["paused"] = true
	expectedCM["global"] = globalCM
	expectedCM["tracing"] = tracingCM

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "kyma-system"},
		Data:       map[string]string{"override-config": conf},
	}
	fakeClient := fake.NewClientBuilder().WithObjects(configMap).Build()
	sut := ConfigmapProber{fakeClient}
	cm, err := sut.IsPresent(context.Background(), types.NamespacedName{Name: "foo", Namespace: "kyma-system"})
	require.NoError(t, err)
	require.Equal(t, expectedCM, cm)
}

func TestConfigMapNotExist(t *testing.T) {
	expectedCM := make(map[string]interface{})
	fakeClient := fake.NewClientBuilder().Build()
	sut := ConfigmapProber{fakeClient}
	cm, err := sut.IsPresent(context.Background(), types.NamespacedName{Name: "foo", Namespace: "kyma-system"})
	require.NoError(t, err)
	require.Equal(t, expectedCM, cm)
}
