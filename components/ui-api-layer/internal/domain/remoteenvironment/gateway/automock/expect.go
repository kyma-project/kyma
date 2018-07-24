package automock

import (
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway"
)

func (gl *gatewayServiceLister) ReturnOnGetGatewayServices(result []gateway.ServiceData) {
	gl.On("ListGatewayServices").Return(result)
}
