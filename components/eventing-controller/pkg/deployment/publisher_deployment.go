package deployment

import (
	"fmt"
	"strconv"

	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	LivenessInitialDelaySecs = int32(5)
	LivenessTimeoutSecs      = int32(1)
	LivenessPeriodSecs       = int32(2)
	BEBNamespacePrefix       = "/"
	InstanceLabelKey         = "app.kubernetes.io/instance"
	InstanceLabelValue       = "eventing"
	DashboardLabelKey        = "kyma-project.io/dashboard"
	DashboardLabelValue      = "eventing"
	PublisherPortName        = "http"
	PublisherPortNum         = int32(8080)
	PublisherMetricsPortName = "http-metrics"
	PublisherMetricsPortNum  = int32(9090)

	PublisherNamespace              = "kyma-system"
	PublisherName                   = "eventing-publisher-proxy"
	AppLabelKey                     = "app.kubernetes.io/name"
	PublisherSecretClientIDKey      = "client-id"
	PublisherSecretClientSecretKey  = "client-secret"
	PublisherSecretTokenEndpointKey = "token-endpoint"
	PublisherSecretEMSURLKey        = "ems-publish-url"
	PublisherSecretBEBNamespaceKey  = "beb-namespace"
)

var (
	TerminationGracePeriodSeconds = int64(30)
)

func NewBEBPublisherDeployment(publisherConfig env.PublisherConfig) *appsv1.Deployment {
	labels := map[string]string{
		AppLabelKey:       PublisherName,
		InstanceLabelKey:  InstanceLabelValue,
		DashboardLabelKey: DashboardLabelValue,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PublisherName,
			Namespace: PublisherNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{

			Replicas: utils.Int32Ptr(publisherConfig.Replicas),
			Selector: metav1.SetAsLabelSelector(labels),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   PublisherName,
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            PublisherName,
							Image:           publisherConfig.Image,
							Ports:           getContainerPorts(),
							Env:             getBEBEnvVars(),
							LivenessProbe:   getLivenessProbe(),
							ReadinessProbe:  getReadinessProbe(),
							ImagePullPolicy: getImagePullPolicy(publisherConfig.ImagePullPolicy),
							SecurityContext: getSecurityContext(),
							Resources: getResources(publisherConfig.RequestsCPU,
								publisherConfig.RequestsMemory,
								publisherConfig.LimitsCPU,
								publisherConfig.LimitsMemory),
						},
					},
					RestartPolicy:                 v1.RestartPolicyAlways,
					ServiceAccountName:            publisherConfig.ServiceAccount,
					TerminationGracePeriodSeconds: &TerminationGracePeriodSeconds,
				},
			},
		},
	}
}

func NewNATSPublisherDeployment(publisherConfig env.PublisherConfig) *appsv1.Deployment {
	labels := map[string]string{
		AppLabelKey:       PublisherName,
		InstanceLabelKey:  InstanceLabelValue,
		DashboardLabelKey: DashboardLabelValue,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PublisherName,
			Namespace: PublisherNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{

			Replicas: utils.Int32Ptr(publisherConfig.Replicas),
			Selector: metav1.SetAsLabelSelector(labels),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   PublisherName,
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            PublisherName,
							Image:           publisherConfig.Image,
							Ports:           getContainerPorts(),
							Env:             getNATSEnvVars(),
							LivenessProbe:   getLivenessProbe(),
							ReadinessProbe:  getReadinessProbe(),
							ImagePullPolicy: getImagePullPolicy(publisherConfig.ImagePullPolicy),
							SecurityContext: getSecurityContext(),
							Resources: getResources(publisherConfig.RequestsCPU,
								publisherConfig.RequestsMemory,
								publisherConfig.LimitsCPU,
								publisherConfig.LimitsMemory),
						},
					},
					RestartPolicy:                 v1.RestartPolicyAlways,
					ServiceAccountName:            publisherConfig.ServiceAccount,
					TerminationGracePeriodSeconds: &TerminationGracePeriodSeconds,
				},
			},
		},
		Status: appsv1.DeploymentStatus{},
	}
}

func getImagePullPolicy(imagePullPolicy string) v1.PullPolicy {
	switch imagePullPolicy {
	case "IfNotPresent":
		return v1.PullIfNotPresent
	case "Always":
		return v1.PullAlways
	case "Never":
		return v1.PullNever
	default:
		return v1.PullIfNotPresent
	}
}

func getSecurityContext() *v1.SecurityContext {
	return &v1.SecurityContext{
		Privileged:               utils.BoolPtr(false),
		AllowPrivilegeEscalation: utils.BoolPtr(false),
	}
}

func getReadinessProbe() *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   "/readyz",
				Port:   intstr.FromInt(8080),
				Scheme: v1.URISchemeHTTP,
			},
		},
		FailureThreshold: 3,
	}
}

func getLivenessProbe() *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   "/healthz",
				Port:   intstr.FromInt(8080),
				Scheme: v1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: LivenessInitialDelaySecs,
		TimeoutSeconds:      LivenessTimeoutSecs,
		PeriodSeconds:       LivenessPeriodSecs,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func getContainerPorts() []v1.ContainerPort {
	return []v1.ContainerPort{
		{
			Name:          PublisherPortName,
			ContainerPort: PublisherPortNum,
		},
		{
			Name:          PublisherMetricsPortName,
			ContainerPort: PublisherMetricsPortNum,
		},
	}
}

func getBEBEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		{Name: "BACKEND", Value: "beb"},
		{Name: "PORT", Value: strconv.Itoa(int(PublisherPortNum))},
		{Name: "REQUEST_TIMEOUT", Value: "5s"},
		{Name: "EVENT_TYPE_PREFIX", Value: "sap.kyma.custom"},
		{
			Name: "CLIENT_ID",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: PublisherName},
					Key:                  PublisherSecretClientIDKey,
				}},
		},
		{
			Name: "CLIENT_SECRET",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: PublisherName},
					Key:                  PublisherSecretClientSecretKey,
				}},
		},
		{
			Name: "TOKEN_ENDPOINT",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: PublisherName},
					Key:                  PublisherSecretTokenEndpointKey,
				}},
		},
		{
			Name: "EMS_PUBLISH_URL",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: PublisherName},
					Key:                  PublisherSecretEMSURLKey,
				}},
		},
		{
			Name: "BEB_NAMESPACE_VALUE",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: PublisherName},
					Key:                  PublisherSecretBEBNamespaceKey,
				}},
		},
		{
			Name:  "BEB_NAMESPACE",
			Value: fmt.Sprintf("%s$(BEB_NAMESPACE_VALUE)", BEBNamespacePrefix),
		},
	}
}

func getNATSEnvVars() []v1.EnvVar {
	return []v1.EnvVar{
		{Name: "BACKEND", Value: "nats"},
		{Name: "PORT", Value: strconv.Itoa(int(PublisherPortNum))},
		{Name: "NATS_URL", Value: "eventing-nats.kyma-system.svc.cluster.local"},
		{Name: "REQUEST_TIMEOUT", Value: "5s"},
		{Name: "LEGACY_NAMESPACE", Value: "kyma"},
		{Name: "LEGACY_EVENT_TYPE_PREFIX", Value: "sap.kyma.custom"},
		{Name: "EVENT_TYPE_PREFIX", Value: "sap.kyma.custom"},
	}
}

func getResources(requestsCPU, requestsMemory, limitsCPU, limitsMemory string) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(requestsCPU),
			v1.ResourceMemory: resource.MustParse(requestsMemory),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(limitsCPU),
			v1.ResourceMemory: resource.MustParse(limitsMemory),
		},
	}
}
