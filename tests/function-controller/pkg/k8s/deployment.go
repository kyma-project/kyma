package k8s

import (
	"context"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
)

const (
	labelKey = "component"
)

type Deployment struct {
	name      string
	namespace string
	image     string
	port      int32
	appsCli   appsclient.DeploymentInterface
	log       *logrus.Entry
}

func NewDeployment(name, namespace, image string, port int32, apps appsclient.DeploymentInterface, log *logrus.Entry) Deployment {
	return Deployment{
		name:      name,
		namespace: namespace,
		image:     image,
		port:      port,
		appsCli:   apps,
		log:       log,
	}
}

func (d Deployment) Create() error {
	rs := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.name,
			Labels: map[string]string{
				labelKey: d.name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &rs,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					labelKey: d.name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						labelKey: d.name,
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  d.name,
							Image: d.image,
							Ports: []v1.ContainerPort{
								{ContainerPort: d.port},
							},
							ImagePullPolicy: v1.PullAlways,
						},
					},
				},
			},
		},
	}
	_, err := d.appsCli.Create(context.Background(), deployment, metav1.CreateOptions{})
	return errors.Wrapf(err, "while creating Deployment %s in namespace %s", d.name, d.namespace)
}

func (d Deployment) Delete(ctx context.Context, options metav1.DeleteOptions) error {
	return d.appsCli.Delete(ctx, d.name, options)
}

func (d Deployment) Get(ctx context.Context, options metav1.GetOptions) (*appsv1.Deployment, error) {
	deployment, err := d.appsCli.Get(ctx, d.name, options)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting deployment %s in namespace %s", d.name, d.namespace)
	}
	return deployment, nil
}
func (d Deployment) LogResource() error {
	deployment, err := d.Get(context.TODO(), metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while getting deployment")
	}
	out, err := helpers.PrettyMarshall(deployment)
	if err != nil {
		return errors.Wrap(err, "while marshalling deployment")
	}
	d.log.Info(out)
	return nil
}
