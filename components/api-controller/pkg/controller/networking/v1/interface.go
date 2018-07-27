package v1

import (
	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
)

type Dto struct {
	MetaDto     meta.Dto
	Hostname    string
	ServiceName string
	ServicePort int
	Status      kymaMeta.GatewayResourceStatus
}

type Interface interface {
	Create(dto *Dto) (*kymaMeta.GatewayResource, error)
	Update(oldDto, newDto *Dto) (*kymaMeta.GatewayResource, error)
	Delete(dto *Dto) error
}
