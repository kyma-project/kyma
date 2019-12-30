package istioinjection

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const namespaceNameRoot = "istio-injection-tests"

// 1. Sidecar is not injected if there is a DISABLING LABEL ON NAMESPACE
func TestIstioInjectNamespaceFalseInject(t *testing.T) {
	disableInjectionForNamespace(true)
	defer disableInjectionForNamespace(false)

	testID := generateTestID(8)
	createDeployment(testID, true)
	defer deleteDeployment(testID)

	time.Sleep(2 * time.Second)

	pods, _ := k8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=" + testID})
	pod := pods.Items[0]
	size := len(pod.Spec.Containers)

	if size != 1 {
		t.Errorf("size must be 1")
	}
}

// 2. Sidecar is not injected if there is NO LABEL ON NAMESPACE but there is a DISABLING LABEL ON DEPLOYMENT
func TestIstioInjectDeploymentFalseInject(t *testing.T) {
	disableInjectionForNamespace(false)
	testID := generateTestID(5)
	createDeployment(testID, false)
	defer deleteDeployment(testID)

	time.Sleep(2 * time.Second)

	pods, _ := k8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=" + testID})
	pod := pods.Items[0]
	size := len(pod.Spec.Containers)

	if size != 1 {
		t.Errorf("size must be 1")
	}
}

// 3. Sidecar is injected
func TestIstioInject(t *testing.T) {
	disableInjectionForNamespace(false)
	testID := generateTestID(5)
	createDeployment(testID, true)
	defer deleteDeployment(testID)

	time.Sleep(2 * time.Second)

	pods, _ := k8sClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=" + testID})
	pod := pods.Items[0]
	size := len(pod.Spec.Containers)

	if size != 2 {
		t.Errorf("size must be 2")
	}
}

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

func createDeployment(name string, injectSidecar bool) (*appv1.Deployment, error) {
	log.Infof("Creating deployment '%s", name)
	labels := labels(name)

	annotations := make(map[string]string)
	if !injectSidecar {
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
