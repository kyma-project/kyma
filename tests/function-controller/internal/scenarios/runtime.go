package scenarios

import (
	"fmt"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/internal/teststep"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/addons"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/url"

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
				//teststep.CreateFunction(config.Log, pythonFunc.fn, pythonFunc.name, pythonFunctionBody("Hello From Python")),
				teststep.NewDefaultedFunctionCheck(pythonFunc.fn),
				//teststep.NewAPIRule(pythonFunc.apiRule, "python api rule", pythonFunc.name, domainName, 80),
				teststep.NewCheck(config.Log, "python function check", expectedPythonURL, "Hello From Python"),
			),
			teststep.NewSerialSteps(config.Log, "Node JS 12 Function tests",
				//teststep.CreateFunction(config.Log, nodejs12Func.fn, nodejs12Func.name, nodejsFunctionBody("Hello From Nodejs12")),
				//teststep.NewAPIRule(nodejs12Func.apiRule, "nodejs api rule", nodejs12Func.name, domainName, 80)
				teststep.NewCheck(config.Log, "NodeJS function check", expectedNodeJSURL, "Hello From Nodejs12"),
			),
		),

		//teststep.NewPause(10 * time.Minute),
	}
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
