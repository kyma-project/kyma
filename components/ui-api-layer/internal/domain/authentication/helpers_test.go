package authentication_test

import (
	"github.com/kyma-project/kyma/components/idppreset/pkg/apis/authentication/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fixIDPPreset() *v1alpha1.IDPPreset {
	return &v1alpha1.IDPPreset{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "fixIDPPreset",
		},
		Spec: v1alpha1.IDPPresetSpec{
			JwksUri: "uri",
			Issuer:  "issuer",
		},
	}
}

func fixIDPPresets() []*v1alpha1.IDPPreset {
	return []*v1alpha1.IDPPreset{fixIDPPreset()}
}

func fixIDPPresetGQL() gqlschema.IDPPreset {
	return gqlschema.IDPPreset{
		JwksURI: "uri",
		Issuer:  "issuer",
	}
}

func fixIDPPresetsGQL() []gqlschema.IDPPreset {
	return []gqlschema.IDPPreset{fixIDPPresetGQL()}
}
