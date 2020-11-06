package validate

import (
	"context"
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/validation"
	log "github.com/sirupsen/logrus"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// StaticCreate runs basic Binding validation for Create operation.
type StaticCreate struct{}

var _ Validator = &StaticCreate{}

// Validate validate Binding instance
func (v *StaticCreate) Validate(ctx context.Context, req admission.Request, binding *v1alpha1.Binding, log log.FieldLogger) *webhook.Error {
	err := validation.ValidateBinding(binding).ToAggregate()
	if err != nil {
		return webhook.NewError(http.StatusForbidden, err.Error())
	}
	return nil
}
