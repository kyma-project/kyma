package automock

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
)

func (m *envLister) ExpectOnListAllEnvironments(envs []gqlschema.Environment, err error) {
	m.On("List").Return(envs, err)
}

func (m *envLister) ExpectOnListEnvironmentsForRemoteEnvironment(reName string, envs []gqlschema.Environment, err error) {
	m.On("ListForRemoteEnvironment", reName).Return(envs, err)
}
