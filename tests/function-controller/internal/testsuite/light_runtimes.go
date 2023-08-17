package testsuite

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/internal"
	"github.com/kyma-project/kyma/tests/function-controller/internal/check"
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/configmap"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/function"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/runtimes"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/secret"
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"
	"time"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const runtimeKey = "runtime"

func SimpleFunctionTest(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-serverless-simple", now.Hour(), now.Minute(), now.Second())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating dynamic client")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s CoreV1Client")
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

	python39Fn := function.NewFunction("python39", cfg.KubectlProxyEnabled, genericContainer.WithLogger(python39Logger))

	nodejs16Fn := function.NewFunction("nodejs16", cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs16Logger))

	nodejs18Fn := function.NewFunction("nodejs18", cfg.KubectlProxyEnabled, genericContainer.WithLogger(nodejs18Logger))

	cm := configmap.NewConfigMap("test-serverless-configmap", genericContainer.WithLogger(nodejs18Logger))
	cmEnvKey := "CM_ENV_KEY"
	cmEnvValue := "Value taken as env from ConfigMap"
	cmData := map[string]string{
		cmEnvKey: cmEnvValue,
	}
	sec := secret.NewSecret("test-serverless-secret", genericContainer.WithLogger(nodejs18Logger))
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

	poll := utils.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		DataKey:            internal.TestDataKey,
	}
	return executor.NewSerialTestRunner(logf, "Runtime test",
		namespace.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		secret.CreateSecret(logf, pkgCfgSecret, "Create package configuration secret", pkgCfgSecretData),
		executor.NewParallelRunner(logf, "Fn tests",
			executor.NewSerialTestRunner(python39Logger, "Python39 test",
				function.CreateFunction(python39Logger, python39Fn, "Create Python39 Function", runtimes.BasicPythonFunction("Hello From python", serverlessv1alpha2.Python39)),
				check.NewHTTPCheck(python39Logger, "Python39 pre update simple check through service", python39Fn.FunctionURL, poll, "Hello From python"),
				function.UpdateFunction(python39Logger, python39Fn, "Update Python39 Function", runtimes.BasicPythonFunctionWithCustomDependency("Hello From updated python", serverlessv1alpha2.Python39)),
				check.NewHTTPCheck(python39Logger, "Python39 post update simple check through service", python39Fn.FunctionURL, poll, "Hello From updated python"),
			),
			executor.NewSerialTestRunner(nodejs16Logger, "NodeJS16 test",
				function.CreateFunction(nodejs16Logger, nodejs16Fn, "Create NodeJS16 Function", runtimes.BasicNodeJSFunction("Hello from nodejs16", serverlessv1alpha2.NodeJs16)),
				check.NewHTTPCheck(nodejs16Logger, "NodeJS16 pre update simple check through service", nodejs16Fn.FunctionURL, poll, "Hello from nodejs16"),
				function.UpdateFunction(nodejs16Logger, nodejs16Fn, "Update NodeJS16 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello from updated nodejs16", serverlessv1alpha2.NodeJs16)),
				check.NewHTTPCheck(nodejs16Logger, "NodeJS16 post update simple check through service", nodejs16Fn.FunctionURL, poll, "Hello from updated nodejs16"),
			),
			executor.NewSerialTestRunner(nodejs18Logger, "NodeJS18 test",
				configmap.CreateConfigMap(nodejs18Logger, cm, "Create Test ConfigMap", cmData),
				secret.CreateSecret(nodejs18Logger, sec, "Create Test Secret", secretData),
				function.CreateFunction(nodejs18Logger, nodejs18Fn, "Create NodeJS18 Function", runtimes.NodeJSFunctionWithEnvFromConfigMapAndSecret(cm.Name(), cmEnvKey, sec.Name(), secEnvKey, serverlessv1alpha2.NodeJs18)),
				check.NewHTTPCheck(nodejs18Logger, "NodeJS18 pre update simple check through service", nodejs18Fn.FunctionURL, poll, fmt.Sprintf("%s-%s", cmEnvValue, secEnvValue)),
				function.UpdateFunction(nodejs18Logger, nodejs18Fn, "Update NodeJS18 Function", runtimes.BasicNodeJSFunctionWithCustomDependency("Hello from updated nodejs18", serverlessv1alpha2.NodeJs18)),
				check.NewHTTPCheck(nodejs18Logger, "NodeJS18 post update simple check through service", nodejs18Fn.FunctionURL, poll, "Hello from updated nodejs18"),
			),
		),
	), nil
}
