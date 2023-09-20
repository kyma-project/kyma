package testsuite

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kyma-project/kyma/tests/function-controller/internal"
	"github.com/kyma-project/kyma/tests/function-controller/internal/assertion"
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/app"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/function"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/runtimes"
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"
	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"time"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

func FunctionCloudEventsTest(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-cloud-events", now.Hour(), now.Minute(), now.Second())

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

	python39Logger := logf.WithField(runtimeKey, "python39")
	nodejs16Logger := logf.WithField(runtimeKey, "nodejs16")
	nodejs18Logger := logf.WithField(runtimeKey, "nodejs18")

	genericContainer := utils.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	publisherProxyMock := function.NewFunction("eventing-publisher-proxy", "kyma-system", cfg.KubectlProxyEnabled, genericContainer.WithLogger(python39Logger))
	python39Fn := function.NewFunction("python39", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(python39Logger))
	nodejs16Fn := function.NewFunction("nodejs16", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs16Logger))
	nodejs18Fn := function.NewFunction("nodejs18", genericContainer.Namespace, cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs18Logger))

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	return executor.NewSerialTestRunner(logf, "Runtime test",
		namespace.NewNamespaceStep(logf, fmt.Sprintf("Create %s namespace", genericContainer.Namespace), genericContainer.Namespace, coreCli),
		function.CreateFunction(logf, publisherProxyMock, "Create publisher proxy mock", runtimes.PythonPublisherProxyMock()),
		app.NewApplication("Create HTTP basic application", HTTPAppName, HTTPAppImage, int32(80), appsCli.Deployments(genericContainer.Namespace), coreCli.Services(genericContainer.Namespace), genericContainer),
		executor.NewParallelRunner(logf, "Fn tests",
			executor.NewSerialTestRunner(python39Logger, "Python39 test",
				function.CreateFunction(python39Logger, python39Fn, "Create Python39 Function", runtimes.PythonCloudEvent(serverlessv1alpha2.Python39)),
				assertion.CloudEventReceiveCheck(python39Logger, "Python39 cloud event structured check", cloudevents.EncodingStructured, python39Fn.FunctionURL),
				assertion.CloudEventReceiveCheck(python39Logger, "Python39 cloud event binary check", cloudevents.EncodingBinary, python39Fn.FunctionURL),
				assertion.CloudEventSendCheck(python39Logger, "Python39 cloud event sent check", python39Fn.FunctionURL),
			),
			executor.NewSerialTestRunner(nodejs16Logger, "NodeJS16 test",
				function.CreateFunction(nodejs16Logger, nodejs16Fn, "Create NodeJS16 Function", runtimes.NodeJSFunctionWithCloudEvent(serverlessv1alpha2.NodeJs18)),
				assertion.CloudEventReceiveCheck(nodejs16Logger, "NodeJS16 cloud event structured check", cloudevents.EncodingStructured, nodejs16Fn.FunctionURL),
				assertion.CloudEventReceiveCheck(nodejs16Logger, "NodeJS16 cloud event binary check", cloudevents.EncodingBinary, nodejs16Fn.FunctionURL),
				assertion.CloudEventSendCheck(nodejs16Logger, "NodeJS16 cloud event sent check", nodejs16Fn.FunctionURL),
			),
			executor.NewSerialTestRunner(nodejs18Logger, "NodeJS18 test",
				function.CreateFunction(nodejs18Logger, nodejs18Fn, "Create NodeJS18 Function", runtimes.NodeJSFunctionWithCloudEvent(serverlessv1alpha2.NodeJs18)),
				assertion.CloudEventReceiveCheck(nodejs18Logger, "NodeJS18 cloud event structured check", cloudevents.EncodingStructured, nodejs18Fn.FunctionURL),
				assertion.CloudEventReceiveCheck(nodejs18Logger, "NodeJS18 cloud event binary check", cloudevents.EncodingBinary, nodejs18Fn.FunctionURL),
				assertion.CloudEventSendCheck(nodejs18Logger, "NodeJS18 cloud event sent check", nodejs18Fn.FunctionURL),
			),
		),
	), nil
}
