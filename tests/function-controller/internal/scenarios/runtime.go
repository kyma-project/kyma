package scenarios

import (
	"fmt"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/internal/teststep"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/addons"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/broker"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/job"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebinding"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebindingusage"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/serviceinstance"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/trigger"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"net/url"
	"testing"

	//"time"
)

type Scenario struct {
}

type Config struct {
	Log        shared.Logger
	CoreCli    typedcorev1.CoreV1Interface
	DynamicCLI dynamic.Interface
}

func Steps(config *Config) []step.Step {
	namespace := "function-test"

	pythonFunc := FunctionConfig{
		fn:      function.NewFunction2("python-37-test", namespace, config.DynamicCLI, config.Log),
		apiRule: apirule.New2("python37", namespace, config.Log, config.DynamicCLI),
		name:    "python-37-test",
	}

	nodejs12Func := FunctionConfig{
		fn:      function.NewFunction2("nodejs12-test", namespace, config.DynamicCLI, config.Log),
		apiRule: apirule.New2("nodejs12", namespace, config.Log, config.DynamicCLI),
		name:    "nodejs12-test",
	}
	addon := addons.New2(config.Log, "test-addon", namespace, config.DynamicCLI)
	addonURL, err := url.Parse("https://github.com/kyma-project/addons/releases/download/0.13.0/index-testing.yaml")
	if err != nil {
		panic(err)
	}

	domainName := "lucky-cancer.wookiee.hudy.ninja"
	expectedPythonURL := fmt.Sprintf("http://%s.%s", pythonFunc.name, domainName)
	expectedNodeJSURL := fmt.Sprintf("http://%s.%s", nodejs12Func.name, domainName)
	return []step.Step{
		//teststep.NewNamespaceStep(config.Log, config.CoreCli, namespace),
		teststep.NewAddonConfiguration("Create Addon configuration", addon, addonURL),
		step.Parallel(
			teststep.NewEmptyFunction(function.NewFunction2("empty function", namespace, config.DynamicCLI, config.Log)),
			teststep.NewSerialSteps(config.Log, "Python 3.7 Function tests",
				teststep.CreateFunction(config.Log, pythonFunc.fn, pythonFunc.name, pythonFunctionBody("Hello From Python")),
				teststep.NewDefaultedFunctionCheck(pythonFunc.fn),
				teststep.NewAPIRule(pythonFunc.apiRule, "python api rule", pythonFunc.name, domainName, 80),
				teststep.NewCheck(config.Log, "python function check", expectedPythonURL, "Hello From Python"),
			),
			teststep.NewSerialSteps(config.Log, "Node JS 12 Function tests",
				teststep.CreateFunction(config.Log, nodejs12Func.fn, nodejs12Func.name, nodejsFunctionBody("Hello From Nodejs12")),
				teststep.NewAPIRule(nodejs12Func.apiRule, "nodejs api rule", nodejs12Func.name, domainName, 80),
				teststep.NewCheck(config.Log, "NodeJS function check", expectedNodeJSURL, "Hello From Nodejs12"),
			),
		),

		//teststep.NewPause(10 * time.Minute),
	}
}

func NewBigStep(restConfig *rest.Config, cfg testsuite.Config, t *testing.T, g *gomega.GomegaWithT) []step.Step {
	cfg.NamespaceBaseName = "old-big-test"
	cfg.NamespaceBaseName = "test-parallel"
	cfg.IngressHost = "lucky-cancer.wookiee.hudy.ninja"
	nodejs12Cfg := modifyConfig(cfg, "nodejs")
	python37Cfg := modifyConfig(cfg, "python")

	nodejs, err := getTestDef(restConfig, nodejs12Cfg, t, g)
	python, err := getTestDef(restConfig, python37Cfg, t, g)
	g.Expect(err).Should(gomega.BeNil())

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	return []step.Step{
		teststep.NewNamespaceStep(python.T, coreCli, cfg.NamespaceBaseName),
		step.Parallel(teststep.NewFunctionTest(nodejs, "nodejs test"),
			teststep.NewFunctionTest(python, "python test")),
	}
}

