package compass_runtime_agent

import (
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/applications"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/director"
	"net/http"
	"testing"

	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CompassRuntimeAgentSuite struct {
	suite.Suite
	cli            *cli.Clientset
	kcli           *kubernetes.Clientset
	directorClient director.Client
	appComparator  applications.Comparator
}

func initDirectorClient() director.Client {
	return nil
}

func (gs *CompassRuntimeAgentSuite) SetupSuite() {
	cfg, err := rest.InClusterConfig()
	gs.Require().Nil(err)

	gs.cli, err = cli.NewForConfig(cfg)
	gs.Require().Nil(err)

	kcli, err := kubernetes.NewForConfig(cfg)
	gs.Require().Nil(err)

	// TODO Pass Tenant from configuration
	gs.directorClient, err = director.NewDirectorClient("")
	gs.Require().Nil(err)

	// TODO: Pass namespaces names
	gs.appComparator, err = applications.NewComparator(gs.Require(), kcli, "", "")
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
