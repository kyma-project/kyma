package teststep

import (
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/git"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/gitserver"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/gitops"
	gitopsv1alpha1 "github.com/kyma-project/kyma/tests/function-controller/testsuite/gitops/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	appsCli "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

type newGitServer struct {
	name      string
	gs        *gitserver.GitServer
	gitClient git.Client
	log       *logrus.Entry
}

var _ step.Step = newGitServer{}

func NewGitServer(cfg gitops.GitopsConfig, stepName string, deployments appsCli.DeploymentInterface, services coreclient.ServiceInterface, useProxy, istioEnabled bool) step.Step {
	repoURL, err := helpers.GetGitURL(cfg.GitServerServiceName, cfg.Toolbox.Namespace, cfg.GitServerRepoName, useProxy)
	if err != nil {
		panic(err)
	}
	return newGitServer{
		name:      stepName,
		gs:        gitserver.New(cfg.Toolbox, cfg.GitServerServiceName, cfg.GitServerImage, cfg.GitServerServicePort, deployments, services, istioEnabled),
		gitClient: git.New(repoURL.String()),
		log:       cfg.Toolbox.Log.WithField(step.LogStepKey, stepName),
	}
}

func NewGitServerV1Alpha1(cfg gitopsv1alpha1.GitopsConfig, stepName string, deployments appsCli.DeploymentInterface, services coreclient.ServiceInterface, useProxy, istioEnabled bool) step.Step {
	repoURL, err := helpers.GetGitURL(cfg.GitServerServiceName, cfg.Toolbox.Namespace, cfg.GitServerRepoName, useProxy)
	if err != nil {
		panic(err)
	}

	return newGitServer{
		name:      stepName,
		gs:        gitserver.New(cfg.Toolbox, cfg.GitServerServiceName, cfg.GitServerImage, cfg.GitServerServicePort, deployments, services, istioEnabled),
		gitClient: git.New(repoURL.String()),
		log:       cfg.Toolbox.Log.WithField(step.LogStepKey, stepName),
	}
}

func (r newGitServer) Name() string {
	return r.name
}

func (r newGitServer) Run() error {
	err := r.gs.Create()
	if err != nil {
		return errors.Wrap(err, "while creating in-cluster Git server")
	}

	err = retry.Do(r.gitClient.TryCloning)
	if err != nil {
		return errors.Wrap(err, "while waiting for in-cluster Git Server to be ready")
	}
	return nil
}

func (r newGitServer) Cleanup() error {
	return errors.Wrap(r.gs.Delete(), "while deleting in-cluster Git server")
}

func (r newGitServer) OnError() error {
	return r.gs.LogResource()
}
