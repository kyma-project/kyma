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

// StaticCreate runs basic TargetKind validation for Create operation.
type StaticCreate struct {}

var _ Validator = &StaticCreate{}

// Validate validate TargetKind instance
func (v *StaticCreate) Validate(ctx context.Context, req admission.Request, targetKind *v1alpha1.TargetKind, log log.FieldLogger) *webhook.Error {
	err := validation.ValidateTargetKind(targetKind).ToAggregate()
	if err != nil {
		return webhook.NewError(http.StatusForbidden, err.Error())
	}
	return nil
}
