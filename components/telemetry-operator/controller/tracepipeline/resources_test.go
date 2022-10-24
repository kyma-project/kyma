package tracepipeline

import (
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
)

var (
	config = Config{
		ResourceName:       "collector",
		CollectorNamespace: "kyma-system",
	}
	tracePipeline = v1alpha1.TracePipelineOutput{
		Otlp: &v1alpha1.OtlpOutput{
			Endpoint: v1alpha1.ValueType{
				Value: "localhost",
			},
		},
	}
)

func TestMakeConfigMap(t *testing.T) {
	cm := makeConfigMap(config, tracePipeline)

	require.NotNil(t, cm)
	require.Equal(t, cm.Name, config.ResourceName)
	require.Equal(t, cm.Namespace, config.CollectorNamespace)
	require.NotEmpty(t, cm.Data[configMapKey])
}

func TestMakeDeployment(t *testing.T) {
	deployment := makeDeployment(config)
	labels := getLabels(config)

	require.NotNil(t, deployment)
	require.Equal(t, deployment.Name, config.ResourceName)
	require.Equal(t, deployment.Namespace, config.CollectorNamespace)
	require.Equal(t, *deployment.Spec.Replicas, int32(1))
	require.Equal(t, deployment.Spec.Selector.MatchLabels, labels)
	require.Equal(t, deployment.Spec.Template.ObjectMeta.Labels, labels)
	require.Equal(t, deployment.Spec.Template.ObjectMeta.Annotations, podAnnotations)
}

func TestMakeCollectorService(t *testing.T) {
	service := makeCollectorService(config)
	labels := getLabels(config)

	require.NotNil(t, service)
	require.Equal(t, service.Name, config.ResourceName)
	require.Equal(t, service.Namespace, config.CollectorNamespace)
	require.Equal(t, service.Spec.Selector, labels)
	require.NotEmpty(t, service.Spec.Ports)
}

func TestMakeServiceMonitor(t *testing.T) {
	serviceMonitor := makeServiceMonitor(config)
	labels := getLabels(config)

	require.NotNil(t, serviceMonitor)
	require.Equal(t, serviceMonitor.Name, config.ResourceName)
	require.Equal(t, serviceMonitor.Namespace, config.CollectorNamespace)
	require.Contains(t, serviceMonitor.Spec.NamespaceSelector.MatchNames, config.CollectorNamespace)
	require.Equal(t, serviceMonitor.Spec.Selector.MatchLabels, labels)
}

func TestMakeMetricsService(t *testing.T) {
	service := makeMetricsService(config)
	labels := getLabels(config)

	require.NotNil(t, service)
	require.Equal(t, service.Name, config.ResourceName+"-metrics")
	require.Equal(t, service.Namespace, config.CollectorNamespace)
	require.Equal(t, service.Spec.Selector, labels)
	require.NotEmpty(t, service.Spec.Ports)
}
