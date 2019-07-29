package applications

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
)

func removeService(id string, re *v1alpha1.Application) {
	serviceIndex := getServiceIndex(id, re)

	if serviceIndex != -1 {
		copy(re.Spec.Services[serviceIndex:], re.Spec.Services[serviceIndex+1:])
		size := len(re.Spec.Services)
		re.Spec.Services = re.Spec.Services[:size-1]
	}
}

func replaceService(id string, re *v1alpha1.Application, service v1alpha1.Service) {
	serviceIndex := getServiceIndex(id, re)

	if serviceIndex != -1 {
		re.Spec.Services[serviceIndex] = service
	}
}

func ensureServiceExists(id string, re *v1alpha1.Application) apperrors.AppError {
	if !serviceExists(id, re) {
		message := fmt.Sprintf("Service with ID %s does not exist", id)

		return apperrors.NotFound(message)
	}

	return nil
}

func ensureServiceNotExists(id string, re *v1alpha1.Application) apperrors.AppError {
	if serviceExists(id, re) {
		message := fmt.Sprintf("Service with ID %s already exists", id)

		return apperrors.AlreadyExists(message)
	}

	return nil
}

func serviceExists(id string, re *v1alpha1.Application) bool {
	return getServiceIndex(id, re) != -1
}

func getServiceIndex(id string, re *v1alpha1.Application) int {
	for i, service := range re.Spec.Services {
		if service.ID == id {
			return i
		}
	}

	return -1
}

var nonAlphaNumeric = regexp.MustCompile("[^A-Za-z0-9]+")

// createServiceName creates the OSB Service Name for given Application Service.
// The OSB Service Name is used in the Service Catalog as the clusterServiceClassExternalName, so it need to be normalized.
//
// Normalization rules:
// - MUST only contain lowercase characters, numbers and hyphens (no spaces).
// - MUST be unique across all service objects returned in this response. MUST be a non-empty string.
func createServiceName(serviceDisplayName, id string) string {
	// generate 5 characters suffix from the id
	sha := sha1.New()
	sha.Write([]byte(id))
	suffix := hex.EncodeToString(sha.Sum(nil))[:5]
	// remove all characters, which is not alpha numeric
	serviceDisplayName = nonAlphaNumeric.ReplaceAllString(serviceDisplayName, "-")
	// to lower
	serviceDisplayName = strings.Map(unicode.ToLower, serviceDisplayName)
	// trim dashes if exists
	serviceDisplayName = strings.TrimSuffix(serviceDisplayName, "-")
	if len(serviceDisplayName) > 57 {
		serviceDisplayName = serviceDisplayName[:57]
	}
	// add suffix
	serviceDisplayName = fmt.Sprintf("%s-%s", serviceDisplayName, suffix)
	// remove dash prefix if exists
	//  - can happen, if the name was empty before adding suffix empty or had dash prefix
	serviceDisplayName = strings.TrimPrefix(serviceDisplayName, "-")
	return serviceDisplayName
}
