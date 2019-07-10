package controller

import "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"

type protection struct{}

func (protection) removeFinalizer(slice []string) []string {
	newSlice := make([]string, 0)
	for _, item := range slice {
		if item == v1alpha1.FinalizerAddonsConfiguration {
			continue
		}
		newSlice = append(newSlice, item)
	}
	return newSlice
}

func (protection) hasFinalizer(slice []string) bool {
	for _, item := range slice {
		if item == v1alpha1.FinalizerAddonsConfiguration {
			return true
		}
	}
	return false
}

func (protection) addFinalizer(slice []string) []string {
	return append(slice, v1alpha1.FinalizerAddonsConfiguration)
}
