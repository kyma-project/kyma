package externalapi

import "github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

type MapToClusterIdentity func(ctx clientcontext.ClientContext) interface{}

type ApplicationIdentity struct {
	Application string `json:"application,omitempty"`
	Group       string `json:"group,omitempty"`
	Tenant      string `json:"tenant,omitempty"`
}

type RuntimeIdentity struct {
	RuntimeID string `json:"runtimeid,omitempty"`
	Group     string `json:"group,omitempty"`
	Tenant    string `json:"tenant,omitempty"`
}

func MapToApplicationIdentity(ctx clientcontext.ClientContext) interface{} {
	return ApplicationIdentity{
		Application: ctx.ID,
		Group:       ctx.Group,
		Tenant:      ctx.Tenant,
	}
}

func MapToRuntimeIdentity(ctx clientcontext.ClientContext) interface{} {
	return RuntimeIdentity{
		RuntimeID: ctx.ID,
		Group:     ctx.Group,
		Tenant:    ctx.Tenant,
	}
}
