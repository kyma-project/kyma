package automock

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/gateway"
)

func (gl *gatewayServiceLister) ReturnOnGetGatewayServices(result []gateway.ServiceData) {
	gl.On("ListGatewayServices").Return(result)
}
