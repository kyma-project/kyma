package applications

import (
	"fmt"
	"strings"
)

const (
	resourceNamePrefixFormat = "app-%s-"
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
func (resolver *NameResolver) GetResourceName(applicaton, id string) string {
	return getResourceNamePrefix(applicaton) + id
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
func (resolver *NameResolver) GetGatewayUrl(applicaton, id string) string {
	return fmt.Sprintf(metadataUrlFormat, resolver.GetResourceName(applicaton, id), resolver.namespace)
}

// ExtractServiceId extracts service ID from given host
func (resolver *NameResolver) ExtractServiceId(applicaton, host string) string {
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
