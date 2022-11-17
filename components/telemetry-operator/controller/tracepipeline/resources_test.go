package tracepipeline

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
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

	tracePipelineWithBasicAuth = v1alpha1.TracePipelineOutput{
		Otlp: &v1alpha1.OtlpOutput{
			Endpoint: v1alpha1.ValueType{
				Value: "localhost",
			},
			Authentication: &v1alpha1.AuthenticationOptions{
				Basic: &v1alpha1.BasicAuthOptions{
					User: v1alpha1.ValueType{
						Value: "user",
					},
					Password: v1alpha1.ValueType{
						Value: "password",
					},
				},
			},
		},
	}
)

func TestMakeConfigMap(t *testing.T) {
	cm := makeConfigMap(config, tracePipeline)

	require.NotNil(t, cm)
	require.Equal(t, cm.Name, config.ResourceName)
	require.Equal(t, cm.Namespace, config.CollectorNamespace)
	expectedEndpoint := fmt.Sprintf("endpoint: ${%s}", otlpEndpointVariable)
	collectorConfig := cm.Data[configMapKey]

	var collectorConfigYaml interface{}
	require.NoError(t, yaml.Unmarshal([]byte(collectorConfig), &collectorConfigYaml), "Otel Collector config must be valid yaml")
	require.True(t, strings.Contains(collectorConfig, expectedEndpoint), "Otel Collector config must contain OTLP endpoint")
}

func TestMakeConfigMapWithBasicAuth(t *testing.T) {
	cm := makeConfigMap(config, tracePipelineWithBasicAuth)

	require.NotNil(t, cm)
	collectorConfigString := cm.Data[configMapKey]
	require.NotEmpty(t, collectorConfigString)

	expectedAuthHeader := "Authorization: ${BASIC_AUTH_HEADER}"
	require.True(t, strings.Contains(collectorConfigString, expectedAuthHeader))
}

func TestMakeSecret(t *testing.T) {
	secretData := map[string][]byte{
		basicAuthHeaderVariable: []byte("basicAuthHeader"),
		otlpEndpointVariable:    []byte("otlpEndpoint"),
	}
	secret := makeSecret(config, secretData)

	require.NotNil(t, secret)
	require.Equal(t, secret.Name, config.ResourceName)
	require.Equal(t, secret.Namespace, config.CollectorNamespace)

	require.Equal(t, "otlpEndpoint", string(secret.Data[otlpEndpointVariable]), "Secret must contain OTLP endpoint")
	require.Equal(t, "basicAuthHeader", string(secret.Data[basicAuthHeaderVariable]), "Secret must contain basic auth header")
}

func TestMakeDeployment(t *testing.T) {
	deployment := makeDeployment(config, "123")
	labels := getLabels(config)

	require.NotNil(t, deployment)
	require.Equal(t, deployment.Name, config.ResourceName)
	require.Equal(t, deployment.Namespace, config.CollectorNamespace)
	require.Equal(t, *deployment.Spec.Replicas, int32(1))
	require.Equal(t, deployment.Spec.Selector.MatchLabels, labels)
	require.Equal(t, deployment.Spec.Template.ObjectMeta.Labels, labels)
	for k, v := range defaultPodAnnotations {
		require.Equal(t, deployment.Spec.Template.ObjectMeta.Annotations[k], v)
	}
	require.Equal(t, deployment.Spec.Template.ObjectMeta.Annotations[configHashAnnotationKey], "123")
	require.NotEmpty(t, deployment.Spec.Template.Spec.Containers[0].EnvFrom)
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
