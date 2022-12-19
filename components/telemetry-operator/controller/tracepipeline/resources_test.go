package tracepipeline

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	config = Config{
		BaseName:  "collector",
		Namespace: "kyma-system",
		Service: ServiceConfig{
			OTLPServiceName: "otlp-traces",
		},
	}
	tracePipeline = v1alpha1.TracePipelineOutput{
		Otlp: &v1alpha1.OtlpOutput{
			Endpoint: v1alpha1.ValueType{
				Value: "localhost",
			},
		},
	}

	tracePipelineHTTP = v1alpha1.TracePipelineOutput{
		Otlp: &v1alpha1.OtlpOutput{
			Endpoint: v1alpha1.ValueType{
				Value: "http://localhost",
			},
		},
	}

	tracePipelineHTTPS = v1alpha1.TracePipelineOutput{
		Otlp: &v1alpha1.OtlpOutput{
			Endpoint: v1alpha1.ValueType{
				Value: "https://localhost",
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
	require.Equal(t, cm.Name, config.BaseName)
	require.Equal(t, cm.Namespace, config.Namespace)
	expectedEndpoint := fmt.Sprintf("endpoint: ${%s}", otlpEndpointVariable)
	collectorConfig := cm.Data[configMapKey]

	var collectorConfigYaml interface{}
	require.NoError(t, yaml.Unmarshal([]byte(collectorConfig), &collectorConfigYaml), "Otel Collector config must be valid yaml")
	require.True(t, strings.Contains(collectorConfig, expectedEndpoint), "Otel Collector config must contain OTLP endpoint")
}

func TestMakeConfigMapTLSInsecureNoScheme(t *testing.T) {
	cm := makeConfigMap(config, tracePipeline)

	require.NotNil(t, cm)
	collectorConfig := cm.Data[configMapKey]
	require.NotEmpty(t, collectorConfig)

	expectedTLSConfig := "insecure: false"
	require.True(t, strings.Contains(collectorConfig, expectedTLSConfig), "Otel Collector config must contain TLS insecure true")
}

func TestMakeConfigMapTLSInsecureHttp(t *testing.T) {
	cm := makeConfigMap(config, tracePipelineHTTP)

	require.NotNil(t, cm)
	collectorConfig := cm.Data[configMapKey]
	require.NotEmpty(t, collectorConfig)

	expectedTLSConfig := "insecure: true"
	require.True(t, strings.Contains(collectorConfig, expectedTLSConfig), "Otel Collector config must contain TLS insecure true")
}

func TestMakeConfigMapTLSInsecureHttps(t *testing.T) {
	cm := makeConfigMap(config, tracePipelineHTTPS)

	require.NotNil(t, cm)
	collectorConfig := cm.Data[configMapKey]
	require.NotEmpty(t, collectorConfig)

	expectedTLSConfig := "insecure: false"
	require.True(t, strings.Contains(collectorConfig, expectedTLSConfig), "Otel Collector config must contain TLS insecure false")
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
	require.Equal(t, secret.Name, config.BaseName)
	require.Equal(t, secret.Namespace, config.Namespace)

	require.Equal(t, "otlpEndpoint", string(secret.Data[otlpEndpointVariable]), "Secret must contain OTLP endpoint")
	require.Equal(t, "basicAuthHeader", string(secret.Data[basicAuthHeaderVariable]), "Secret must contain basic auth header")
}

func TestMakeDeployment(t *testing.T) {
	deployment := makeDeployment(config, "123")
	labels := makeDefaultLabels(config)

	require.NotNil(t, deployment)
	require.Equal(t, deployment.Name, config.BaseName)
	require.Equal(t, deployment.Namespace, config.Namespace)
	require.Equal(t, *deployment.Spec.Replicas, int32(1))
	require.Equal(t, deployment.Spec.Selector.MatchLabels, labels)
	require.Equal(t, deployment.Spec.Template.ObjectMeta.Labels, labels)
	for k, v := range defaultPodAnnotations {
		require.Equal(t, deployment.Spec.Template.ObjectMeta.Annotations[k], v)
	}
	require.Equal(t, deployment.Spec.Template.ObjectMeta.Annotations[configHashAnnotationKey], "123")
	require.NotEmpty(t, deployment.Spec.Template.Spec.Containers[0].EnvFrom)

	require.NotNil(t, deployment.Spec.Template.Spec.Containers[0].LivenessProbe, "liveness probe must be defined")
	require.NotNil(t, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe, "readiness probe must be defined")

	podSecurityContext := deployment.Spec.Template.Spec.SecurityContext
	require.NotNil(t, podSecurityContext, "pod security context must be defined")
	require.NotZero(t, podSecurityContext.RunAsUser, "must run as non-root")
	require.True(t, *podSecurityContext.RunAsNonRoot, "must run as non-root")

	containerSecurityContext := deployment.Spec.Template.Spec.Containers[0].SecurityContext
	require.NotNil(t, containerSecurityContext, "container security context must be defined")
	require.NotZero(t, containerSecurityContext.RunAsUser, "must run as non-root")
	require.True(t, *containerSecurityContext.RunAsNonRoot, "must run as non-root")
	require.False(t, *containerSecurityContext.Privileged, "must not be privileged")
	require.False(t, *containerSecurityContext.AllowPrivilegeEscalation, "must not escalate to privileged")
	require.True(t, *containerSecurityContext.ReadOnlyRootFilesystem, "must use readonly fs")
}

func TestMakeOTLPService(t *testing.T) {
	service := makeOTLPService(config)
	labels := makeDefaultLabels(config)

	require.NotNil(t, service)
	require.Equal(t, service.Name, config.Service.OTLPServiceName)
	require.Equal(t, service.Namespace, config.Namespace)
	require.Equal(t, service.Spec.Selector, labels)
	require.Equal(t, service.Spec.Type, corev1.ServiceTypeClusterIP)
	require.NotEmpty(t, service.Spec.Ports)
	require.Len(t, service.Spec.Ports, 2)
}

func TestMakeMetricsService(t *testing.T) {
	service := makeMetricsService(config)
	labels := makeDefaultLabels(config)

	require.NotNil(t, service)
	require.Equal(t, service.Name, config.BaseName+"-metrics")
	require.Equal(t, service.Namespace, config.Namespace)
	require.Equal(t, service.Spec.Selector, labels)
	require.Equal(t, service.Spec.Type, corev1.ServiceTypeClusterIP)
	require.NotEmpty(t, service.Spec.Ports)
	require.Len(t, service.Spec.Ports, 1)
}

func TestMakeOpenCensusService(t *testing.T) {
	service := makeOpenCensusService(config)
	labels := makeDefaultLabels(config)

	require.NotNil(t, service)
	require.Equal(t, service.Name, config.BaseName+"-internal")
	require.Equal(t, service.Namespace, config.Namespace)
	require.Equal(t, service.Spec.Selector, labels)
	require.Equal(t, service.Spec.Type, corev1.ServiceTypeClusterIP)
	require.NotEmpty(t, service.Spec.Ports)
	require.Len(t, service.Spec.Ports, 1)
}

func TestMakeServiceMonitor(t *testing.T) {
	serviceMonitor := makeServiceMonitor(config)
	labels := makeDefaultLabels(config)

	require.NotNil(t, serviceMonitor)
	require.Equal(t, serviceMonitor.Name, config.BaseName)
	require.Equal(t, serviceMonitor.Namespace, config.Namespace)
	require.Contains(t, serviceMonitor.Spec.NamespaceSelector.MatchNames, config.Namespace)
	require.Equal(t, serviceMonitor.Spec.Selector.MatchLabels, labels)
}
