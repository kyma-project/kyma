package application_gateway_test

import (
	"net/http"
	"testing"

	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/rest"
)

type GatewaySuite struct {
	suite.Suite
	cli *cli.Clientset
}

func (gs *GatewaySuite) SetupSuite() {
	cfg, err := rest.InClusterConfig()
	gs.Require().Nil(err)

	gs.cli, err = cli.NewForConfig(cfg)
	gs.Require().Nil(err)
}

func (gs *GatewaySuite) TearDownSuite() {
	_, err := http.Post("http://localhost:15000/quitquitquit", "", nil)
	gs.Nil(err)
	_, err = http.Post("http://localhost:15020/quitquitquit", "", nil)
	gs.Nil(err)
}

func TestGatewaySuite(t *testing.T) {
	suite.Run(t, new(GatewaySuite))
}
