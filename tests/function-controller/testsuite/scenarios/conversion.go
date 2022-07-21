package scenarios

import (
	"fmt"
	"time"

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
	currentDate := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%dh-%dm-%ds", "test-serverless-conversion-v1alpha1", currentDate.Hour(), currentDate.Minute(), currentDate.Second())

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

	genericContainer := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	python39Cfg, err := runtimesv1alpha1.NewFunctionSimpleConfig("python39", genericContainer.WithLogger(python39Logger))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating python39 config")
	}

	gitCfg, err := gitopsv1alpha1.NewGitopsConfig("gitfunc", cfg.GitServerImage, cfg.GitServerRepoName, genericContainer)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating Git config")
	}

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := poller.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            testsuite.TestDataKey,
	}

	return step.NewSerialTestRunner(logf, "Serverless conversion tests",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		step.NewParallelRunner(logf, "",
			step.NewSerialTestRunner(logf, "Convert Git function",
				teststep.NewGitServerV1Alpha1(gitCfg, "Start in-cluster Git Server", appsCli.Deployments(genericContainer.Namespace), coreCli.Services(genericContainer.Namespace), cfg.IstioEnabled),
				teststep.CreateFunctionV1Alpha1(genericContainer.Log, gitCfg.Fn, "Create Git Function", gitopsv1alpha1.GitopsFunction(gitCfg.GetGitServerInClusterURL(), "/", "master", serverlessv1alpha1.Nodejs14)),
				teststep.NewHTTPCheck(genericContainer.Log, "gitops function check ", gitCfg.InClusterURL, poll, "GITOPS 1")),
			step.NewSerialTestRunner(logf, "Convert Inline Python39 function",
				teststep.CreateFunctionV1Alpha1(python39Logger, python39Cfg.Fn, "Create Python39 Function in version v1alpha1", runtimesv1alpha1.BasicPythonFunction("Hello From python", serverlessv1alpha1.Python39)),
				teststep.NewHTTPCheck(python39Logger, "Python39 v1alpha1 simple check through service", python39Cfg.InClusterURL, poll.WithLogger(python39Logger), "Hello From python"),
			)),
	), nil
}
