package scenarios

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/app"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
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

func SimpleFunctionCloudEventsTest(restConfig *rest.Config, cfg testsuite.Config, logf *logrus.Entry) (step.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-simple-cloud-events", now.Hour(), now.Minute(), now.Second())

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

	//python39Logger := logf.WithField(scenarioKey, "python39")
	//nodejs16Logger := logf.WithField(scenarioKey, "nodejs16")
	nodejs18Logger := logf.WithField(scenarioKey, "nodejs18")

	genericContainer := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	//python39Fn := function.NewFunction("python39", cfg.KubectlProxyEnabled, genericContainer.WithLogger(python39Logger))
	//nodejs16Fn := function.NewFunction("nodejs16", cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs16Logger))
	nodejs18Fn := function.NewFunction("nodejs18", cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs18Logger))

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	return step.NewSerialTestRunner(logf, "Runtime test",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		app.NewApplication("Create HTTP basic application", HTTPAppName, HTTPAppImage, int32(80), appsCli.Deployments(genericContainer.Namespace), coreCli.Services(genericContainer.Namespace), genericContainer),
		step.NewParallelRunner(logf, "Fn tests",
			//step.NewSerialTestRunner(python39Logger, "Python39 test",
			//	teststep.CreateFunction(python39Logger, python39Fn, "Create Python39 Function", runtimes.BasicCloudEventPythonFunction(serverlessv1alpha2.Python39)),
			//	teststep.NewCloudEventCheck(cloudevents.EncodingStructured, python39Logger, "Python39 cloud event structured check", python39Fn.FunctionURL),
			//),
			//step.NewSerialTestRunner(nodejs16Logger, "NodeJS16 test",
			//	teststep.CreateFunction(nodejs16Logger, nodejs16Fn, "Create NodeJS16 Function", runtimes.NodeJSFunctionWithCloudEvent(serverlessv1alpha2.NodeJs18)),
			//	teststep.NewCloudEventCheck(cloudevents.EncodingStructured, nodejs16Logger, "Python39 cloud event structured check", python39Fn.FunctionURL),
			//),
			step.NewSerialTestRunner(nodejs18Logger, "NodeJS18 test",
				teststep.CreateFunction(nodejs18Logger, nodejs18Fn, "Create NodeJS18 Function", runtimes.NodeJSFunctionWithCloudEvent(serverlessv1alpha2.NodeJs18)),
				teststep.NewCloudEventCheck(nodejs18Logger, "NodeJS18 cloud event structured check", cloudevents.EncodingStructured, nodejs18Fn.FunctionURL),
				teststep.NewCloudEventCheck(nodejs18Logger, "NodeJS18 cloud event binary check", cloudevents.EncodingBinary, nodejs18Fn.FunctionURL),
			),
		),
	), nil
}
