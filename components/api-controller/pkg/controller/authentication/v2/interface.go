package v2

import (
	"fmt"

	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/meta/v1"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
)

type Interface interface {
	Create(dto *Dto) (*kymaMeta.GatewayResource, error)
	Update(oldDto, newDto *Dto) (*kymaMeta.GatewayResource, error)
	Delete(dto *Dto) error
}

type JwtDefaultConfig Jwt

type Dto struct {
	MetaDto                meta.Dto
	ServiceName            string
	DisablePolicyPeersMTLS bool
	AuthenticationEnabled  bool
	Rules                  Rules
	Status                 kymaMeta.GatewayResourceStatus
}

type Rules []Rule

type Rule struct {
	Type Type
	Jwt  Jwt
}

func (r *Rule) String() string {
	return fmt.Sprintf("{Type: %s, Jwt: %s}", r.Type, r.Jwt)
}

type Type string

const (
	JwtType Type = "JWT"
)

type Jwt struct {
	Issuer  string
	JwksUri string
}
