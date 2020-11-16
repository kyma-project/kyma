package validate

import (
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"net/http"
)

// StaticCreate runs basic Binding validation for Create operation.
type StaticCreate struct{}

var _ Validator = &StaticCreate{}

// Validate validate Binding instance
func (v *StaticCreate) Validate(binding *v1alpha1.Binding) *webhook.Error {
	err := BindingValidation(binding).ToAggregate()
	if err != nil {
		return webhook.NewError(http.StatusForbidden, err.Error())
	}
	return nil
}
