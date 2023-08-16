package tests

import (
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/runtimes"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/teststep"
)

const (
	nodejs16 = "nodejs16"
	nodejs18 = "nodejs18"
	python39 = "python39"
)

func SimpleFunctionAPIGatewayTest(restConfig *rest.Config, cfg testsuite.Config, logf *logrus.Entry) (step.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-simple-api-gateway", now.Hour(), now.Minute(), now.Second())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating dynamic client")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	python39Logger := logf.WithField(runtimeKey, "python39")
	nodejs16Logger := logf.WithField(runtimeKey, "nodejs16")
	nodejs18Logger := logf.WithField(runtimeKey, "nodejs18")

	genericContainer := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	python39Fn := function.NewFunction("python39", cfg.KubectlProxyEnabled, genericContainer.WithLogger(python39Logger))

	nodejs16Fn := function.NewFunction("nodejs16", cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs16Logger))

	nodejs18Fn := function.NewFunction("nodejs18", cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs18Logger))

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	return step.NewSerialTestRunner(logf, "Runtime test",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		step.NewParallelRunner(logf, "Fn tests",
			step.NewSerialTestRunner(python39Logger, "Python39 test",
				teststep.CreateFunction(python39Logger, python39Fn, "Create Python39 Function", runtimes.BasicPythonFunction("Hello from python39", serverlessv1alpha2.Python39)),
				teststep.NewAPIGatewayFunctionCheck("python39", python39Fn, coreCli, genericContainer.Namespace, python39),
			),
			step.NewSerialTestRunner(nodejs16Logger, "NodeJS16 test",
				teststep.CreateFunction(nodejs16Logger, nodejs16Fn, "Create NodeJS16 Function", runtimes.BasicNodeJSFunction("Hello from nodejs16", serverlessv1alpha2.NodeJs16)),
				teststep.NewAPIGatewayFunctionCheck("nodejs16", nodejs16Fn, coreCli, genericContainer.Namespace, nodejs16),
			),
			step.NewSerialTestRunner(nodejs18Logger, "NodeJS18 test",
				teststep.CreateFunction(nodejs18Logger, nodejs18Fn, "Create NodeJS18 Function", runtimes.BasicNodeJSFunction("Hello from nodejs18", serverlessv1alpha2.NodeJs18)),
				teststep.NewAPIGatewayFunctionCheck("nodejs18", nodejs18Fn, coreCli, genericContainer.Namespace, nodejs18),
			),
		),
	), nil
}
