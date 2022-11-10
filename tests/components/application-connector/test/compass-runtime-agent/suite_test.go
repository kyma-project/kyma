package compass_runtime_agent

import (
	"crypto/tls"
	"fmt"
	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	ccclientset "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/applications"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/director"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/graphql"
	initcra "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init"
	compassruntimeagentinittypes "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/oauth"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/random"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"testing"
	"time"
)

type CompassRuntimeAgentSuite struct {
	suite.Suite
	applicationsClientSet           *cli.Clientset
	compassConnectionClientSet      *ccclientset.Clientset
	coreClientSet                   *kubernetes.Clientset
	compassRuntimeAgentConfigurator initcra.CompassRuntimeAgentConfigurator
	directorClient                  director.Client
	appComparator                   applications.Comparator
	testConfig                      config
	rollbackTestFunc                compassruntimeagentinittypes.RollbackFunc
	formationName                   string
}

func (cs *CompassRuntimeAgentSuite) SetupSuite() {

	err := envconfig.InitWithPrefix(&cs.testConfig, "APP")
	cs.Require().Nil(err)

	cs.T().Logf("Config: %s", cs.testConfig.String())

	cs.T().Logf("Init Kubernetes APIs")
	cs.initKubernetesApis()

	cs.T().Logf("Configure Compass Runtime Agent for test")
	cs.initCompassRuntimeAgentConfigurator()
	cs.initComparators()
	cs.configureRuntimeAgent()
}

func (cs *CompassRuntimeAgentSuite) initKubernetesApis() {
	var cfg *rest.Config
	var err error

	cs.T().Logf("Initializing with in cluster config")
	cfg, err = rest.InClusterConfig()
	cs.Assert().NoError(err)

	if err != nil {
		cs.T().Logf("Initializing kubeconfig")
		kubeconfig, ok := os.LookupEnv("KUBECONFIG")
		cs.Require().True(ok)

		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		cs.Require().NoError(err)
	}

	cs.applicationsClientSet, err = cli.NewForConfig(cfg)
	cs.Require().NoError(err)

	cs.compassConnectionClientSet, err = ccclientset.NewForConfig(cfg)
	cs.Require().NoError(err)

	cs.coreClientSet, err = kubernetes.NewForConfig(cfg)
	cs.Require().NoError(err)
}

func (cs *CompassRuntimeAgentSuite) initComparators() {
	secretComparator, err := applications.NewSecretComparator(cs.Require(), cs.coreClientSet, cs.testConfig.OAuthCredentialsNamespace, cs.testConfig.SystemNamespace)
	cs.Require().NoError(err)

	applicationGetter := cs.applicationsClientSet.ApplicationconnectorV1alpha1().Applications()
	cs.appComparator, err = applications.NewComparator(cs.Assertions, secretComparator, applicationGetter, "kyma-system", "kyma-system")
}

func (cs *CompassRuntimeAgentSuite) configureRuntimeAgent() {
	var err error
	runtimeName := "cratest"
	cs.formationName = "cratest" + random.RandomString(5)

	cs.rollbackTestFunc, err = cs.compassRuntimeAgentConfigurator.Do(runtimeName, cs.formationName)
	cs.Require().NoError(err)
}

func (cs *CompassRuntimeAgentSuite) initCompassRuntimeAgentConfigurator() {
	var err error
	cs.directorClient, err = cs.makeCompassDirectorClient()
	cs.Require().NoError(err)

	cs.compassRuntimeAgentConfigurator = initcra.NewCompassRuntimeAgentConfigurator(
		initcra.NewCompassConfigurator(cs.directorClient, cs.testConfig.TestingTenant),
		initcra.NewCertificateSecretConfigurator(cs.coreClientSet),
		initcra.NewConfigurationSecretConfigurator(cs.coreClientSet),
		initcra.NewCompassConnectionCRConfiguration(cs.compassConnectionClientSet.CompassV1alpha1().CompassConnections()),
		initcra.NewDeploymentConfiguration(cs.coreClientSet, "compass-runtime-agent", cs.testConfig.CompassSystemNamespace),
		cs.testConfig.OAuthCredentialsNamespace)
}

func (cs *CompassRuntimeAgentSuite) TearDownSuite() {
	if cs.rollbackTestFunc != nil {
		cs.T().Logf("Restore Compass Runtime Agent configuration")
		err := cs.rollbackTestFunc()

		if err != nil {
			cs.T().Logf("Failed to rollback test configuration: %v", err)
		}
	}
	_, err := http.Post("http://localhost:15000/quitquitquit", "", nil)
	if err != nil {
		cs.T().Logf("Failed to quit sidecar: %v", err)
	}
	_, err = http.Post("http://localhost:15020/quitquitquit", "", nil)
	if err != nil {
		cs.T().Logf("Failed to quit sidecar: %v", err)
	}
}

func TestCompassRuntimeAgentSuite(t *testing.T) {
	suite.Run(t, new(CompassRuntimeAgentSuite))
}

func (cs *CompassRuntimeAgentSuite) makeCompassDirectorClient() (director.Client, error) {

	secretsRepo := cs.coreClientSet.CoreV1().Secrets(cs.testConfig.OAuthCredentialsNamespace)

	if secretsRepo == nil {
		return nil, fmt.Errorf("could not access secrets in %s namespace", cs.testConfig.OAuthCredentialsNamespace)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: cs.testConfig.SkipDirectorCertVerification},
		},
		Timeout: 10 * time.Second,
	}

	gqlClient := graphql.NewGraphQLClient(cs.testConfig.DirectorURL, true, cs.testConfig.SkipDirectorCertVerification)
	if gqlClient == nil {
		return nil, fmt.Errorf("could not create GraphQLClient for endpoint %s", cs.testConfig.DirectorURL)
	}

	oauthClient, err := oauth.NewOauthClient(client, secretsRepo, cs.testConfig.OAuthCredentialsSecretName)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create OAuthClient client")
	}

	return director.NewDirectorClient(gqlClient, oauthClient, cs.testConfig.TestingTenant), nil
}
