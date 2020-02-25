package helloworld

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
)

type statefulSetTest struct {
	statefulSetName, output string
	coreClient              *kubernetes.Clientset
	log                     logrus.FieldLogger
}

func NewStatefulSetTest() (*statefulSetTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return nil, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &statefulSetTest{
		coreClient:      coreClient,
		statefulSetName: "hello",
		log:             logrus.WithField("test", "statefulset"),
	}, nil
}

func (sst *statefulSetTest) CreateResources(t *testing.T, namespace string) {
	replicas := int32(1)
	err := sst.createService(namespace)
	require.NoError(t, err)
	err = sst.createStatefulSet(namespace, replicas)
	require.NoError(t, err)
}

func (sst *statefulSetTest) TestResources(t *testing.T, namespace string) {
	err := sst.waitForPodDeployment(namespace, 2*time.Minute)
	require.NoError(t, err)
	host := fmt.Sprintf("http://%s.%s/date", sst.statefulSetName, namespace)
	sst.log.Infof("HTTP GET to: %s", host)
	value, err := sst.getOutput(host, 2*time.Minute)
	require.NoError(t, err)
	if sst.output == "" {
		sst.output = value
		require.NotEmpty(t, value)
	} else {
		require.Equal(t, value, sst.output)
	}
}

func (sst *statefulSetTest) getOutput(host string, waitmax time.Duration) (string, error) {
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

func (sst *statefulSetTest) createStatefulSet(namespace string, replicas int32) error {
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: sst.statefulSetName,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: sst.statefulSetName,
			Replicas:    int32Ptr(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "statefulSet",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "statefulSet",
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						corev1.Container{
							Name:  "busybox",
							Image: "busybox",
							Command: []string{
								"sh",
								"-c",
								"test -e /usr/share/nginx/html/date || date > /usr/share/nginx/html/date; cat /usr/share/nginx/html/date",
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "www",
									MountPath: "/usr/share/nginx/html",
								},
							},
						},
					},
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "nginx",
							Image: "nginx:alpine",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 80,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "www",
									MountPath: "/usr/share/nginx/html",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "www",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("5M"),
							},
						},
					},
				},
			},
		},
	}
	_, err := sst.coreClient.AppsV1().StatefulSets(namespace).Create(statefulSet)
	return err
}

func (sst *statefulSetTest) createService(namespace string) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: sst.statefulSetName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "statefulSet",
			},
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port:     int32(80),
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
	_, err := sst.coreClient.CoreV1().Services(namespace).Create(service)
	return err
}

func (sst *statefulSetTest) waitForPodDeployment(namespace string, waitmax time.Duration) error {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("statefulSet %v could not be created within given time  %v", sst.statefulSetName, waitmax)
		case <-tick:
			pods, err := sst.coreClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "app=statefulSet"})
			if err != nil {
				return err
			}
			if len(pods.Items) != 1 {
				break
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
			statefulSet, err := sst.coreClient.AppsV1().StatefulSets(namespace).Get(sst.statefulSetName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if statefulSet.Status.ReadyReplicas != int32(len(pods.Items)) {
				break
			}
			return nil
		}
	}
}
