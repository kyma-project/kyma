package scenarios

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/configmap"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/poller"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/secret"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/runtimes"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/teststep"
)

func SimpleFunctionTest(restConfig *rest.Config, cfg testsuite.Config, logf *logrus.Entry) (step.Step, error) {
	currentDate := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%dh-%dm-%d", "test-serverless-simple", currentDate.Hour(), currentDate.Minute(), rand.Int())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating dynamic client")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
	}

	python38Logger := logf.WithField(scenarioKey, "python38")
	nodejs12Logger := logf.WithField(scenarioKey, "nodejs12")
	nodejs14Logger := logf.WithField(scenarioKey, "nodejs14")

	genericContainer := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	nodejs12Cfg, err := runtimes.NewFunctionSimpleConfig("nodejs12", genericContainer.WithLogger(nodejs12Logger))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating nodejs12 config")
	}

	python38Cfg, err := runtimes.NewFunctionSimpleConfig("python38", genericContainer.WithLogger(python38Logger))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating python38 config")
	}

	nodejs14Cfg, err := runtimes.NewFunctionSimpleConfig("nodejs14", genericContainer.WithLogger(nodejs14Logger))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating nodejs14 config")
	}

	cm := configmap.NewConfigMap("test-serverless-configmap", genericContainer.WithLogger(nodejs14Logger))
	cmEnvKey := "CM_ENV_KEY"
	cmEnvValue := "Value taken as env from ConfigMap"
	cmData := map[string]string{
		cmEnvKey: cmEnvValue,
	}
	sec := secret.NewSecret("test-serverless-secret", genericContainer.WithLogger(nodejs14Logger))
	secEnvKey := "SECRET_ENV_KEY"
	secEnvValue := "Value taken as env from Secret"
	secretData := map[string]string{
		secEnvKey: secEnvValue,
	}

	pkgCfgSecret := secret.NewSecret(cfg.PackageRegistryConfigSecretName, genericContainer)
	pkgCfgSecretData := map[string]string{
		".npmrc":   fmt.Sprintf("@kyma:registry=%s\nalways-auth=true", cfg.PackageRegistryConfigURLNode),
		"pip.conf": fmt.Sprintf("[global]\nextra-index-url = %s", cfg.PackageRegistryConfigURLPython),
	}

	logf.Infof("Testing function in namespace: %s", cfg.Namespace)

	poll := poller.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            testsuite.TestDataKey,
	}
	return step.NewSerialTestRunner(logf, "runtime Test",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		teststep.CreateSecret(logf, pkgCfgSecret, "Create package configuration secret", pkgCfgSecretData),
		step.NewParallelRunner(logf, "fn_tests",
			step.NewSerialTestRunner(python38Logger, "Python38 test",
				teststep.CreateFunction(python38Logger, python38Cfg.Fn, "Create Python38 Function", runtimes.BasicPythonFunction("Hello From python")),
				teststep.NewHTTPCheck(python38Logger, "Python38 pre update simple check through service", python38Cfg.InClusterURL, poll.WithLogger(python38Logger), "Hello From python"),
				teststep.UpdateFunction(python38Logger, python38Cfg.Fn, "Update Python38 Function", runtimes.BasicPythonFunctionWithCustomDependency("Hello From updated python")),
				teststep.NewHTTPCheck(python38Logger, "Python38 post update simple check through service", python38Cfg.InClusterURL, poll.WithLogger(python38Logger), "Hello From updated python"),
			),
			step.NewSerialTestRunner(nodejs12Logger, "NodeJS12 test",
				teststep.CreateFunction(nodejs12Logger, nodejs12Cfg.Fn, "Create NodeJS12 Function", runtimes.BasicNodeJSFunction("Hello From nodejs", serverlessv1alpha1.Nodejs12)),
				teststep.NewDefaultedFunctionCheck("Check NodeJS12 function has correct default values", nodejs12Cfg.Fn),
				teststep.UpdateFunction(nodejs12Logger, nodejs12Cfg.Fn, "Update NodeJS12 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello From updated nodejs12", serverlessv1alpha1.Nodejs12)),
				teststep.NewHTTPCheck(nodejs12Logger, "NodeJS12 pre update simple check through service", nodejs12Cfg.InClusterURL, poll.WithLogger(nodejs12Logger), "Hello From updated nodejs12"),
			),
			step.NewSerialTestRunner(nodejs14Logger, "NodeJS14 test",
				teststep.CreateConfigMap(nodejs14Logger, cm, "Create Test ConfigMap", cmData),
				teststep.CreateSecret(nodejs14Logger, sec, "Create Test Secret", secretData),
				teststep.CreateFunction(nodejs14Logger, nodejs14Cfg.Fn, "Create NodeJS14 Function", runtimes.NodeJSFunctionWithEnvFromConfigMapAndSecret(cm.Name(), cmEnvKey, sec.Name(), secEnvKey, serverlessv1alpha1.Nodejs14)),
				teststep.NewHTTPCheck(nodejs14Logger, "NodeJS14 pre update simple check through service", nodejs14Cfg.InClusterURL, poll.WithLogger(nodejs14Logger), fmt.Sprintf("%s-%s", cmEnvValue, secEnvValue)),
				teststep.UpdateFunction(nodejs14Logger, nodejs14Cfg.Fn, "Update NodeJS14 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello From updated nodejs14", serverlessv1alpha1.Nodejs14)),
				teststep.NewHTTPCheck(nodejs14Logger, "NodeJS14 post update simple check through service", nodejs14Cfg.InClusterURL, poll.WithLogger(nodejs14Logger), "Hello From updated nodejs14"),
			),
		),
	), nil
}
