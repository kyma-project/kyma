package fixture

import (
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/domain/shared"
)

func ServiceBinding(bindingName, instanceName string) shared.ServiceBinding {
	return shared.ServiceBinding{
		Name:                bindingName,
		Environment:         tester.DefaultNamespace,
		ServiceInstanceName: instanceName,
		Secret: shared.Secret{
			Name:        bindingName,
			Environment: tester.DefaultNamespace,
		},
		Status: shared.ServiceBindingStatus{
			Type: shared.ServiceBindingStatusTypeReady,
		},
	}
}
