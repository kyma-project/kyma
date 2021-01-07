package validate

import (
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// validateBindingName is the validation function for Binding names.
var validateBindingName = apivalidation.NameIsDNS1035Label

// BindingValidation validates a Binding and returns a list of errors.
func BindingValidation(binding *v1alpha1.Binding) field.ErrorList {
	return internalValidateBinding(binding)
}

func internalValidateBinding(binding *v1alpha1.Binding) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, apivalidation.ValidateObjectMeta(&binding.ObjectMeta, true, validateBindingName,
		field.NewPath("metadata"))...)
	return allErrs
}
