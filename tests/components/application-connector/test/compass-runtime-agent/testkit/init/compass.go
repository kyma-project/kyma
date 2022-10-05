package init

import (
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/init/types"
)

type compassconfigurator struct {
	directorClient types.DirectorClient
	tenant         string
}

func NewCompassConfigurator(directorClient types.DirectorClient, tenant string) compassconfigurator {
	return compassconfigurator{
		directorClient: directorClient,
		tenant:         tenant,
	}
}

func (cc compassconfigurator) Do(runtimeName, formationName string) (types.CompassRuntimeAgentConfig, types.RollbackFunc, error) {
	runtimeID, err := cc.directorClient.RegisterRuntime(runtimeName)
	if err != nil {
		return types.CompassRuntimeAgentConfig{}, nil, err
	}

	unregisterRuntimeRollbackFunc := func() error { return cc.directorClient.UnregisterRuntime(runtimeID) }

	err = cc.directorClient.RegisterFormation(formationName)
	if err != nil {
		return types.CompassRuntimeAgentConfig{}, unregisterRuntimeRollbackFunc, err
	}

	unregisterFormationRollbackFunc := func() error { return cc.directorClient.UnregisterFormation(formationName) }
	rollBackFunc := newRollbackFunc(unregisterRuntimeRollbackFunc, unregisterFormationRollbackFunc)

	err = cc.directorClient.AssignRuntimeToFormation(runtimeID, formationName)
	if err != nil {
		return types.CompassRuntimeAgentConfig{}, rollBackFunc, err
	}

	token, compassConnectorUrl, err := cc.directorClient.GetConnectionToken(runtimeID)
	if err != nil {
		return types.CompassRuntimeAgentConfig{}, rollBackFunc, err
	}

	config := types.CompassRuntimeAgentConfig{
		ConnectorUrl: compassConnectorUrl,
		RuntimeID:    runtimeID,
		Token:        token,
		Tenant:       cc.tenant,
	}

	return config, rollBackFunc, nil
}
