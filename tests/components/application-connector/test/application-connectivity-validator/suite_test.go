package application_connectivity_validator

import (
	"github.com/stretchr/testify/suite"
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

func TestValidatorSuite(t *testing.T) {
	suite.Run(t, new(ValidatorSuite))
}
