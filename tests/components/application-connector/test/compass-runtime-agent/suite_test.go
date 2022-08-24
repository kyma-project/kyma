package compass_runtime_agent

import (
	"crypto/tls"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/director"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/graphql"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/oauth"
	"github.com/vrischmann/envconfig"
	"net/http"
	"testing"
	"time"

	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CompassRuntimeAgentSuite struct {
	suite.Suite
	appClientSet *cli.Clientset
	testConfig   config
}

func (gs *CompassRuntimeAgentSuite) SetupSuite() {

	err := envconfig.InitWithPrefix(&gs.testConfig, "APP")
	gs.Require().Nil(err)

	cfg, err := rest.InClusterConfig()
	gs.Require().Nil(err)

	gs.appClientSet, err = cli.NewForConfig(cfg)
	gs.Require().Nil(err)

	gs.T().Logf("Config: %s", gs.testConfig.String())

	directorClient, err := gs.newDirectorClient(gs.testConfig.DirectorURL, gs.testConfig.OauthCredentialsNamespace, gs.testConfig.OauthCredentialsSecretName, gs.testConfig.SkipDirectorCertVerification, cfg)
	gs.Require().Nil(err)

	gs.T().Log("Attempt to unregister application in compass")
	//appID, err := directorClient.RegisterApplication("oko", "auto-testing", "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae")
	err = directorClient.UnregisterApplication("218a1089-fb05-47a2-b1a9-4d09d854eeab", "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae")
	gs.Require().Nil(err)

	//gs.T().Logf("Unregistered applicationID is %s", appID)

	// Checmy pokryc wsyztkie typy autoryzacji czyli miec jedna apkę która będzie miala wszystkie typy autoryzacji
	// Mozemy po prostu zahardkodować pelna mutację albo kilka mutacji
	//
	// jakie będą scenariusze?
	//
}

func (gs *CompassRuntimeAgentSuite) TearDownSuite() {
	_, err := http.Post("http://localhost:15000/quitquitquit", "", nil)
	gs.Nil(err)
	_, err = http.Post("http://localhost:15020/quitquitquit", "", nil)
	gs.Nil(err)
}

func TestCompassRuntimeAgentSuite(t *testing.T) {
	suite.Run(t, new(CompassRuntimeAgentSuite))
}

func (gs *CompassRuntimeAgentSuite) newDirectorClient(directorURL, secretNamespace, secretName string, skipDirectorCertVerification bool, config *rest.Config) (director.DirectorClient, error) {
	coreClientset, err := kubernetes.NewForConfig(config)
	gs.Require().Nil(err)
	secretsRepo := coreClientset.CoreV1().Secrets(secretNamespace)

	gqlClient := graphql.NewGraphQLClient(directorURL, true, skipDirectorCertVerification)
	oauthClient := oauth.NewOauthClient(newHTTPClient(skipDirectorCertVerification), secretsRepo, secretName)

	return director.NewDirectorClient(gqlClient, oauthClient), nil
}

func newHTTPClient(skipCertVerification bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipCertVerification},
		},
		Timeout: 30 * time.Second,
	}
}
