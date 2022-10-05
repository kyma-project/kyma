package compass_runtime_agent

import (
	"crypto/tls"
	"fmt"
	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	ccclientset "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/applications"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit"
	compassruntimeagentinittypes "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/director"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/graphql"
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
	compassRuntimeAgentConfigurator compassruntimeagentinit.CompassRuntimeAgentConfigurator
	directorClient                  director.Client
	appComparator                   applications.Comparator
	testConfig                      config
	rollbackTestFunc                compassruntimeagentinittypes.RollbackFunc
	formationName                   string
}

func (gs *CompassRuntimeAgentSuite) SetupSuite() {

	err := envconfig.InitWithPrefix(&gs.testConfig, "APP")
	gs.Require().Nil(err)

	gs.T().Logf("Config: %s", gs.testConfig.String())

	gs.T().Logf("Init Kubernetes APIs")
	gs.initKubernetesApis()

	gs.T().Logf("Configure Compass Runtime Agent for test")
	gs.initCompassRuntimeAgentConfigurator()
	gs.initComparators()
	gs.configureRuntimeAgent()
}

func (gs *CompassRuntimeAgentSuite) initKubernetesApis() {
	var cfg *rest.Config
	var err error

	gs.T().Logf("Initializing with in cluster config")
	cfg, err = rest.InClusterConfig()
	gs.Assert().NoError(err)

	if err != nil {
		gs.T().Logf("Initializing kubeconfig")
		kubeconfig, ok := os.LookupEnv("KUBECONFIG")
		gs.Require().True(ok)

		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		gs.Require().NoError(err)
	}

	gs.applicationsClientSet, err = cli.NewForConfig(cfg)
	gs.Require().NoError(err)

	gs.compassConnectionClientSet, err = ccclientset.NewForConfig(cfg)
	gs.Require().NoError(err)

	gs.coreClientSet, err = kubernetes.NewForConfig(cfg)
	gs.Require().NoError(err)
}

func (gs *CompassRuntimeAgentSuite) initComparators() {
	secretComparator, err := applications.NewSecretComparator(gs.Require(), gs.coreClientSet, gs.testConfig.TestNamespace, gs.testConfig.IntegrationNamespace)
	gs.Require().NoError(err)

	applicationGetter := gs.applicationsClientSet.ApplicationconnectorV1alpha1().Applications()
	gs.appComparator, err = applications.NewComparator(gs.Assertions, secretComparator, applicationGetter, "kyma-integration", "kyma-integration")
}

func (gs *CompassRuntimeAgentSuite) configureRuntimeAgent() {
	var err error
	runtimeName := "cratest"
	gs.formationName = "cratest" + random.RandomString(5)

	gs.rollbackTestFunc, err = gs.compassRuntimeAgentConfigurator.Do(runtimeName, gs.formationName)
	gs.Require().NoError(err)
}

func (gs *CompassRuntimeAgentSuite) initCompassRuntimeAgentConfigurator() {
	var err error
	gs.directorClient, err = gs.makeCompassDirectorClient()
	gs.Require().NoError(err)

	gs.compassRuntimeAgentConfigurator = compassruntimeagentinit.NewCompassRuntimeAgentConfigurator(
		compassruntimeagentinit.NewCompassConfigurator(gs.directorClient, gs.testConfig.TestingTenant),
		compassruntimeagentinit.NewCertificateSecretConfigurator(gs.coreClientSet),
		compassruntimeagentinit.NewConfigurationSecretConfigurator(gs.coreClientSet),
		compassruntimeagentinit.NewCompassConnectionCRConfiguration(gs.compassConnectionClientSet.CompassV1alpha1().CompassConnections()),
		compassruntimeagentinit.NewDeploymentConfiguration(gs.coreClientSet, "compass-runtime-agent", gs.testConfig.CompassSystemNamespace),
		gs.testConfig.TestNamespace)
}

func (gs *CompassRuntimeAgentSuite) TearDownSuite() {
	if gs.rollbackTestFunc != nil {
		gs.T().Logf("Restore Compass Runtime Agent configuration")
		err := gs.rollbackTestFunc()

		if err != nil {
			gs.T().Logf("Failed to rollback test configuration: %v", err)
		}
	}
	_, err := http.Post("http://localhost:15000/quitquitquit", "", nil)
	if err != nil {
		gs.T().Logf("Failed to quit sidecar: %v", err)
	}
	_, err = http.Post("http://localhost:15020/quitquitquit", "", nil)
	if err != nil {
		gs.T().Logf("Failed to quit sidecar: %v", err)
	}
}

func TestCompassRuntimeAgentSuite(t *testing.T) {
	suite.Run(t, new(CompassRuntimeAgentSuite))
}

func (gs *CompassRuntimeAgentSuite) makeCompassDirectorClient() (director.Client, error) {

	secretsRepo := gs.coreClientSet.CoreV1().Secrets(gs.testConfig.TestNamespace)

	if secretsRepo == nil {
		return nil, fmt.Errorf("could not access secrets in %s namespace", gs.testConfig.TestNamespace)
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: gs.testConfig.SkipDirectorCertVerification},
		},
		Timeout: 10 * time.Second,
	}

	gqlClient := graphql.NewGraphQLClient(gs.testConfig.DirectorURL, true, gs.testConfig.SkipDirectorCertVerification)
	if gqlClient == nil {
		return nil, fmt.Errorf("could not create GraphQLClient for endpoint %s", gs.testConfig.DirectorURL)
	}

	oauthClient, err := oauth.NewOauthClient(client, secretsRepo, gs.testConfig.OauthCredentialsSecretName)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create OAuthClient client")
	}

	return director.NewDirectorClient(gqlClient, oauthClient, gs.testConfig.TestingTenant), nil
}
