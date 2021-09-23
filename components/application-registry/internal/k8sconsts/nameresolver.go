package k8sconsts

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/application-operator/pkg/normalization"
)

const (
	resourceNamePrefixFormat = "%s-"
	metadataUrlFormat        = "http://%s.%s.svc.cluster.local"

	maxResourceNameLength = 63 // Kubernetes limit for services
	uuidLength            = 36 // UUID has 36 characters
)

// NameResolver provides names for Kubernetes resources
type NameResolver interface {
	// GetResourceName returns resource name with given ID
	GetResourceName(application, id string) string
	// GetGatewayUrl return gateway url with given ID
	GetGatewayUrl(application, id string) string
	// GetCentralGatewayUrl return central gateway url
	GetCentralGatewayUrl(applicationName, serviceDisplayName string) string
	// ExtractServiceId extracts service ID from given host
	ExtractServiceId(applicaton, host string) string
}

type nameResolver struct {
	namespace               string
	centralGatewayUrlFormat string
}

// NewNameResolver creates NameResolver that uses application name and namespace.
func NewNameResolver(namespace, centralGatewayUrlFormat string) NameResolver {
	return nameResolver{
		namespace:               namespace,
		centralGatewayUrlFormat: centralGatewayUrlFormat,
	}
}

// GetResourceName returns resource name with given ID
func (resolver nameResolver) GetResourceName(application, id string) string {
	return getResourceNamePrefix(application) + id
}

// GetGatewayUrl return gateway url with given ID
func (resolver nameResolver) GetGatewayUrl(application, id string) string {
	return fmt.Sprintf(metadataUrlFormat, resolver.GetResourceName(application, id), resolver.namespace)
}

// TODO: Check if it's ok
func (resolver nameResolver) GetCentralGatewayUrl(applicationName, serviceDisplayName string) string {
	return fmt.Sprintf(resolver.centralGatewayUrlFormat, applicationName, normalization.NormalizeName(serviceDisplayName))
}

// ExtractServiceId extracts service ID from given host
func (resolver nameResolver) ExtractServiceId(application, host string) string {
	resourceName := strings.Split(host, ".")[0]
	return strings.TrimPrefix(resourceName, getResourceNamePrefix(application))
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
