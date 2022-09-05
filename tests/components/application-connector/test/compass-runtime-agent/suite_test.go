package compass_runtime_agent

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	cli "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CompassRuntimeAgentSuite struct {
	suite.Suite
	cli cli.Interface
}

func (gs *CompassRuntimeAgentSuite) SetupSuite() {
	cfg, err := rest.InClusterConfig()
	gs.Require().Nil(err)

	gs.cli, err = cli.NewForConfig(cfg)
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
