package scenarios

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"

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

func FunctionTestStep(restConfig *rest.Config, cfg testsuite.Config, logf *logrus.Entry) (step.Step, error) {
	currentDate := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%dh-%dm-%d", "test-serverless-full", currentDate.Hour(), currentDate.Minute(), rand.Int())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating dynamic client")
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s clientset")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	nodejs12Logger := logf.WithField(scenarioKey, "nodejs12")

	genericContainer := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}
	emptyFn := function.NewFunction("empty-fn", genericContainer.WithLogger(nodejs12Logger))

	nodejs12Cfg, err := runtimes.NewFunctionConfig("nodejs12", cfg.UsageKindName, cfg.DomainName, genericContainer.WithLogger(nodejs12Logger), clientset)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating nodejs12 config")
	}

	addon := addons.New("test-addon", genericContainer)
	addonURL, err := url.Parse(testsuite.AddonsConfigUrl)
	if err != nil {
		return nil, err
	}

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := poller.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            testsuite.TestDataKey,
	}
	return step.NewSerialTestRunner(logf, "runtime Test",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		teststep.NewAddonConfiguration("Create addon configuration", addon, addonURL, genericContainer),
		teststep.CreateEmptyFunction(emptyFn),
		teststep.CreateFunction(nodejs12Logger, nodejs12Cfg.Fn, "Create NodeJS12 Function", runtimes.BasicNodeJSFunction("Hello From nodejs", serverlessv1alpha1.Nodejs12)),
		teststep.NewDefaultedFunctionCheck("Check NodeJS12 function has correct default values", nodejs12Cfg.Fn),
		teststep.NewConfigureFunction(nodejs12Logger, "Check NodeJS12 function post-upgrade", nodejs12Cfg.FnName, nodejs12Cfg.ApiRule, nodejs12Cfg.APIRuleURL, nodejs12Cfg.SinkURL,
			nodejs12Cfg.Subscription, nodejs12Cfg.SvcInstance, nodejs12Cfg.SvcBinding, nodejs12Cfg.SvcBindingUsage, cfg.DomainPort),
		teststep.NewHTTPCheck(nodejs12Logger, "NodeJS12 pre update simple check through service", nodejs12Cfg.APIRuleURL, poll.WithLogger(nodejs12Logger), "Hello From nodejs"),
		teststep.NewHTTPCheck(nodejs12Logger, "NodeJS12 pre update simple check through gateway", nodejs12Cfg.InClusterURL, poll.WithLogger(nodejs12Logger), "Hello From nodejs"),
		teststep.UpdateFunction(nodejs12Logger, nodejs12Cfg.Fn, "Update NodeJS12 Function", runtimes.GetUpdatedNodeJSFunction()),
		teststep.NewE2EFunctionCheck(nodejs12Logger, "NodeJS12 post update e2e check", cfg.PublishURL, nodejs12Cfg.InClusterURL, nodejs12Cfg.APIRuleURL, poll.WithLogger(nodejs12Logger)),
	), nil
}
