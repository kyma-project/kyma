package backend

import (
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ServiceAccountName       = "eventing-event-publisher-nats"
	LivenessInitialDelaySecs = int32(5)
	LivenessTimeoutSecs      = int32(1)
	LivenessPeriodSecs       = int32(2)
	BEBNamespacePrefix       = "/"
	InstanceLabelKey         = "app.kubernetes.io/instance"
	InstanceLabelValue       = "eventing"
	DashboardLabelKey        = "kyma-project.io/dashboard"
	DashboardLabelValue      = "eventing"
	PublisherReplicas        = 1
	PublisherImage           = "eu.gcr.io/kyma-project/event-publisher-proxy:88360eed"
	PublisherPortName        = "http"
	PublisherPortNum         = int32(8080)
	PublisherMetricsPortName = "http-metrics"
	PublisherMetricsPortNum  = int32(9090)
)

var (
	TerminationGracePeriodSeconds = int64(30)
)

func newBEBPublisherDeployment() *appsv1.Deployment {
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
			Replicas: intPtr(PublisherReplicas),
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
							Image:           PublisherImage,
							Ports:           getContainerPorts(),
							Env:             getBEBEnvVars(),
							LivenessProbe:   getLivenessProbe(),
							ReadinessProbe:  getReadinessProbe(),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: getSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyAlways,
					ServiceAccountName:            ServiceAccountName,
					TerminationGracePeriodSeconds: &TerminationGracePeriodSeconds,
				},
			},
		},
	}
}

func newNATSPublisherDeployment() *appsv1.Deployment {
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
			Replicas: intPtr(PublisherReplicas),
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
							Image:           PublisherImage,
							Ports:           getContainerPorts(),
							Env:             getNATSEnvVars(),
							LivenessProbe:   getLivenessProbe(),
							ReadinessProbe:  getReadinessProbe(),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: getSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyAlways,
					ServiceAccountName:            ServiceAccountName,
					TerminationGracePeriodSeconds: &TerminationGracePeriodSeconds,
				},
			},
		},
		Status: appsv1.DeploymentStatus{},
	}
}

func getSecurityContext() *v1.SecurityContext {
	return &v1.SecurityContext{
		Privileged:               boolPtr(false),
		AllowPrivilegeEscalation: boolPtr(false),
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
					Key:                  "client-id",
				}},
		},
		{
			Name: "CLIENT_SECRET",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: PublisherName},
					Key:                  "client-secret",
				}},
		},
		{
			Name: "TOKEN_ENDPOINT",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: PublisherName},
					Key:                  "token-endpoint",
				}},
		},
		{
			Name: "EMS_PUBLISH_URL",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: PublisherName},
					Key:                  "ems-publish-url",
				}},
		},
		{
			Name: "BEB_NAMESPACE_VALUE",
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: PublisherName},
					Key:                  "beb-namespace",
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
