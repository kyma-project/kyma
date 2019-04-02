package source

import "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"

// returns index to the first source object from given slice with given source type
// or -1 if not found
func IndexByType(slice []v1alpha1.Source, sourceType string) int {
	for i, source := range slice {
		if source.Type != sourceType {
			continue
		}
		return i
	}
	return -1
}

// returns a copy of given slice that will not contain sources with given source type
func FilterByType(sources []v1alpha1.Source, sourceType string) []v1alpha1.Source {
	var result []v1alpha1.Source
	for _, source := range sources {
		if source.Type == sourceType {
			continue
		}
		result = append(result, source)
	}
	return result
}
