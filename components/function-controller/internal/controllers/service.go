package controllers

import (
	"reflect"

	"github.com/go-logr/logr"
	funcerr "github.com/kyma-project/kyma/components/function-controller/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/reconciler/route/config"
)

var (
	errLambdaNotFound   = funcerr.NewInvalidValue("lambda container not found")
	errImageTagNotFound = funcerr.NewInvalidValue("image tag not found")
	errInvalidPodSpec   = funcerr.NewInvalidState("invalid pod specification")
	envVarsForRevision  = []corev1.EnvVar{
		{
			Name:  "FUNC_HANDLER",
			Value: "main",
		},
		{
			Name:  "MOD_NAME",
			Value: "handler",
		},
		{
			Name:  "FUNC_TIMEOUT",
			Value: "180",
		},
		{
			Name:  "FUNC_RUNTIME",
			Value: "nodejs8",
		},
		{
			Name:  "FUNC_MEMORY_LIMIT",
			Value: "128Mi",
		},
		{
			Name:  "FUNC_PORT",
			Value: "8080",
		},
		{
			Name:  "NODE_PATH",
			Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules",
		},
	}
)

func applyClusterLocalVisibleLabel(fnLabels map[string]string) map[string]string {
	newLabels := make(map[string]string)
	for key, value := range fnLabels {
		newLabels[key] = value
	}
	newLabels[config.VisibilityLabelKey] = config.VisibilityClusterLocal
	return newLabels
}

func newService(
	name, namespace string,
	podspec *corev1.PodSpec,
	fnLabels map[string]string) *servingv1.Service {
	svcLabels := applyClusterLocalVisibleLabel(fnLabels)

	return &servingv1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    svcLabels,
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{
						PodSpec: *podspec,
					},
				},
			},
		},
	}
}

// newPodSpec creates spec for knative service;
// image is a Docker image name; secret is a name of a secret in the same
// namespace to use for pulling any of the images used by this PodSpec;
// service account is is the name of the ServiceAccount to use to run pod
func newPodSpec(
	image, secret, serviceAccount string,
	envs ...corev1.EnvVar) *corev1.PodSpec {
	return &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "lambda",
				Image: image,
				Env:   append(envVarsForRevision, envs...),
			},
		},
		ServiceAccountName: serviceAccount,
	}
}

func lambdaEvs(svc *servingv1.Service) (*[]corev1.EnvVar, error) {
	index := -1
	for i := 0; i < len(svc.Spec.Template.Spec.Containers); i++ {
		c := svc.Spec.Template.Spec.Containers[i]
		if c.Name == containerName {
			index = i
			break
		}
	}
	if index < 0 {
		return nil, errLambdaNotFound
	}

	return &svc.Spec.Template.Spec.Containers[index].Env, nil
}

func shouldUpdateServing(
	log logr.Logger,
	svc *servingv1.Service,
	envs []corev1.EnvVar,
	imgTag string) (bool, error) {
	hasSameImgTag, err := hasSameImageTag(imgTag, svc.Labels)
	if err != nil {
		return false, err
	}

	if !hasSameImgTag {
		log.WithValues("expectedImageTag", imgTag).Info("image tags do not match")
		return true, nil
	}

	cenvs, err := lambdaEvs(svc)
	if err != nil {
		return false, err
	}

	equal := equal(envs, *cenvs)

	if !equal {
		log.WithValues("envs", toMap(envs), "cenvs", toMap(*cenvs)).Info("environmental variables do not match")
	}

	return !equal, nil
}

func hasSameImageTag(imgTag string, l map[string]string) (bool, error) {
	value, found := l["imageTag"]
	if !found {
		return false, errImageTagNotFound
	}

	return value == imgTag, nil
}

func equal(l, r []corev1.EnvVar) bool {
	lm := toMap(l)
	rm := toMap(r)
	return reflect.DeepEqual(lm, rm)
}

func toMap(vars []corev1.EnvVar) map[string]string {
	result := make(map[string]string, len(vars))
	for _, v := range vars {
		result[v.Name] = v.Value
	}
	return result
}
