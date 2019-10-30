package k8sconsts

import (
	"fmt"
	"strings"
)

const (
	resourceNamePrefixFormat = "%s-"

	maxResourceNameLength = 63 // Kubernetes limit for services
	uuidLength            = 36 // UUID has 36 characters
)

// NameResolver provides names for Kubernetes resources
type NameResolver interface {
	// GetResourceName returns resource name with given ID
	GetResourceName(id string) string
	// ExtractServiceId extracts service ID from given host
	ExtractServiceId(host string) string
}

type nameResolver struct {
	resourceNamePrefix string
}

// NewNameResolver creates NameResolver that uses application name and namespace.
func NewNameResolver(application string) NameResolver {
	return nameResolver{
		resourceNamePrefix: getResourceNamePrefix(application),
	}
}

// GetResourceName returns resource name with given ID
func (resolver nameResolver) GetResourceName(id string) string {
	return resolver.resourceNamePrefix + id
}

// ExtractServiceId extracts service ID from given host
func (resolver nameResolver) ExtractServiceId(host string) string {
	resourceName := strings.Split(host, ".")[0]
	return strings.TrimPrefix(resourceName, resolver.resourceNamePrefix)
}

func getResourceNamePrefix(application string) string {
	truncatedApplication := truncateApplication(application)
	return fmt.Sprintf(resourceNamePrefixFormat, truncatedApplication)
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
