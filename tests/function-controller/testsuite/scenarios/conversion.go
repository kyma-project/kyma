package scenarios

import (
	"fmt"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/poller"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	runtimesv1alpha1 "github.com/kyma-project/kyma/tests/function-controller/testsuite/runtimes/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/teststep"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"time"
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

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := poller.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            testsuite.TestDataKey,
	}

	return step.NewSerialTestRunner(logf, "Python39 conversion test",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		teststep.CreateFunctionV1Alpha1(python39Logger, python39Cfg.Fn, "Create Python39 Function in version v1alpha1", runtimesv1alpha1.BasicPythonFunction("Hello From python", serverlessv1alpha1.Python39)),
		teststep.NewHTTPCheck(python39Logger, "Python39 v1alpha1 simple check through service", python39Cfg.InClusterURL, poll.WithLogger(python39Logger), "Hello From python"),
	), nil
}
