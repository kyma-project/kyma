package k8sconsts

import (
	"fmt"
)

const (
	resourceNamePrefixFormat = "%s-"
	maxResourceNameLength    = 63 // Kubernetes limit for services
	uuidLength               = 36 // UUID has 36 characters
)

// NameResolver provides names for Kubernetes resources
type NameResolver interface {
	// GetCredentialsSecretName returns credential secret name
	GetCredentialsSecretName(application, packageID string) string

	// GetRequestParametersSecretName returns request parameters secret name
	GetRequestParametersSecretName(application, packageID string) string
}

type nameResolver struct{}

// NewNameResolver creates NameResolver
func NewNameResolver() NameResolver {
	return nameResolver{}
}

// GetCredentialsSecretName returns credential secret name
func (resolver nameResolver) GetCredentialsSecretName(application, packageID string) string {
	return getResourceName(fmt.Sprintf("%s-credentials", application), packageID)
}

// GetRequestParametersSecretName returns request parameters secret name
func (resolver nameResolver) GetRequestParametersSecretName(application, packageID string) string {
	return getResourceName(fmt.Sprintf("%s-params", application), packageID)
}

// GetResourceName returns resource name with given ID
func getResourceName(application, id string) string {
	return getResourceNamePrefix(application) + id
}

func getResourceNamePrefix(application string) string {
	truncatedApplicaton := truncateApplication(application)
	return fmt.Sprintf(resourceNamePrefixFormat, truncatedApplicaton)
}

func truncateApplication(application string) string {
	maxResourceNamePrefixLength := maxResourceNameLength - uuidLength
	testResourceNamePrefix := fmt.Sprintf(resourceNamePrefixFormat, application)
	testResourceNamePrefixLength := len(testResourceNamePrefix)

	overflowLength := testResourceNamePrefixLength - maxResourceNamePrefixLength

	if overflowLength > 0 {
		newApplicationLength := len(application) - overflowLength
		return application[0:newApplicationLength]
	}
	return application
}
