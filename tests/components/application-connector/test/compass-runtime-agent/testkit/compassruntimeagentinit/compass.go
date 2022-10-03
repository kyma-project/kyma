package compassruntimeagentinit

import (
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/compassruntimeagentinit/types"
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

	err = cc.directorClient.RegisterFormation(formationName)
	if err != nil {
		return types.CompassRuntimeAgentConfig{}, nil, err
	}

	err = cc.directorClient.AssignRuntimeToFormation(runtimeID, formationName)
	if err != nil {
		return types.CompassRuntimeAgentConfig{}, nil, err
	}

	rollBackFunc := cc.getRollbackFunction(runtimeID, formationName)

	token, compassConnectorUrl, err := cc.directorClient.GetConnectionToken(runtimeID)
	if err != nil {
		var result *multierror.Error

		multierror.Append(result, err)
		{
			err := rollBackFunc()
			if err != nil {
				multierror.Append(result, err)
			}
		}

		return types.CompassRuntimeAgentConfig{}, nil, result.ErrorOrNil()
	}

	config := types.CompassRuntimeAgentConfig{
		ConnectorUrl: compassConnectorUrl,
		RuntimeID:    runtimeID,
		Token:        token,
		Tenant:       cc.tenant,
	}

	return config, rollBackFunc, nil
}

func (cc compassconfigurator) getRollbackFunction(runtimeID, formationName string) types.RollbackFunc {
	return func() error {
		var result *multierror.Error

		if err := cc.directorClient.UnregisterRuntime(runtimeID); err != nil {
			multierror.Append(result, err)
		}

		if err := cc.directorClient.UnregisterFormation(formationName); err != nil {
			multierror.Append(result, err)
		}

		return result.ErrorOrNil()
	}
}
