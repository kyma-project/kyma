package scenarios

import (
	"fmt"
	"net/url"
	"time"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/sirupsen/logrus"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/addons"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/poller"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/runtimes"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/teststep"
)

const scenarioKey = "scenario"

func FunctionTestStep(restConfig *rest.Config, cfg testsuite.Config, logf *logrus.Logger) ([]step.Step, error) {
	currentDate := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%dh-%dm-%ds", "test-parallel", currentDate.Hour(), currentDate.Minute(), currentDate.Second())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return []step.Step{}, errors.Wrapf(err, "while creating dynamic client")
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s clientset")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	python38Logger := logf.WithField(scenarioKey, "python37")
	nodejs10Logger := logf.WithField(scenarioKey, "nodejs10")
	nodejs12Logger := logf.WithField(scenarioKey, "nodejs12")

	genericContainer := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logrus.NewEntry(logf),
	}

	nodejs12Cfg, err := runtimes.NewFunctionConfig("nodejs12", cfg.UsageKindName, cfg.DomainName, genericContainer.WithLogger(nodejs12Logger), clientset)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating nodejs12 config")
	}

	python38Cfg, err := runtimes.NewFunctionSimpleConfig("python37", genericContainer.WithLogger(python38Logger))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating python38 config")
	}

	nodejs10Cfg, err := runtimes.NewFunctionSimpleConfig("nodejs10", genericContainer.WithLogger(nodejs10Logger))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating nodejs10 config")
	}

	addon := addons.New("test-addon", genericContainer)

	addonURL, err := url.Parse(testsuite.AddonsConfigUrl)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := poller.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            testsuite.TestDataKey,
	}

	return []step.Step{
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		teststep.NewAddonConfiguration("Create addon configuration", addon, addonURL, genericContainer),
		step.Parallel(
			teststep.NewSerialSteps(python38Logger, "Python37 test",
				teststep.CreateFunction(python38Logger, python38Cfg.Fn, "Create Python37 Function", runtimes.BasicPythonFunction("Hello From python")),
				teststep.NewHTTPCheck(python38Logger, "Python37 pre update simple check through service", python38Cfg.InClusterURL, poll.WithLogger(python38Logger), "Hello From python"),
				teststep.UpdateFunction(python38Logger, python38Cfg.Fn, "Update Python37 Function", runtimes.BasicPythonFunction("Hello From updated python")),
				teststep.NewHTTPCheck(python38Logger, "Python37 post update simple check through service", python38Cfg.InClusterURL, poll.WithLogger(python38Logger), "Hello From updated python"),
			),
			teststep.NewSerialSteps(nodejs10Logger, "NodeJS10 test",
				teststep.CreateFunction(nodejs10Logger, nodejs10Cfg.Fn, "Create NodeJS10 Function", runtimes.BasicNodeJSFunction("Hello From nodejs10", serverlessv1alpha1.Nodejs10)),
				teststep.NewHTTPCheck(nodejs10Logger, "NodeJS10 pre update simple check through service", nodejs10Cfg.InClusterURL, poll.WithLogger(nodejs10Logger), "Hello From nodejs10"),
				teststep.UpdateFunction(nodejs10Logger, nodejs10Cfg.Fn, "Update NodeJS10 Function", runtimes.BasicNodeJSFunction("Hello From updated nodejs10", serverlessv1alpha1.Nodejs10)),
				teststep.NewHTTPCheck(nodejs10Logger, "NodeJS10 post update simple check through service", nodejs10Cfg.InClusterURL, poll.WithLogger(nodejs10Logger), "Hello From updated nodejs10"),
			),
			teststep.NewSerialSteps(nodejs12Logger, "NodeJS12 test",
				teststep.CreateEmptyFunction(nodejs12Cfg.Fn),
				teststep.CreateFunction(nodejs12Logger, nodejs12Cfg.Fn, "Create NodeJS12 Function", runtimes.BasicNodeJSFunction("Hello From nodejs", serverlessv1alpha1.Nodejs12)),
				teststep.NewDefaultedFunctionCheck("Check NodeJS12 function has correct default values", nodejs12Cfg.Fn),
				teststep.ConfigureFunction(nodejs12Logger, "Check NodeJS12 function post-upgrade", nodejs12Cfg.FnName, nodejs12Cfg.ApiRule, nodejs12Cfg.APIRuleURL,
					nodejs12Cfg.SvcInstance, nodejs12Cfg.SvcBinding, nodejs12Cfg.SvcBindingUsage,
					nodejs12Cfg.Broker, nodejs12Cfg.Trigger, cfg.DomainPort),
				teststep.NewHTTPCheck(nodejs12Logger, "NodeJS12 pre update simple check through service", nodejs12Cfg.APIRuleURL, poll.WithLogger(nodejs12Logger), "Hello From nodejs"),
				teststep.NewHTTPCheck(nodejs12Logger, "NodeJS12 pre update simple check through gateway", nodejs12Cfg.InClusterURL, poll.WithLogger(nodejs12Logger), "Hello From nodejs"),
				teststep.UpdateFunction(nodejs12Logger, nodejs12Cfg.Fn, "Update NodeJS12 Function", runtimes.GetUpdatedNodeJSFunction()),
				teststep.NewE2EFunctionCheck(nodejs12Logger, "NodeJS12 post update e2e check", nodejs12Cfg.InClusterURL, nodejs12Cfg.APIRuleURL, nodejs12Cfg.BrokerURL, poll.WithLogger(nodejs12Logger)),
			)),
	}, nil
}
