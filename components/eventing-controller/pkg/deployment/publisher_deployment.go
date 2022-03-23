package deployment

import (
	"fmt"
	"strconv"

	"github.com/kyma-project/kyma/components/eventing-controller/utils"

	"k8s.io/apimachinery/pkg/api/resource"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

//nolint:gosec
const (
	livenessInitialDelaySecs = int32(5)
	livenessTimeoutSecs      = int32(1)
	livenessPeriodSecs       = int32(2)
	bebNamespacePrefix       = "/"
	instanceLabelKey         = "app.kubernetes.io/instance"
	instanceLabelValue       = "eventing"
	dashboardLabelKey        = "kyma-project.io/dashboard"
	dashboardLabelValue      = "eventing"
	publisherPortName        = "http"
	publisherPortNum         = int32(8080)
	publisherMetricsPortName = "http-metrics"
	publisherMetricsPortNum  = int32(9090)

	PublisherNamespace              = "kyma-system"
	PublisherName                   = "eventing-publisher-proxy"
	AppLabelKey                     = "app.kubernetes.io/name"
	PublisherSecretClientIDKey      = "client-id"
	PublisherSecretClientSecretKey  = "client-secret"
	PublisherSecretTokenEndpointKey = "token-endpoint"
	PublisherSecretEMSURLKey        = "ems-publish-url"
	PublisherSecretBEBNamespaceKey  = "beb-namespace"

	configMapName               = "eventing"
	configMapKeyEventTypePrefix = "eventTypePrefix"
)

var (
	TerminationGracePeriodSeconds = int64(30)
)

func NewBEBPublisherDeployment(publisherConfig env.PublisherConfig) *appsv1.Deployment {
	labels := map[string]string{
		AppLabelKey:       PublisherName,
		instanceLabelKey:  instanceLabelValue,
		dashboardLabelKey: dashboardLabelValue,
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
							Env:             getBEBEnvVars(publisherConfig),
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
					PriorityClassName:             publisherConfig.PriorityClassName,
				},
			},
		},
	}
}

func NewNATSPublisherDeployment(publisherConfig env.PublisherConfig) *appsv1.Deployment {
	labels := map[string]string{
		AppLabelKey:       PublisherName,
		instanceLabelKey:  instanceLabelValue,
		dashboardLabelKey: dashboardLabelValue,
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
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: v1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{AppLabelKey: PublisherName},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:            PublisherName,
							Image:           publisherConfig.Image,
							Ports:           getContainerPorts(),
							Env:             getNATSEnvVars(publisherConfig),
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
					PriorityClassName:             publisherConfig.PriorityClassName,
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
		ProbeHandler: v1.ProbeHandler{
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
		ProbeHandler: v1.ProbeHandler{
			HTTPGet: &v1.HTTPGetAction{
				Path:   "/healthz",
				Port:   intstr.FromInt(8080),
				Scheme: v1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: livenessInitialDelaySecs,
		TimeoutSeconds:      livenessTimeoutSecs,
		PeriodSeconds:       livenessPeriodSecs,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func getContainerPorts() []v1.ContainerPort {
	return []v1.ContainerPort{
		{
			Name:          publisherPortName,
			ContainerPort: publisherPortNum,
		},
		{
			Name:          publisherMetricsPortName,
			ContainerPort: publisherMetricsPortNum,
		},
	}
}

func getBEBEnvVars(publisherConfig env.PublisherConfig) []v1.EnvVar {
	return []v1.EnvVar{
		{Name: "BACKEND", Value: "beb"},
		{Name: "PORT", Value: strconv.Itoa(int(publisherPortNum))},
		{
			Name: "EVENT_TYPE_PREFIX",
			ValueFrom: &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: configMapName,
					},
					Key: configMapKeyEventTypePrefix,
				},
			},
		},
		{Name: "REQUEST_TIMEOUT", Value: publisherConfig.RequestTimeout},
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
			Value: fmt.Sprintf("%s$(BEB_NAMESPACE_VALUE)", bebNamespacePrefix),
		},
	}
}

func getNATSEnvVars(publisherConfig env.PublisherConfig) []v1.EnvVar {
	return []v1.EnvVar{
		{Name: "BACKEND", Value: "nats"},
		{Name: "PORT", Value: strconv.Itoa(int(publisherPortNum))},
		{Name: "NATS_URL", Value: "eventing-nats.kyma-system.svc.cluster.local"},
		{Name: "REQUEST_TIMEOUT", Value: publisherConfig.RequestTimeout},
		{Name: "LEGACY_NAMESPACE", Value: "kyma"},
		{
			Name: "LEGACY_EVENT_TYPE_PREFIX",
			ValueFrom: &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: configMapName,
					},
					Key: configMapKeyEventTypePrefix,
				},
			},
		},
		{
			Name: "EVENT_TYPE_PREFIX",
			ValueFrom: &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: configMapName,
					},
					Key: configMapKeyEventTypePrefix,
				},
			},
		},
		{Name: "ENABLE_JETSTREAM_BACKEND", Value: strconv.FormatBool(publisherConfig.EnableJetStreamBackend)},
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
