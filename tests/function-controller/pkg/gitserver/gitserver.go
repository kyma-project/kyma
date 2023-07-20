package gitserver

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/k8s"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
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
	deployment   k8s.Deployment
	services     coreclient.ServiceInterface
	resCli       *resource.Resource
	istioEnabled bool
	name         string
	namespace    string
	image        string
	port         int32
	log          *logrus.Entry
}

func New(c shared.Container, name string, image string, port int32, deployments appsclient.DeploymentInterface, services coreclient.ServiceInterface, istioEnabled bool) *GitServer {
	return &GitServer{
		deployment: k8s.NewDeployment(name, c.Namespace, image, port, deployments, c.Log),
		services:   services,
		resCli: resource.New(c.DynamicCli, schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1alpha3",
			Resource: "destinationrules"}, c.Namespace, c.Log, c.Verbose),
		name:         name,
		image:        image,
		port:         port,
		namespace:    c.Namespace,
		log:          c.Log,
		istioEnabled: istioEnabled,
	}
}

func (gs *GitServer) Create() error {
	err := gs.deployment.Create()
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
					Port:       gs.port,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(int(gs.port)),
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
	errDeployment := gs.deployment.Delete(context.Background(), metav1.DeleteOptions{})
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
	svc.Kind = "service"
	out, err := helpers.PrettyMarshall(svc)
	if err != nil {
		return errors.Wrap(err, "while marshalling service")
	}
	gs.log.Info(out)

	err = gs.deployment.LogResource()
	if err != nil {
		return errors.Wrap(err, "while logging deployment status")
	}

	return nil
}
