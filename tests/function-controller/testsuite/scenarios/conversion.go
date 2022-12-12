package scenarios

import (
	"fmt"
	"time"

	functionv1alpha1 "github.com/kyma-project/kyma/tests/function-controller/pkg/function/v1alpha1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/gitrepository"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/poller"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	gitopsv1alpha1 "github.com/kyma-project/kyma/tests/function-controller/testsuite/gitops/v1alpha1"
	runtimesv1alpha1 "github.com/kyma-project/kyma/tests/function-controller/testsuite/runtimes/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/teststep"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func ConversionTest(restConfig *rest.Config, cfg testsuite.Config, logf *logrus.Entry) (step.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-serverless-conversion-v1alpha1", now.Hour(), now.Minute(), now.Second())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating dynamic client")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	appsCli, err := typedappsv1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating k8s apps client")
	}

	python39Logger := logf.WithField(scenarioKey, "python39-v1alpha1")
	gitFuncLogger := logf.WithField(scenarioKey, "git-function-v1alpha1")

	genericContainer := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	python39Fn := functionv1alpha1.NewFunction("python39", cfg.KubectlProxyEnabled, genericContainer)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating python39 config")
	}

	gitCfg, err := gitopsv1alpha1.NewGitopsConfig("gitfunc", cfg.GitServerImage, cfg.GitServerRepoName, cfg.KubectlProxyEnabled, genericContainer)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating Git config")
	}

	poll := poller.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            testsuite.TestDataKey,
	}

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)
	return step.NewSerialTestRunner(logf, "Serverless conversion tests",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		step.NewParallelRunner(logf, "",
			step.NewSerialTestRunner(gitFuncLogger, "Convert Git function",
				teststep.NewGitServerV1Alpha1(gitCfg, "Start in-cluster Git Server", appsCli.Deployments(genericContainer.Namespace), coreCli.Services(genericContainer.Namespace), cfg.KubectlProxyEnabled, cfg.IstioEnabled),
				teststep.NewCreateGitRepository(gitFuncLogger, gitrepository.New("git-repo", genericContainer), "Create Git Repository", gitopsv1alpha1.NoAuthRepositorySpec(gitCfg.GitInClusterURL.String())),
				teststep.CreateFunctionV1Alpha1(gitFuncLogger, gitCfg.Fn, "Create Git Function", gitopsv1alpha1.GitopsFunction("git-repo", "/", "master", serverlessv1alpha1.Nodejs14)),
				teststep.NewHTTPCheck(gitFuncLogger, "gitops function check ", gitCfg.InClusterURL, poll, "GITOPS 1")),
			step.NewSerialTestRunner(logf, "Convert Inline Python39 function",
				teststep.CreateFunctionV1Alpha1(python39Logger, python39Fn, "Create Python39 Function in version v1alpha1", runtimesv1alpha1.BasicPythonFunction("Hello From python", serverlessv1alpha1.Python39)),
				teststep.NewHTTPCheck(python39Logger, "Python39 v1alpha1 simple check through service", python39Fn.FunctionURL, poll, "Hello From python"),
			)),
	), nil
}
