package applications

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

type NameResolver struct {
	namespace string
}

// NewNameResolver creates NameResolver that uses application name and namespace.
func NewNameResolver(namespace string) *NameResolver {
	return &NameResolver{
		namespace: namespace,
	}
}

// GetResourceName returns resource name with given ID
func (resolver *NameResolver) GetResourceName(application, id string) string {
	return getResourceNamePrefix(application) + id
}

func (resolver *NameResolver) GetCredentialsSecretName(application, id string) string {
	return resolver.GetResourceName(application, id)
}

func (resolver *NameResolver) GetRequestParamsSecretName(application, id string) string {
	name := resolver.GetResourceName(application, id)

	resourceName := fmt.Sprintf(requestParamsNameFormat, name)
	if len(resourceName) > k8sResourceNameMaxLength {
		return resourceName[0 : k8sResourceNameMaxLength-1]
	}

	return resourceName
}

// GetGatewayUrl return gateway url with given ID
func (resolver *NameResolver) GetGatewayUrl(application, id string) string {
	return fmt.Sprintf(metadataUrlFormat, resolver.GetResourceName(application, id), resolver.namespace)
}

// ExtractServiceId extracts service ID from given host
func (resolver *NameResolver) ExtractServiceId(application, host string) string {
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
