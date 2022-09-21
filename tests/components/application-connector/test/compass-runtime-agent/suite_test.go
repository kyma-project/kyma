package compass_runtime_agent

import (
	"crypto/tls"
	"fmt"
	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/applications"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/director"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/graphql"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/oauth"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"testing"
	"time"
)

type CompassRuntimeAgentSuite struct {
	suite.Suite
	applicationsClientSet *cli.Clientset
	coreClientSet         *kubernetes.Clientset
	directorClient        director.Client
	appComparator         applications.Comparator
	testConfig            config
}

func (gs *CompassRuntimeAgentSuite) SetupSuite() {

	err := envconfig.InitWithPrefix(&gs.testConfig, "APP")
	gs.Require().Nil(err)

	cfg, err := rest.InClusterConfig()
	gs.Require().Nil(err)

	gs.applicationsClientSet, err = cli.NewForConfig(cfg)
	gs.Require().Nil(err)

	gs.coreClientSet, err = kubernetes.NewForConfig(cfg)
	gs.applicationsClientSet, err = cli.NewForConfig(cfg)
	gs.Require().Nil(err)

	gs.coreClientSet, err = kubernetes.NewForConfig(cfg)
	gs.Require().Nil(err)

	gs.T().Logf("Config: %s", gs.testConfig.String())

	gs.directorClient, err = gs.makeCompassDirectorClient()
	gs.Require().Nil(err)

	// TODO: Pass namespaces names
	secretComparator, err := applications.NewSecretComparator(gs.Require(), gs.coreClientSet, "", "")
	gs.Require().Nil(err)

	applicationGetter := gs.applicationsClientSet.ApplicationconnectorV1alpha1().Applications()
	gs.appComparator, err = applications.NewComparator(gs.Require(), secretComparator, applicationGetter, "", "")
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

func (gs *CompassRuntimeAgentSuite) makeCompassDirectorClient() (director.Client, error) {

	secretsRepo := gs.coreClientSet.CoreV1().Secrets(gs.testConfig.OauthCredentialsNamespace)

	if secretsRepo == nil {
		return nil, fmt.Errorf("could not access secrets in %s namespace", gs.testConfig.OauthCredentialsNamespace)
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
