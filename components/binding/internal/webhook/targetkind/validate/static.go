package validate

import (
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"net/http"
)

// StaticCreate runs basic TargetKind validation for Create operation.
type StaticCreate struct{}

var _ Validator = &StaticCreate{}

// Validate validate TargetKind instance
func (v *StaticCreate) Validate(targetKind *v1alpha1.TargetKind) *webhook.Error {
	err := TargetKindValidation(targetKind).ToAggregate()
	if err != nil {
		return webhook.NewError(http.StatusForbidden, err.Error())
	}
	return nil
}
