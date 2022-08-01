package helper

import (
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

const (
	kymaSystem, kubeSystem = "kyma-system", "kube-system"
)

// SetDefaults sets the defaults for excluding system namespaces.
// Returns true if changes to the logpipeline have been made
func SetDefaults(logPipeline *telemetryv1alpha1.LogPipeline) bool {
	var applicationInput = &logPipeline.Spec.Input.Application
	systemNamespaces := make(map[string]string)
	systemNamespaces[kymaSystem] = kymaSystem
	systemNamespaces[kubeSystem] = kubeSystem

	// system namespaces are excluded and selectors are not modified
	if !(applicationInput.IncludeSystemNamespaces || applicationInput.HasSelectors()) {
		applicationInput.ExcludeNamespaces = values(systemNamespaces)
		return true
	}

	// system namespaces are set to be included and need to be added to the include-list if not present
	if applicationInput.IncludeSystemNamespaces && len(applicationInput.Namespaces) > 0 {
		for _, namespace := range applicationInput.Namespaces {
			if v, found := systemNamespaces[namespace]; found {
				delete(systemNamespaces, v)
			}
			if len(systemNamespaces) == 0 {
				return false
			}
		}
		applicationInput.Namespaces = append(applicationInput.Namespaces, values(systemNamespaces)...)
		return true
	}

	// system namespaces are set to be excluded and need to be added to the exclude-list if not present
	if !applicationInput.IncludeSystemNamespaces && len(applicationInput.ExcludeNamespaces) > 0 {
		for _, namespace := range applicationInput.ExcludeNamespaces {
			if v, found := systemNamespaces[namespace]; found {
				delete(systemNamespaces, v)
			}
			if len(systemNamespaces) == 0 {
				return false
			}
		}
		applicationInput.ExcludeNamespaces = append(applicationInput.ExcludeNamespaces, values(systemNamespaces)...)
		return true
	}

	return false
}

func values(from map[string]string) []string {
	var result []string
	for _, value := range from {
		result = append(result, value)
	}
	return result
}
