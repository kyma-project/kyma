package tracepipeline

import (
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

var (
	config = Config{
		CollectorConfigMapName:  "collector-config",
		CollectorDeploymentName: "collector",
		CollectorNamespace:      "kyma-system",
		ConfigMapKey:            "key",
		Replicas:                1,
		PodSelectorLabels: map[string]string{
			"app.kubernetes.io/name": "collector",
		},
	}
	tracePipeline = v1alpha1.TracePipelineOutput{
		Otlp: v1alpha1.OtlpOutput{
			Endpoint: v1alpha1.ValueType{
				Value: "localhost",
			},
		},
	}
)

func TestMakeConfigMap(t *testing.T) {
	cm := makeConfigMap(config, tracePipeline)

	require.NotNil(t, cm)
	require.Equal(t, cm.Name, config.CollectorConfigMapName)
	require.Equal(t, cm.Namespace, config.CollectorNamespace)
	require.NotEmpty(t, cm.Data[config.ConfigMapKey])
}

func TestMakeDeployment(t *testing.T) {
	deployment := makeDeployment(config)

	require.NotNil(t, deployment)
	require.Equal(t, deployment.Name, config.CollectorDeploymentName)
	require.Equal(t, deployment.Namespace, config.CollectorNamespace)
	require.Equal(t, *deployment.Spec.Replicas, config.Replicas)
	require.Equal(t, deployment.Spec.Selector.MatchLabels, config.PodSelectorLabels)
	require.Equal(t, deployment.Spec.Template.ObjectMeta.Labels, config.PodSelectorLabels)
	require.Equal(t, deployment.Spec.Template.ObjectMeta.Annotations, config.PodAnnotations)
}

func TestMakeService(t *testing.T) {
	service := makeService(config)

	require.NotNil(t, service)
	require.Equal(t, service.Name, config.CollectorDeploymentName)
	require.Equal(t, service.Namespace, config.CollectorNamespace)
	require.Equal(t, service.Spec.Selector, config.PodSelectorLabels)
	require.NotEmpty(t, service.Spec.Ports)
}

func TestMakeServiceMonitor(t *testing.T) {
	serviceMonitor := makeServiceMonitor(config)

	require.NotNil(t, serviceMonitor)
	require.Equal(t, serviceMonitor.Name, config.CollectorDeploymentName)
	require.Equal(t, serviceMonitor.Namespace, config.CollectorNamespace)
	require.Contains(t, serviceMonitor.Spec.NamespaceSelector.MatchNames, config.CollectorNamespace)
	require.Equal(t, serviceMonitor.Spec.Selector.MatchLabels, config.PodSelectorLabels)
}
