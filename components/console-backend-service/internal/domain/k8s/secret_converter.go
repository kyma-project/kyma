package k8s

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type secretConverter struct{}

func (*secretConverter) ToGQL(in *v1.Secret) *gqlschema.Secret {
	if in == nil {
		return nil
	}

	out := &gqlschema.Secret{
		Name:      in.Name,
		Namespace: in.Namespace,
	}
	out.Data = make(gqlschema.JSON)
	for k, v := range in.Data {
		out.Data[k] = string(v)
	}
	return out
}
