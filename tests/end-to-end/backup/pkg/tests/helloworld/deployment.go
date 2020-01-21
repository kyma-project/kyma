package helloworld

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/kubernetes"
)

type deploymentTest struct {
	deploymentName string
	coreClient     *kubernetes.Clientset
}

func NewDeploymentTest() (deploymentTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return deploymentTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return deploymentTest{}, err
	}

	return deploymentTest{
		coreClient:     coreClient,
		deploymentName: "hello",
	}, nil
}

func (d deploymentTest) CreateResources(namespace string) {
	replicas := int32(2)
	err := d.createDeployment(namespace, replicas)
	So(err, ShouldBeNil)
	err = d.createService(namespace)
	So(err, ShouldBeNil)
}

func (d deploymentTest) TestResources(namespace string) {
	replicas := int32(2)
	err := d.waitForPodDeployment(namespace, replicas, 2*time.Minute)
	So(err, ShouldBeNil)
	host := fmt.Sprintf("http://%s.%s:8080", d.deploymentName, namespace)
	value, err := d.getOutput(host, 2*time.Minute)
	So(err, ShouldBeNil)
	So(value, ShouldContainSubstring, "Welcome to nginx!")
}

func (d deploymentTest) getOutput(host string, waitmax time.Duration) (string, error) {
	tick := time.Tick(2 * time.Second)
	timeout := time.After(waitmax)
	messages := ""

	for {
		select {
		case <-tick:
			resp, err := http.Get(host)
			if err != nil {
				messages += fmt.Sprintf("%+v\n", err)
				break
			}
			if resp.StatusCode == http.StatusOK {
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "", err
				}
				return string(bodyBytes), nil
			}
			messages += fmt.Sprintf("%+v", err)

		case <-timeout:
			return "", fmt.Errorf("Could not get output:\n %v", messages)
		}
	}

}

func (d deploymentTest) createDeployment(namespace string, replicas int32) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "deployment",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "deployment",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "nginx",
							Image: "nginx:alpine",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := d.coreClient.AppsV1().Deployments(namespace).Create(deployment)
	return err
}

func (d deploymentTest) createService(namespace string) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.deploymentName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "deployment",
			},
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					TargetPort: intstr.FromInt(80),
					Port:       int32(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	_, err := d.coreClient.CoreV1().Services(namespace).Create(service)
	return err
}

func (d deploymentTest) waitForPodDeployment(namespace string, replicas int32, waitmax time.Duration) error {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Deployment %v could not be created within given time  %v", d.deploymentName, waitmax)
		case <-tick:
			pods, err := d.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=deployment"})
			if err != nil {
				return err
			}
			if len(pods.Items) < int(replicas) {
				log.Printf("%+v", pods.Items)
				break
			}
			if len(pods.Items) > int(replicas) {
				return fmt.Errorf("Deployed %v pod, got %v: %+v", replicas, len(pods.Items), pods)
			}

			stillStarting := false
			errorMessage := ""
			for _, pod := range pods.Items {
				if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodUnknown {
					errorMessage += fmt.Sprintf("Pod in state %v: \n%+v\n", pod.Status.Phase, pod)
				}
				if pod.Status.Phase == corev1.PodPending {
					stillStarting = true
				}
			}
			if errorMessage != "" {
				return fmt.Errorf(errorMessage)
			}
			if stillStarting {
				break
			}
			deployment, err := d.coreClient.AppsV1().Deployments(namespace).Get(d.deploymentName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if deployment.Status.ReadyReplicas != replicas {
				break
			}
			return nil
		}
	}
}

func int32Ptr(i int32) *int32 { return &i }
