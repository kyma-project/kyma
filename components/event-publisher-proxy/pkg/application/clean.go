package application

import (
	"regexp"
	"strings"

	applicationv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

const (
	// typeLabel is an optional label for the application to determine its type
	typeLabel = "application-type"
)

var (
	// invalidApplicationNameSegment used to match and replace none-alphanumeric characters in the application name
	invalidApplicationNameSegment = regexp.MustCompile("\\W|_")
)

// CleanName cleans the application name form none-alphanumeric characters and returns it
// if the application type label exists, it will be cleaned and returned instead of the application name
func CleanName(application *applicationv1alpha1.Application) string {
	applicationName := application.Name
	for k, v := range application.Labels {
		if strings.ToLower(k) == typeLabel {
			applicationName = v
			break
		}
	}
	return invalidApplicationNameSegment.ReplaceAllString(applicationName, "")
}

// IsCleanName returns true if the name contains alphanumeric characters only, otherwise returns false
func IsCleanName(name string) bool {
	return !invalidApplicationNameSegment.MatchString(name)
}
