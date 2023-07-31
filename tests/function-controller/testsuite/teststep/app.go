package teststep

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/k8s"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsCli "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

var _ step.Step = &Application{}

/*
Application consist of deployment and service
*/
type Application struct {
	deployment k8s.Deployment
	svc        k8s.Service
	name       string
	stepName   string
	namespace  string
	log        *logrus.Entry
	port       int32
	image      string
}

func NewApplication(stepName, name string, image string, port int32, appCli appsCli.DeploymentInterface, coreCli coreclient.ServiceInterface, c shared.Container) step.Step {
	return &Application{
		deployment: k8s.NewDeployment(name, c.Namespace, image, port, appCli, c.Log),
		svc:        k8s.NewService(name, c.Namespace, port, coreCli, c.Log),
		name:       name,
		stepName:   stepName,
		namespace:  c.Namespace,
		log:        c.Log,
		port:       port,
		image:      image,
	}
}

func (a Application) Name() string {
	return a.name
}

func (a Application) Run() error {
	err := a.deployment.Create()
	if err != nil {
		return errors.Wrap(err, "while creating deployment for application")
	}
	err = a.svc.Create()
	if err != nil {
		return errors.Wrap(err, "while creating service for application ")
	}
	return nil
}

func (a Application) Cleanup() error {
	ctx := context.Background()
	deploymentErr := a.deployment.Delete(ctx, metav1.DeleteOptions{})
	svcErr := a.svc.Delete(ctx, metav1.DeleteOptions{})
	err := multierror.Append(deploymentErr, svcErr)
	return err.ErrorOrNil()
}

func (a Application) OnError() error {
	err := a.svc.LogResource()
	if err != nil {
		return errors.Wrap(err, "while logging application service status")
	}

	err = a.deployment.LogResource()
	if err != nil {
		return errors.Wrap(err, "while logging application deployment status")
	}

	return nil
}
