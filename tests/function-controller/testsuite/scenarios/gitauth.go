package scenarios

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/vrischmann/envconfig"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/gitrepository"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/poller"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/secret"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"

	"github.com/kyma-project/kyma/tests/function-controller/testsuite/gitops"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/teststep"
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
	runtime          serverlessv1alpha1.Runtime
	auth             *serverlessv1alpha1.RepositoryAuth
}

type config struct {
	AzureAuth  BasicAuth
	GithubAuth SSHAuth
}

type SSHAuth struct {
	Key string
}

type BasicAuth struct {
	Username string
	Password string
}

func GitAuthTestSteps(restConfig *rest.Config, cfg testsuite.Config, logf *logrus.Entry) (step.Step, error) {
	testCfg := &config{}
	if err := envconfig.InitWithPrefix(testCfg, "APP"); err != nil {
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
	poll := poller.Poller{
		MaxPollingTime:     cfg.MaxPollingTime,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		Log:                genericContainer.Log,
		DataKey:            testsuite.TestDataKey,
	}

	var testCases []testRepo
	testCases = append(testCases, getGithubTestcase(createSSHAuthSecretData(testCfg.GithubAuth)))
	testCases = append(testCases, getAzureDevopsTestcase(createBasicAuthSecretData(testCfg.AzureAuth)))

	steps := []step.Step{}
	for _, testCase := range testCases {
		testSteps, err := gitAuthFunctionTestSteps(genericContainer, testCase, poll)
		if err != nil {
			return nil, errors.Wrapf(err, "while generated test case steps")
		}
		steps = append(steps, testSteps)
	}
	return step.NewSerialTestRunner(logf, "Test Git function authentication",
		teststep.NewNamespaceStep("Create test namespace", coreCli, genericContainer),
		step.NewParallelRunner(logf, "fn_tests", steps...)), nil
}

func setupSharedContainer(restConfig *rest.Config, cfg testsuite.Config, logf *logrus.Entry) (shared.Container, error) {
	currentDate := time.Now()
	cfg.Namespace = fmt.Sprintf("%s-%dh-%dm-%d", "test-serverless-gitauth", currentDate.Hour(), currentDate.Minute(), rand.Int())

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return shared.Container{}, errors.Wrapf(err, "while creating dynamic client")
	}

	return shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   cfg.Namespace,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         logf,
	}, nil
}

func gitAuthFunctionTestSteps(genericContainer shared.Container, tr testRepo, poll poller.Poller) (step.Step, error) {
	genericContainer.Log.Infof("Testing Git Function in namespace: %s", genericContainer.Namespace)

	secret := secret.NewSecret(tr.auth.SecretName, genericContainer)

	inClusterURL, err := url.Parse(fmt.Sprintf("http://%s.%s.svc.cluster.local", tr.name, genericContainer.Namespace))
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing in-cluster URL")
	}

	return step.NewSerialTestRunner(genericContainer.Log, fmt.Sprintf("%s Function auth test", tr.provider),
		teststep.CreateSecret(
			genericContainer.Log,
			secret,
			fmt.Sprintf("Create %s Auth Secret", tr.provider),
			tr.secretData),
		teststep.NewCreateGitRepository(
			genericContainer.Log,
			gitrepository.New(fmt.Sprintf("%s-repo", tr.name), genericContainer),
			fmt.Sprintf("Create GitRepository for %s", tr.provider),
			gitops.AuthRepositorySpecWithURL(
				tr.url,
				tr.auth,
			)),
		teststep.CreateFunction(
			genericContainer.Log,
			function.NewFunction(tr.name, genericContainer),
			fmt.Sprintf("Create %s Function", tr.provider),
			gitops.GitopsFunction(
				fmt.Sprintf("%s-repo", tr.name),
				tr.baseDir,
				tr.reference,
				tr.runtime),
		),
		teststep.NewHTTPCheck(
			genericContainer.Log,
			"Git Function simple check through gateway",
			inClusterURL,
			poll, tr.expectedResponse)), nil
}

func createBasicAuthSecretData(basicAuth BasicAuth) map[string]string {
	data := map[string]string{}
	data["username"] = basicAuth.Username
	data["password"] = basicAuth.Password
	return data
}

func createSSHAuthSecretData(auth SSHAuth) map[string]string {
	return map[string]string{"key": auth.Key}
}

func getAzureDevopsTestcase(secretData map[string]string) testRepo {
	return testRepo{name: "azure-devops-func",
		provider:         "AzureDevOps",
		url:              "https://kyma-wookiee@dev.azure.com/kyma-wookiee/kyma-function/_git/kyma-function",
		baseDir:          "/code",
		reference:        "main",
		expectedResponse: "Hello Serverless",
		runtime:          serverlessv1alpha1.Nodejs14,
		auth: &serverlessv1alpha1.RepositoryAuth{
			Type:       serverlessv1alpha1.RepositoryAuthBasic,
			SecretName: "azure-devops-auth-secret",
		},
		secretData: secretData}
}

func getGithubTestcase(secretData map[string]string) testRepo {
	return testRepo{
		name:             "github-func",
		provider:         "Github",
		url:              "git@github.com:moelsayed/pyhello.git",
		baseDir:          "/",
		reference:        "main",
		expectedResponse: "hello world",
		runtime:          serverlessv1alpha1.Python39,
		auth: &serverlessv1alpha1.RepositoryAuth{
			Type:       serverlessv1alpha1.RepositoryAuthSSHKey,
			SecretName: "github-auth-secret",
		},
		secretData: secretData,
	}
}
