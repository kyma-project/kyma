package k8sconsts

import (
	"fmt"
	"strings"
)

const (
	resourceNamePrefixFormat = "re-%s-"
	metadataUrlFormat        = "http://%s.%s.svc.cluster.local"

	maxResourceNameLength = 63 // Kubernetes limit for services
	uuidLength            = 36 // UUID has 36 characters
)

// NameResolver provides names for Kubernetes resources
type NameResolver interface {
	// GetResourceName returns resource name with given ID
	GetResourceName(remoteEnvironment, id string) string
	// GetGatewayUrl return gateway url with given ID
	GetGatewayUrl(remoteEnvironment, id string) string
	// ExtractServiceId extracts service ID from given host
	ExtractServiceId(remoteEnvironment, host string) string
}

type nameResolver struct {
	namespace string
}

// NewNameResolver creates NameResolver that uses remote environment name and namespace.
func NewNameResolver(namespace string) NameResolver {
	return nameResolver{
		namespace: namespace,
	}
}

// GetResourceName returns resource name with given ID
func (resolver nameResolver) GetResourceName(remoteEnvironment, id string) string {
	return getResourceNamePrefix(remoteEnvironment) + id
}

// GetGatewayUrl return gateway url with given ID
func (resolver nameResolver) GetGatewayUrl(remoteEnvironment, id string) string {
	return fmt.Sprintf(metadataUrlFormat, resolver.GetResourceName(remoteEnvironment, id), resolver.namespace)
}

// ExtractServiceId extracts service ID from given host
func (resolver nameResolver) ExtractServiceId(remoteEnvironment, host string) string {
	resourceName := strings.Split(host, ".")[0]
	return strings.TrimPrefix(resourceName, getResourceNamePrefix(remoteEnvironment))
}

func getResourceNamePrefix(remoteEnvironment string) string {
	truncatedRemoteEnvironment := truncateRemoteEnvironment(remoteEnvironment)
	return fmt.Sprintf(resourceNamePrefixFormat, truncatedRemoteEnvironment)
}

func truncateRemoteEnvironment(remoteEnvironment string) string {
	maxResourceNamePrefixLength := maxResourceNameLength - uuidLength
	testResourceNamePrefix := fmt.Sprintf(resourceNamePrefixFormat, remoteEnvironment)
	testResourceNamePrefixLength := len(testResourceNamePrefix)

	overflowLength := testResourceNamePrefixLength - maxResourceNamePrefixLength

	if overflowLength > 0 {
		newRemoteEnvironmentLength := len(remoteEnvironment) - overflowLength
		return remoteEnvironment[0:newRemoteEnvironmentLength]
	}
	return remoteEnvironment
}
