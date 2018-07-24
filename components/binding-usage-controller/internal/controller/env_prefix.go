package controller

import (
	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
)

// envPrefixGetter get prefix from `ServiceBindingUsage`
type envPrefixGetter struct{}

func (e *envPrefixGetter) GetPrefix(bUsage *sbuTypes.ServiceBindingUsage) string {
	if bUsage.Spec.Parameters != nil && bUsage.Spec.Parameters.EnvPrefix != nil {
		return bUsage.Spec.Parameters.EnvPrefix.Name
	}
	return ""
}
