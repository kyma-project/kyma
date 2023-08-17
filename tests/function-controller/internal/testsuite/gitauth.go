package testsuite

import (
	"encoding/base64"
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/internal"
	"github.com/kyma-project/kyma/tests/function-controller/internal/assertion"
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/function"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/runtimes"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/secret"
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"
	"time"

	"github.com/vrischmann/envconfig"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type testRepo struct {
	name             string
	provider         string
	url              string
	baseDir          string
	expectedResponse string
	reference        string
	secretData       map[string]string
	runtime          serverlessv1alpha2.Runtime
	auth             *serverlessv1alpha2.RepositoryAuth
}

type config struct {
	Azure  AzureRepo  `envconfig:"AZURE"`
	Github GithubRepo `envconfig:"GITHUB"`
}

type SSHAuth struct {
	Key string
}

type BasicAuth struct {
	Username string
	Password string
}

type GithubRepo struct {
	Reference string `envconfig:"default=main"`
	URL       string `envconfig:"default=git@github.com:kyma-project/private-fn-for-e2e-serverless-tests.git"`
	BaseDir   string `envconfig:"default=/"`
	SSHAuth
}

type AzureRepo struct {
	Reference string `envconfig:"default=main"`
	URL       string `envconfig:"default=https://kyma-wookiee@dev.azure.com/kyma-wookiee/kyma-function/_git/kyma-function"`
	BaseDir   string `envconfig:"default=/code"`
	BasicAuth
}

func GitAuthTestSteps(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (executor.Step, error) {
	testCfg := &config{}
	if err := envconfig.InitWithPrefix(testCfg, "APP_TEST"); err != nil {
		return nil, errors.Wrap(err, "while loading git auth test config")
	}

	coreCli, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating k8s core client")
	}

	genericContainer, err := setupSharedContainer(restConfig, cfg, logf)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating Shared Container")
	}
	poll := utils.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		Log:                genericContainer.Log,
		DataKey:            internal.TestDataKey,
	}

	var testCases []testRepo
	testCases = append(testCases, getAzureDevopsTestcase(testCfg))

	githubTestcase, err := getGithubTestcase(testCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "while setting github testcase")
	}
	testCases = append(testCases, githubTestcase)

	steps := []executor.Step{}
	for _, testCase := range testCases {
		testSteps, err := gitAuthFunctionTestSteps(genericContainer, testCase, poll, cfg.KubectlProxyEnabled)
		if err != nil {
			return nil, errors.Wrapf(err, "while generated test case steps")
		}
		steps = append(steps, testSteps)
	}
	return executor.NewSerialTestRunner(logf, "Test Git function authentication",
		namespace.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		executor.NewParallelRunner(logf, "fn_tests", steps...)), nil
}

func setupSharedContainer(restConfig *rest.Config, cfg internal.Config, logf *logrus.Entry) (utils.Container, error) {
	now := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%02dh%02dm%02ds", "test-serverless-gitauth", now.Hour(), now.Minute(), now.Second())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return utils.Container{}, errors.Wrapf(err, "while creating dynamic client")
	}

	return utils.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}, nil
}

func gitAuthFunctionTestSteps(genericContainer utils.Container, tr testRepo, poll utils.Poller, useProxy bool) (executor.Step, error) {
	genericContainer.Log.Infof("Testing Git Function in namespace: %s", genericContainer.Namespace)

	sc := secret.NewSecret(tr.auth.SecretName, genericContainer)

	testFn := function.NewFunction(tr.name, useProxy, genericContainer)

	return executor.NewSerialTestRunner(genericContainer.Log, fmt.Sprintf("%s Function auth test", tr.provider),
		secret.CreateSecret(
			genericContainer.Log,
			sc,
			fmt.Sprintf("Create %s Auth Secret", tr.provider),
			tr.secretData),
		function.CreateFunction(
			genericContainer.Log,
			testFn,
			fmt.Sprintf("Create %s Function", tr.provider),
			runtimes.GitopsFunction(
				tr.url,
				tr.baseDir,
				tr.reference,
				tr.runtime,
				tr.auth),
		),
		assertion.NewHTTPCheck(
			genericContainer.Log,
			"Git Function simple check through gateway",
			testFn.FunctionURL,
			poll, tr.expectedResponse)), nil
}

func createBasicAuthSecretData(basicAuth BasicAuth) map[string]string {
	return map[string]string{
		"username": basicAuth.Username,
		"password": basicAuth.Password,
	}
}

func createSSHAuthSecretData(auth SSHAuth) (map[string]string, error) {
	// this value will be base64 encoded since it's passed as an environment variable.
	// we have to decode it first before it's passed to the secret creator since it will re-encode it again.
	decoded, err := base64.StdEncoding.DecodeString(auth.Key)

	return map[string]string{"key": string(decoded)}, err
}

func getAzureDevopsTestcase(cfg *config) testRepo {
	return testRepo{name: "azure-devops-func",
		provider:         "AzureDevOps",
		url:              cfg.Azure.URL,
		baseDir:          cfg.Azure.BaseDir,
		reference:        cfg.Azure.Reference,
		expectedResponse: "Hello azure",
		runtime:          serverlessv1alpha2.NodeJs18,
		auth: &serverlessv1alpha2.RepositoryAuth{
			Type:       serverlessv1alpha2.RepositoryAuthBasic,
			SecretName: "azure-devops-auth-secret",
		},
		secretData: createBasicAuthSecretData(cfg.Azure.BasicAuth)}
}

func getGithubTestcase(cfg *config) (testRepo, error) {
	secretData, err := createSSHAuthSecretData(cfg.Github.SSHAuth)
	if err != nil {
		return testRepo{}, errors.Wrapf(err, "while decoding ssh key")
	}
	return testRepo{
		name:             "github-func",
		provider:         "Github",
		url:              cfg.Github.URL,
		baseDir:          cfg.Github.BaseDir,
		reference:        cfg.Github.Reference,
		expectedResponse: "hello github",
		runtime:          serverlessv1alpha2.Python39,
		auth: &serverlessv1alpha2.RepositoryAuth{
			Type:       serverlessv1alpha2.RepositoryAuthSSHKey,
			SecretName: "github-auth-secret",
		},
		secretData: secretData,
	}, nil
}
