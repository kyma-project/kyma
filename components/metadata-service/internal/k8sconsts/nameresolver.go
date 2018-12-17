package k8sconsts

import (
	"fmt"
	"strings"
)

const (
	resourceNamePrefixFormat = "app-%s-"
	metadataUrlFormat        = "http://%s.%s.svc.cluster.local"

	maxResourceNameLength = 63 // Kubernetes limit for services
	uuidLength            = 36 // UUID has 36 characters
)

// NameResolver provides names for Kubernetes resources
type NameResolver interface {
	// GetResourceName returns resource name with given ID
	GetResourceName(applicaton, id string) string
	// GetGatewayUrl return gateway url with given ID
	GetGatewayUrl(applicaton, id string) string
	// ExtractServiceId extracts service ID from given host
	ExtractServiceId(applicaton, host string) string
}

type nameResolver struct {
	namespace string
}

// NewNameResolver creates NameResolver that uses application name and namespace.
func NewNameResolver(namespace string) NameResolver {
	return nameResolver{
		namespace: namespace,
	}
}

// GetResourceName returns resource name with given ID
func (resolver nameResolver) GetResourceName(applicaton, id string) string {
	return getResourceNamePrefix(applicaton) + id
}

// GetGatewayUrl return gateway url with given ID
func (resolver nameResolver) GetGatewayUrl(applicaton, id string) string {
	return fmt.Sprintf(metadataUrlFormat, resolver.GetResourceName(applicaton, id), resolver.namespace)
}

// ExtractServiceId extracts service ID from given host
func (resolver nameResolver) ExtractServiceId(applicaton, host string) string {
	resourceName := strings.Split(host, ".")[0]
	return strings.TrimPrefix(resourceName, getResourceNamePrefix(applicaton))
}

func getResourceNamePrefix(applicaton string) string {
	truncatedApplicaton := truncateapplicaton(applicaton)
	return fmt.Sprintf(resourceNamePrefixFormat, truncatedApplicaton)
}

func truncateapplicaton(applicaton string) string {
	maxResourceNamePrefixLength := maxResourceNameLength - uuidLength
	testResourceNamePrefix := fmt.Sprintf(resourceNamePrefixFormat, applicaton)
	testResourceNamePrefixLength := len(testResourceNamePrefix)

	overflowLength := testResourceNamePrefixLength - maxResourceNamePrefixLength

	if overflowLength > 0 {
		newApplicationLength := len(applicaton) - overflowLength
		return applicaton[0:newApplicationLength]
	}
	return applicaton
}
