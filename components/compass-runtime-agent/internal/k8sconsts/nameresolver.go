package k8sconsts

import (
	"fmt"
	"strings"
)

const (
	resourceNamePrefixFormat = "%s-"
	metadataUrlFormat        = "http://%s.%s.svc.cluster.local"

	maxResourceNameLength = 63 // Kubernetes limit for services
	uuidLength            = 36 // UUID has 36 characters

	k8sResourceNameMaxLength = 64

	requestParamsNameFormat = "params-%s"
)

// NameResolver provides names for Kubernetes resources
type NameResolver interface {
	// GetResourceName returns resource name with given ID
	GetResourceName(application, id string) string
	// GetGatewayUrl return gateway url with given ID
	GetGatewayUrl(application, id string) string
	// ExtractServiceId extracts service ID from given host
	ExtractServiceId(application, host string) string
	// GetCredentialsSecretName returns secret name with given ID
	GetCredentialsSecretName(application, id string) string
	// GetRequestParamsSecretName returns secret name with given ID
	GetRequestParamsSecretName(application, id string) string
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
func (resolver nameResolver) GetResourceName(application, id string) string {
	return getResourceNamePrefix(application) + id
}

func (resolver nameResolver) GetCredentialsSecretName(application, id string) string {
	return resolver.GetResourceName(application, id)
}

func (resolver nameResolver) GetRequestParamsSecretName(application, id string) string {
	name := resolver.GetResourceName(application, id)

	resourceName := fmt.Sprintf(requestParamsNameFormat, name)
	if len(resourceName) > k8sResourceNameMaxLength {
		return resourceName[0 : k8sResourceNameMaxLength-1]
	}

	return resourceName
}

// GetGatewayUrl return gateway url with given ID
func (resolver nameResolver) GetGatewayUrl(application, id string) string {
	return fmt.Sprintf(metadataUrlFormat, resolver.GetResourceName(application, id), resolver.namespace)
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
