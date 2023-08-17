package tests

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/poller"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/gitops"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/teststep"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func GitopsSteps(restConfig *rest.Config, cfg testsuite.Config, logf *logrus.Entry) (step.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-serverless-gitops", now.Hour(), now.Minute(), now.Second())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating dynamic client")
	}
	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating k8s core client")
	}
	appsCli, err := typedappsv1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating k8s apps client")
	}

	genericContainer := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	gitFnName := "gitfunc"
	gitCfg, err := gitops.NewGitopsConfig(gitFnName, cfg.GitServerImage, cfg.GitServerRepoName, genericContainer)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating Git config")
	}

	gitFn := function.NewFunction(gitFnName, cfg.KubectlProxyEnabled, genericContainer)
	logf.Infof("Testing Git Function in namespace: %s", cfg.Namespace)

	poll := poller.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		Log:                logf,
		DataKey:            testsuite.TestDataKey,
	}
	return step.NewSerialTestRunner(logf, "create git func",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		teststep.NewGitServer(gitCfg, "Start in-cluster Git Server", appsCli.Deployments(genericContainer.Namespace), coreCli.Services(genericContainer.Namespace), cfg.KubectlProxyEnabled, cfg.IstioEnabled),
		teststep.CreateFunction(logf, gitFn, "Create Git Function", gitops.GitopsFunction(gitCfg.GetGitServerInClusterURL(), "/", "master", serverlessv1alpha2.NodeJs18, nil)),
		teststep.NewDefaultedFunctionCheck("Check if Git Function has correct default values", gitFn),
		teststep.NewHTTPCheck(logf, "Git Function pre update simple check through service", gitFn.FunctionURL, poll, "GITOPS 1"),
		teststep.NewCommitChanges(logf, "Commit changes to Git Function", gitCfg.GetGitServerURL(cfg.KubectlProxyEnabled)),
		teststep.NewHTTPCheck(logf, "Git Function post update simple check through service", gitFn.FunctionURL, poll, "GITOPS 2")), nil
}
