package v1

import (
	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	k8sApiExtensions "k8s.io/api/extensions/v1beta1"
)

type Dto struct {
	MetaDto     meta.Dto
	Hostname    string
	ServiceName string
	ServicePort int
}

type Interface interface {
	Get(dto *Dto) (k8sApiExtensions.Ingress, error)
	Create(dto *Dto) (*kymaMeta.GatewayResource, error)
	Update(oldDto, newDto *Dto) (*kymaMeta.GatewayResource, error)
	Delete(dto *Dto) error
}
