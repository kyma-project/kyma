package application

import (
	"regexp"
	"strings"

	applicationv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

const (
	// TypeLabel is an optional label for the application custom resource to determine its type
	TypeLabel = "application-type"
)

var (
	// invalidApplicationNameSegment used to match and replace none-alphanumeric characters in the application name
	invalidApplicationNameSegment = regexp.MustCompile("\\W|_")
)

// GetCleanTypeOrName cleans the application name form none-alphanumeric characters and returns it
// if the application type label exists, it will be cleaned and returned instead of the application name
func GetCleanTypeOrName(application *applicationv1alpha1.Application) string {
	if application == nil {
		return ""
	}
	applicationName := application.Name
	for k, v := range application.Labels {
		if strings.ToLower(k) == TypeLabel {
			applicationName = v
			break
		}
	}
	return GetCleanName(applicationName)
}

// GetCleanName cleans the name form none-alphanumeric characters and returns the clean name
func GetCleanName(name string) string {
	return invalidApplicationNameSegment.ReplaceAllString(name, "")
}
