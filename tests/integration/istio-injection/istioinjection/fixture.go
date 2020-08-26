package istioinjection

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const HTTPDBSERVICE_IMAGE = "eu.gcr.io/kyma-project/example/http-db-service:0.0.6"

func (testSuite *TestSuite) createNamespace() error {
	log.Infof("Creating namespace '%s", testSuite.namespace)
	err := retry.Do(func() error {
		_, err := testSuite.k8sClient.CoreV1().Namespaces().Create(&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testSuite.namespace,
			},
			Spec: v1.NamespaceSpec{},
		})
		return err
	},
		retry.Attempts(5),
		retry.Delay(3*time.Second))
	if err != nil {
		log.Errorf("Cannot create namespace '%s': %v", testSuite.namespace, err)
	}
	return err
}

func (testSuite *TestSuite) deleteNamespace() error {
	log.Infof("Deleting namespace '%s", testSuite.namespace)
	var deleteImmediately int64
	err := retry.Do(func() error {
		err := testSuite.k8sClient.CoreV1().Namespaces().Delete(testSuite.namespace, &metav1.DeleteOptions{
			GracePeriodSeconds: &deleteImmediately,
		})
		return err
	},
		retry.Attempts(5),
		retry.Delay(3*time.Second))
	if err != nil {
		log.Errorf("I have run the tests, but haven't cleaned up '%s': %v", testSuite.namespace, err)
	}
	return err
}

func loadKubeConfigOrDie() *rest.Config {
	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			log.Errorf("Cannot create in-cluster config: %v", err)
			panic(err)
		}
		return cfg
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Errorf("Cannot read kubeconfig: %s", err)
		panic(err)
	}
	return cfg
}

func (testSuite *TestSuite) initK8sClient() {
	kubeConfig := loadKubeConfigOrDie()
	testSuite.k8sClient = kubernetes.NewForConfigOrDie(kubeConfig)
}

func (testSuite *TestSuite) disableInjectionForNamespace(disable bool) {
	var namespace *corev1.Namespace
	err := retry.Do(func() error {
		ns, er := testSuite.k8sClient.CoreV1().Namespaces().Get(testSuite.namespace, metav1.GetOptions{})
		namespace = ns
		return er
	},
		retry.Attempts(5),
		retry.Delay(3*time.Second))

	if err != nil {
		log.Errorf("Cannot get namespace '%s': %v", testSuite.namespace, err)
		panic(err)
	}

	if disable {
		namespace.ObjectMeta.Labels = map[string]string{
			"istio-injection": "disabled",
		}
	} else {
		namespace.ObjectMeta.Labels = map[string]string{}
	}

	err = retry.Do(func() error {
		_, err = testSuite.k8sClient.CoreV1().Namespaces().Update(namespace)
		return err
	},
		retry.Attempts(5),
		retry.Delay(3*time.Second))

	if err != nil {
		log.Errorf("Cannot update namespace '%s': %v", testSuite.namespace, err)
		panic(err)
	}
}

func (testSuite *TestSuite) createDeployment(name string, disableSidecarInjection bool) (*appv1.Deployment, error) {
	log.Infof("Creating deployment '%s", name)
	labels := testSuite.labels(name)

	annotations := make(map[string]string)
	if disableSidecarInjection {
		annotations["sidecar.istio.io/inject"] = "false"
	}

	deploymentManifest := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testSuite.namespace,
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
							Image:           HTTPDBSERVICE_IMAGE,
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

	var deployment *appv1.Deployment
	err := retry.Do(func() error {
		dep, err := testSuite.k8sClient.AppsV1().Deployments(testSuite.namespace).Create(deploymentManifest)
		deployment = dep
		return err
	},
		retry.Attempts(5),
		retry.Delay(3*time.Second))

	if err != nil {
		log.Errorf("Cannot create deployment '%s': %v", deploymentManifest, err)
	}

	return deployment, err
}

func (testSuite *TestSuite) getPods(appLabelValue string) (*v1.PodList, error) {
	var pods *v1.PodList
	var err error

	err = retry.Do(func() error {
		// check if Pod list is empty, if it is, retry
		pods, err = testSuite.k8sClient.CoreV1().Pods(testSuite.namespace).List(metav1.ListOptions{LabelSelector: "app=" + appLabelValue})
		if err != nil {
			return errors.Errorf("Cannot list the pods with label '%s': %v", appLabelValue, err)
		}
		if len(pods.Items) == 0 {
			return errors.Errorf("The pod is not created yet")
		}
		return nil
	},
		retry.Attempts(5),
		retry.Delay(3*time.Second))

	return pods, err
}

func (testSuite *TestSuite) deleteDeployment(name string) {
	log.Infof("Deleting deployment '%s", name)

	err := retry.Do(func() error {
		// Retry if deployment is not deleted
		err := testSuite.k8sClient.AppsV1().Deployments(testSuite.namespace).Delete(name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}

		_, err = testSuite.k8sClient.AppsV1().Deployments(testSuite.namespace).Get(name, metav1.GetOptions{})
		if err != nil && k8sErrors.IsNotFound(err) {
			return nil
		}

		return err
	},
		retry.Attempts(5),
		retry.Delay(3*time.Second))

	if err != nil {
		log.Fatal("Deployment could not be deleted: ", err)
	}
}

func (testSuite *TestSuite) labels(testID string) map[string]string {
	labels := make(map[string]string)
	labels["createdBy"] = "istio-injection-tests"
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
