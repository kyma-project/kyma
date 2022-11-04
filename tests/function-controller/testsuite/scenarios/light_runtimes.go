package scenarios

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/configmap"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
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

const scenarioKey = "scenario"

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

	python39Logger := logf.WithField(scenarioKey, "python39")
	nodejs14Logger := logf.WithField(scenarioKey, "nodejs14")
	nodejs16Logger := logf.WithField(scenarioKey, "nodejs16")

	genericContainer := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}

	python39Cfg, err := runtimes.NewFunctionSimpleConfig("python39", genericContainer.WithLogger(python39Logger))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating python39 config")
	}

	nodejs14Cfg, err := runtimes.NewFunctionSimpleConfig("nodejs14", genericContainer.WithLogger(nodejs14Logger))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating nodejs14 config")
	}
	nodejs16Cfg, err := runtimes.NewFunctionSimpleConfig("nodejs16", genericContainer.WithLogger(nodejs16Logger))
	if err != nil {
		return nil, errors.Wrapf(err, "while creating nodejs16 config")
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
	return step.NewSerialTestRunner(logf, "Runtime test",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		teststep.CreateSecret(logf, pkgCfgSecret, "Create package configuration secret", pkgCfgSecretData),
		step.NewParallelRunner(logf, "Fn tests",
			step.NewSerialTestRunner(python39Logger, "Python39 test",
				teststep.CreateFunction(python39Logger, python39Cfg.Fn, "Create Python39 Function", runtimes.BasicPythonFunction("Hello From python", serverlessv1alpha2.Python39)),
				teststep.NewHTTPCheck(python39Logger, "Python39 pre update simple check through service", python39Cfg.InClusterURL, poll.WithLogger(python39Logger), "Hello From python"),
				teststep.UpdateFunction(python39Logger, python39Cfg.Fn, "Update Python39 Function", runtimes.BasicPythonFunctionWithCustomDependency("Hello From updated python", serverlessv1alpha2.Python39)),
				teststep.NewHTTPCheck(python39Logger, "Python39 post update simple check through service", python39Cfg.InClusterURL, poll.WithLogger(python39Logger), "Hello From updated python"),
			),
			step.NewSerialTestRunner(nodejs14Logger, "NodeJS14 test",
				teststep.CreateConfigMap(nodejs14Logger, cm, "Create Test ConfigMap", cmData),
				teststep.CreateSecret(nodejs14Logger, sec, "Create Test Secret", secretData),
				teststep.CreateFunction(nodejs14Logger, nodejs14Cfg.Fn, "Create NodeJS14 Function", runtimes.NodeJSFunctionWithEnvFromConfigMapAndSecret(cm.Name(), cmEnvKey, sec.Name(), secEnvKey, serverlessv1alpha2.NodeJs14)),
				teststep.NewHTTPCheck(nodejs14Logger, "NodeJS14 pre update simple check through service", nodejs14Cfg.InClusterURL, poll.WithLogger(nodejs14Logger), fmt.Sprintf("%s-%s", cmEnvValue, secEnvValue)),
				teststep.UpdateFunction(nodejs14Logger, nodejs14Cfg.Fn, "Update NodeJS14 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello From updated nodejs14", serverlessv1alpha2.NodeJs14)),
				teststep.NewHTTPCheck(nodejs14Logger, "NodeJS14 post update simple check through service", nodejs14Cfg.InClusterURL, poll.WithLogger(nodejs14Logger), "Hello From updated nodejs14"),
			),
			step.NewSerialTestRunner(nodejs16Logger, "NodeJS16 test",
				teststep.CreateFunction(nodejs16Logger, nodejs16Cfg.Fn, "Create NodeJS16 Function", runtimes.BasicNodeJSFunction("Hello from nodejs16", serverlessv1alpha2.NodeJs16)),
				teststep.NewHTTPCheck(nodejs16Logger, "NodeJS16 pre update simple check through service", nodejs16Cfg.InClusterURL, poll.WithLogger(nodejs16Logger), "Hello from nodejs16"),
				teststep.UpdateFunction(nodejs16Logger, nodejs16Cfg.Fn, "Update NodeJS16 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello from updated nodejs16", serverlessv1alpha2.NodeJs16)),
				teststep.NewHTTPCheck(nodejs16Logger, "NodeJS16 post update simple check through service", nodejs16Cfg.InClusterURL, poll.WithLogger(nodejs16Logger), "Hello from updated nodejs16"),
			),
		),
	), nil
}
