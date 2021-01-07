package gitserver

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

const (
	labelKey = "component"
)

type GitServer struct {
	deployments  appsclient.DeploymentInterface
	services     coreclient.ServiceInterface
	resCli       *resource.Resource
	istioEnabled bool
	name         string
	namespace    string
	image        string
	port         int
	waitTimeout  time.Duration
	log          *logrus.Entry
	verbose      bool
}

func New(c shared.Container, name string, image string, port int, deployments appsclient.DeploymentInterface, services coreclient.ServiceInterface, istioEnabled bool) *GitServer {
	return &GitServer{
		deployments: deployments,
		services:    services,
		resCli: resource.New(c.DynamicCli, schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules"}, c.Namespace, c.Log, c.Verbose),
		name:         name,
		image:        image,
		port:         port,
		namespace:    c.Namespace,
		waitTimeout:  c.WaitTimeout,
		log:          c.Log,
		verbose:      c.Verbose,
		istioEnabled: istioEnabled,
	}
}

func (gs *GitServer) Create() error {
	err := gs.createDeployment()
	if err != nil {
		return err
	}

	err = gs.createService()
	if err != nil {
		return err
	}

	if gs.istioEnabled {
		return gs.createDestinationRule()
	}
	return nil
}

func (gs *GitServer) createDeployment() error {
	rs := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: gs.name,
			Labels: map[string]string{
				labelKey: gs.name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &rs,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					labelKey: gs.name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						labelKey: gs.name,
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  gs.name,
							Image: gs.image,
							Ports: []v1.ContainerPort{
								{ContainerPort: int32(gs.port)},
							},
							ImagePullPolicy: v1.PullAlways,
						},
					},
				},
			},
		},
	}
	_, err := gs.deployments.Create(context.Background(), deployment, metav1.CreateOptions{})
	return errors.Wrapf(err, "while creating Deployment %s in namespace %s", gs.name, gs.namespace)
}

func (gs *GitServer) createService() error {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: gs.name,
		},
		Spec: v1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []v1.ServicePort{
				{
					Name:       gs.name,
					Port:       int32(gs.port),
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(gs.port),
				},
			},
			Selector: map[string]string{
				labelKey: gs.name,
			},
		},
	}

	_, err := gs.services.Create(context.Background(), service, metav1.CreateOptions{})
	return errors.Wrapf(err, "while creating Service %s in namespace %s", gs.name, gs.namespace)
}

func (gs *GitServer) createDestinationRule() error {
	destRule := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1alpha3",
			"kind":       "DestinationRule",
			"metadata": map[string]interface{}{
				"name":      gs.name,
				"namespace": gs.namespace,
			},
			"spec": map[string]interface{}{
				"host": fmt.Sprintf("%s.%s.svc.cluster.local", gs.name, gs.namespace),
				"trafficPolicy": map[string]interface{}{
					"tls": map[string]interface{}{
						"mode": "DISABLE",
					},
				},
			},
		},
	}

	_, err := gs.resCli.Create(&destRule)
	return errors.Wrapf(err, "while creating DestinationRule %s in namespace %s", gs.name, gs.namespace)
}

func (gs *GitServer) Delete() error {
	var errDestRule error = nil
	if gs.istioEnabled {
		errDestRule = gs.resCli.Delete(gs.name)
	}

	errService := gs.services.Delete(context.Background(), gs.name, metav1.DeleteOptions{})
	errDeployment := gs.deployments.Delete(context.Background(), gs.name, metav1.DeleteOptions{})
	err := multierror.Append(errDeployment, errService, errDestRule)
	return err.ErrorOrNil()
}

func (gs *GitServer) LogResource() error {
	if gs.istioEnabled {
		obj, err := gs.resCli.Get(gs.name)
		if err != nil {
			return errors.Wrap(err, "while getting destination rule")
		}
		out, err := helpers.PrettyMarshall(obj)
		if err != nil {
			return errors.Wrap(err, "while marshalling destination rule")
		}
		gs.log.Info(out)
	}

	svc, err := gs.services.Get(context.Background(), gs.name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while getting service")
	}
	// The client doesn't fill service TypeMeta
	svc.TypeMeta.Kind = "service"
	out, err := helpers.PrettyMarshall(svc)
	if err != nil {
		return errors.Wrap(err, "while marshalling service")
	}
	gs.log.Info(out)

	deployment, err := gs.deployments.Get(context.Background(), gs.name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while getting deployment")
	}
	// The client doesn't fill deployment TypeMeta
	deployment.TypeMeta.Kind = "deployment"
	out, err = helpers.PrettyMarshall(deployment)
	if err != nil {
		return errors.Wrap(err, "while marshalling deployment")
	}
	gs.log.Info(out)

	return nil
}
