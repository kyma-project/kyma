package testsuite

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/internal"
	"github.com/kyma-project/kyma/tests/function-controller/internal/assertion"
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/function"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/git"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/runtimes"
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"
	"time"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func GitopsSteps(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
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

	genericContainer := utils.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	gitFnName := "gitfunc"
	gitCfg, err := git.NewGitopsConfig(gitFnName, cfg.GitServerImage, cfg.GitServerRepoName, genericContainer)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating Git config")
	}

	gitFn := function.NewFunction(gitFnName, genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer)
	logf.Infof("Testing Git Function in namespace: %s", cfg.Namespace)

	poll := utils.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		Log:                logf,
		DataKey:            internal.TestDataKey,
	}
	return executor.NewSerialTestRunner(logf, "create git func",
		namespace.NewNamespaceStep(logf, "Create test namespace", genericContainer.Namespace, coreCli),
		git.NewGitServer(gitCfg, "Start in-cluster Git Server", appsCli.Deployments(genericContainer.Namespace), coreCli.Services(genericContainer.Namespace), cfg.KubectlProxyEnabled, cfg.IstioEnabled),
		function.CreateFunction(logf, gitFn, "Create Git Function", runtimes.GitopsFunction(gitCfg.GetGitServerInClusterURL(), "/", "master", serverlessv1alpha2.NodeJs18, nil)),
		assertion.NewHTTPCheck(logf, "Git Function pre update simple check through service", gitFn.FunctionURL, poll, "GITOPS 1"),
		git.NewCommitChanges(logf, "Commit changes to Git Function", gitCfg.GetGitServerURL(cfg.KubectlProxyEnabled)),
		assertion.NewHTTPCheck(logf, "Git Function post update simple check through service", gitFn.FunctionURL, poll, "GITOPS 2")), nil
}
