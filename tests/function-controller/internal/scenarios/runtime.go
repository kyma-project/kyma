package scenarios

import (
	"fmt"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/internal/teststep"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
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

	ingressHost := "lucky-cancer.wookiee.hudy.ninja"
	return []step.Step{
		teststep.NewNamespaceStep(config.Log, config.CoreCli, namespace),

		step.Parallel(teststep.CreateFunction(config.Log, pythonFunc.fn, pythonFunc.name, pythonFunctionBody("Hello From Python")),
			teststep.CreateFunction(config.Log, nodejs12Func.fn, nodejs12Func.name, nodejsFunctionBody("Hello From Nodejs12"))),
		step.Parallel(teststep.NewAPIRule(pythonFunc.apiRule, "python api rule", pythonFunc.name, ingressHost, 80),
			teststep.NewAPIRule(nodejs12Func.apiRule, "nodejs api rule", nodejs12Func.name, ingressHost, 80)),
		step.Parallel(),
		//teststep.UpdateFunction(config.Log, nodejs10Fn, fnName, nodeFnData),
		teststep.NewPause(10 * time.Minute),
	}
}

type FunctionConfig struct {
	fn      *function.Function
	apiRule *apirule.APIRule
	name    string
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
