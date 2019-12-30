package istioinjection

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func disableInjectionForNamespace(disable bool) {
	ns, _ := k8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if disable {
		ns.ObjectMeta.Labels = map[string]string{
			"istio-injection": "disabled",
		}
	} else {
		ns.ObjectMeta.Labels = map[string]string{}
	}
	k8sClient.CoreV1().Namespaces().Update(ns)
}

func generateTestID(n int) string {
	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func createDeployment(name string, disableSidecarInjection bool) (*appv1.Deployment, error) {
	log.Infof("Creating deployment '%s", name)
	labels := labels(name)

	annotations := make(map[string]string)
	if disableSidecarInjection {
		annotations["sidecar.istio.io/inject"] = "false"
	}

	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            fmt.Sprintf("cont-%s", name),
							Image:           "eu.gcr.io/kyma-project/example/http-db-service:0.0.6",
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8017,
								},
							},
						},
					},
				},
			},
		},
	}
	return k8sClient.AppsV1().Deployments(namespace).Create(deployment)
}

func deleteDeployment(name string) {
	log.Infof("Deleting deployment '%s", name)
	err := k8sClient.AppsV1().Deployments(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		log.Errorf("Cannot delete deployment '%s': %v", name, err)
	}
}

func labels(testID string) map[string]string {
	labels := make(map[string]string)
	labels["createdBy"] = "api-controller-acceptance-tests"
	labels["app"] = fmt.Sprintf(testID)
	labels["test"] = "true"
	return labels
}

func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}