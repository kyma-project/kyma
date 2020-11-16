package validate

import (
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// validateTargetKindName is the validation function for ServiceTargetKind names.
var validateTargetKindName = apivalidation.NameIsDNSSubdomain

// TargetKindValidation validates a TargetKind and returns a list of errors.
func TargetKindValidation(targetKind *v1alpha1.TargetKind) field.ErrorList {
	return internalValidateTargetKind(targetKind)
}

func internalValidateTargetKind(targetKind *v1alpha1.TargetKind) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, apivalidation.ValidateObjectMeta(&targetKind.ObjectMeta, true, validateTargetKindName,
		field.NewPath("metadata"))...)
	return allErrs
}
