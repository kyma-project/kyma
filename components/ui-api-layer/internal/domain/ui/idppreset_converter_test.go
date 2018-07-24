package ui

import (
	"testing"

	"github.com/kyma-project/kyma/components/idppreset/pkg/apis/ui/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIDPPresetConverter_ToGQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		fix := v1alpha1.IDPPreset{
			ObjectMeta: metav1.ObjectMeta{
				Name: "name",
			},
			Spec: v1alpha1.IDPPresetSpec{
				Name:    "name",
				Issuer:  "issuer",
				JwksUri: "jwksUri",
			},
		}

		converter := &idpPresetConverter{}

		// when
		dto := converter.ToGQL(&fix)

		// then
		assert.Equal(t, dto.Name, fix.Spec.Name)
		assert.Equal(t, dto.Issuer, fix.Spec.Issuer)
		assert.Equal(t, dto.JwksUri, fix.Spec.JwksUri)
	})

	t.Run("Empty", func(t *testing.T) {
		// given
		converter := &idpPresetConverter{}

		// when
		result := converter.ToGQL(&v1alpha1.IDPPreset{})

		// then
		assert.Empty(t, result)
	})

	t.Run("Nil", func(t *testing.T) {
		// given
		converter := &idpPresetConverter{}

		// when
		result := converter.ToGQL(nil)

		// then
		assert.Empty(t, result)
	})
}
