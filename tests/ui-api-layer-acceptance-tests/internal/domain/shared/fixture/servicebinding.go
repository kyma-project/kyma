package fixture

import (
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/domain/shared"
)

func ServiceBinding(bindingName, instanceName, namespace string) shared.ServiceBinding {
	return shared.ServiceBinding{
		Name:                bindingName,
		Namespace:           namespace,
		ServiceInstanceName: instanceName,
		Secret: shared.Secret{
			Name:      bindingName,
			Namespace: namespace,
		},
		Status: shared.ServiceBindingStatus{
			Type: shared.ServiceBindingStatusTypeReady,
		},
	}
}
