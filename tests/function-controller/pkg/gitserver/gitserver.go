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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

const (
	labelKey = "component"
)

type GitServer struct {
	deployment   k8s.Deployment
	services     k8s.Service
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
		services:   k8s.NewService(name, c.Namespace, port, services, c.Log),
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

	err = gs.services.Create()
	if err != nil {
		return err
	}

	if gs.istioEnabled {
		return gs.createDestinationRule()
	}
	return nil
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

	errService := gs.services.Delete(context.Background(), metav1.DeleteOptions{})
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

	err := gs.services.LogResource()
	if err != nil {
		return errors.Wrap(err, "while logging service status")
	}

	err = gs.deployment.LogResource()
	if err != nil {
		return errors.Wrap(err, "while logging deployment status")
	}

	return nil
}
