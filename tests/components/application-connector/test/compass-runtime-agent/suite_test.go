package compass_runtime_agent

import (
	"crypto/tls"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/director"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/graphql"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/oauth"
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
}

func (gs *CompassRuntimeAgentSuite) SetupSuite() {
	cfg, err := rest.InClusterConfig()
	gs.Require().Nil(err)

	gs.appClientSet, err = cli.NewForConfig(cfg)
	gs.Require().Nil(err)

	_, err = gs.newDirectorClient("drectorURL", "test", "oauthCreds", true, cfg)
	gs.Require().Nil(err)
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