func modifyConfig(cfg testsuite.Config, runtime string) testsuite.Config {
	newCfg := cfg

	newCfg.FunctionName = runtime
	newCfg.APIRuleName = fmt.Sprintf("%s-rule", runtime)
	newCfg.TriggerName = fmt.Sprintf("%s-trigger", runtime)
	newCfg.AddonName = fmt.Sprintf("%s-addon", runtime)
	newCfg.ServiceInstanceName = fmt.Sprintf("%s-service-instance", runtime)
	newCfg.ServiceBindingName = fmt.Sprintf("%s-service-binding", runtime)
	newCfg.ServiceBindingUsageName = fmt.Sprintf("%s-service-binding-usage", runtime)
	newCfg.DomainName = fmt.Sprintf("%s-function", runtime)

	return newCfg
}

func getTestDef(restConfig *rest.Config, cfg testsuite.Config, t *testing.T, g *gomega.GomegaWithT) (*testsuite.TestSuite, error) {
	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Core client")
	}

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s clientset")
	}

	//namespaceName := fmt.Sprintf("%s-%d", cfg.NamespaceBaseName, rand.Uint32())
	namespaceName := cfg.NamespaceBaseName

	container := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   namespaceName,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         t,
	}

	ns := namespace.New(namespaceName, coreCli, container)
	f := function.NewFunction(cfg.FunctionName, container)
	ar := apirule.New(cfg.APIRuleName, container)
	br := broker.New(container)
	tr := trigger.New(cfg.TriggerName, container)
	ac := addons.New(cfg.AddonName, container)
	si := serviceinstance.New(cfg.ServiceInstanceName, container)
	sb := servicebinding.New(cfg.ServiceBindingName, container)
	sbu := servicebindingusage.New(cfg.ServiceBindingUsageName, cfg.UsageKindName, container)
	jobList := job.New(cfg.FunctionName, clientset.BatchV1(), container)

	return &testsuite.TestSuite{
		Namespace:           ns,
		Function:            f,
		ApiRule:             ar,
		Broker:              br,
		Trigger:             tr,
		AddonsConfig:        ac,
		Serviceinstance:     si,
		Servicebinding:      sb,
		Servicebindingusage: sbu,
		Jobs:                jobList,
		T:                   t,
		G:                   g,
		DynamicCli:          dynamicCli,
		Cfg:                 cfg,
	}, nil
}

type FunctionConfig struct {
	fn          *function.Function
	apiRule     *apirule.APIRule
	name        string
	expectedMsg string
}

func pythonFunctionBody(msg string) function.FunctionData {
	return function.FunctionData{
		Body: fmt.Sprintf(
			`def main(event, context):
	return "{}".format('%s')`, msg),
		Deps:    "",
		Runtime: serverlessv1alpha1.Python37,
	}
}

func nodejsFunctionBody(msg string) function.FunctionData {
	return function.FunctionData{
		Body:        fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, msg),
		Deps:        `{ "name": "hellowithoutdeps", "version": "0.0.1", "dependencies": { } }`,
		MaxReplicas: 2,
		MinReplicas: 1,
		Runtime:     serverlessv1alpha1.Nodejs12,
	}
}

//type KymaClients struct {
//	AppOperatorClientset         *appoperatorclientset.Clientset
//	AppBrokerClientset           *appbrokerclientset.Clientset
//	CoreClientset                *k8s.Clientset
//	Pods                         coreclient.PodInterface
//	ServiceCatalogClientset      *servicecatalogclientset.Clientset
//	ServiceBindingUsageClientset *sbuclientset.Clientset
//	ApiRules                     dynamic.ResourceInterface
//	Function                     dynamic.ResourceInterface
//}
//
//var (
//	apiRuleRes = schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha1", Resource: "apirules"}
//	function   = schema.GroupVersionResource{Group: "serverless.kyma-project.io", Version: "v1alpha1", Resource: "functions"}
//)
//
//func InitKymaClients(config *rest.Config, testID string) KymaClients {
//	coreClientset := k8s.NewForConfigOrDie(config)
//	client := dynamic.NewForConfigOrDie(config)
//
//	return KymaClients{
//		AppOperatorClientset:         appoperatorclientset.NewForConfigOrDie(config),
//		AppBrokerClientset:           appbrokerclientset.NewForConfigOrDie(config),
//		CoreClientset:                coreClientset,
//		Pods:                         coreClientset.CoreV1().Pods(testID),
//		ServiceCatalogClientset:      servicecatalogclientset.NewForConfigOrDie(config),
//		ServiceBindingUsageClientset: sbuclientset.NewForConfigOrDie(config),
//		ApiRules:                     client.Resource(apiRuleRes).Namespace(testID),
//		Function:                     client.Resource(function).Namespace(testID),
//	}
//}
