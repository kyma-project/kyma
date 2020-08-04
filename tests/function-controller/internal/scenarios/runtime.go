package scenarios

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/internal/teststep"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
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
	//	fnName := "test-function"
	//
	//	nodejs10Fn := function.NewFunction2("nodejs-10-test", namespace, config.DynamicCLI, config.Log)
	//	nodeFnData := function.FunctionData{
	//		Body:        fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, "damian jest fajny"),
	//		Deps:        `{ "name": "hellowithoutdeps", "version": "0.0.1", "dependencies": { } }`,
	//		MaxReplicas: 2,
	//		MinReplicas: 1,
	//		Runtime:     serverlessv1alpha1.Nodejs12,
	//	}
	//
	//	python37Fn := function.NewFunction2("python-37-test", namespace, config.DynamicCLI, config.Log)
	//	python37Data := function.FunctionData{
	//		Body:
	//		`import os
	//def main(event, context):
	//	envs = os.environ
	//	builder = ''
	//	for key in envs:
	//	  builder = builder + "key: {}, value: {}<br>".format(key, envs[key])
	//	return "hello world from final PR, envs: {}".format(builder)`,
	//		Deps:    "",
	//		Runtime: serverlessv1alpha1.Python37,
	//	}

	pythonAPIRule := apirule.New2("python", namespace, config.Log, config.DynamicCLI)
	ingressHost := "lucky-cancer.wookiee.hudy.ninja"
	pythonFunctionService := "python-37-test"
	return []step.Step{
		//teststep.NewNamespaceStep(config.Log, config.CoreCli, namespace),

		//step.Parallel(teststep.CreateFunction(config.Log, nodejs10Fn, fnName, nodeFnData),
		//	teststep.CreateFunction(config.Log, python37Fn, fnName, python37Data)),
		step.Parallel(teststep.NewAPIRule(pythonAPIRule, "python api rule", pythonFunctionService, ingressHost, 80)),
		//teststep.UpdateFunction(config.Log, nodejs10Fn, fnName, nodeFnData),
		teststep.NewPause(10 * time.Minute),
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
