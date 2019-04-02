package helloworld

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsTypes "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsCli "k8s.io/client-go/kubernetes/typed/apps/v1"
)

const deployName = "hello-world-nginx"

// HelloWorld shows the example hello world backup test
type HelloWorld struct {
	deployCli appsCli.DeploymentsGetter
}

// NewTest returns new instance of the HellowWord test
func NewTest(deployCli appsCli.DeploymentsGetter) *HelloWorld {
	return &HelloWorld{
		deployCli: deployCli,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (h *HelloWorld) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	replicas := int32(1)
	labels := map[string]string{
		"app": "hello-world",
	}

	_, err := h.deployCli.Deployments(namespace).Create(&appsTypes.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
		Spec: appsTypes.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.7.9",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "while creating hello-world deployment")
	}

	return nil
}

// TestResources tests resources after backup phase
func (h *HelloWorld) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	_, err := h.deployCli.Deployments(namespace).Get(deployName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while checking if deplyment %q still exists", deployName)
	}

	return nil
}
