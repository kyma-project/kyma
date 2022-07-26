package application_connectivity_validator

import (
	//	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"

	//	cli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/stretchr/testify/suite"
	//	"k8s.io/client-go/rest"
	"net/http"
	"testing"
)

type ValidatorSuite struct {
	suite.Suite
}

func (gs *ValidatorSuite) SetupSuite() {
}

func (gs *ValidatorSuite) TearDownSuite() {
	_, err := http.Post("http://localhost:15000/quitquitquit", "", nil)
	gs.Nil(err)
	_, err = http.Post("http://localhost:15020/quitquitquit", "", nil)
	gs.Nil(err)
}

func TestGatewaySuite(t *testing.T) {
	suite.Run(t, new(ValidatorSuite))
}
