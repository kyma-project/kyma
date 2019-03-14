package authentication

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/idppreset/pkg/apis/authentication/v1alpha1"
)

type idpPresetConverter struct{}

func (c *idpPresetConverter) ToGQL(in *v1alpha1.IDPPreset) gqlschema.IDPPreset {
	if in == nil {
		return gqlschema.IDPPreset{}
	}

	return gqlschema.IDPPreset{
		Name:    in.Name,
		Issuer:  in.Spec.Issuer,
		JwksURI: in.Spec.JwksUri,
	}
}
