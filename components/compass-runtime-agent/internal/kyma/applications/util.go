package applications

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

func ServiceExists(id string, application v1alpha1.Application) bool {
	return getServiceIndex(id, application) != -1
}

func GetService(id string, application v1alpha1.Application) v1alpha1.Service {
	for _, service := range application.Spec.Services {
		if service.ID == id {
			return service
		}
	}

	return v1alpha1.Service{}
}

func getServiceIndex(id string, application v1alpha1.Application) int {
	for i, service := range application.Spec.Services {
		if service.ID == id {
			return i
		}
	}

	return -1
}

func ApplicationExists(applicationName string, applicationList []v1alpha1.Application) bool {
	if applicationList == nil {
		return false
	}

	for _, runtimeApplication := range applicationList {
		if runtimeApplication.Name == applicationName {
			return true
		}
	}

	return false
}

func GetApplication(applicationName string, applicationList []v1alpha1.Application) v1alpha1.Application {
	if applicationList == nil {
		return v1alpha1.Application{}
	}

	for _, runtimeApplication := range applicationList {
		if runtimeApplication.Name == applicationName {
			return runtimeApplication
		}
	}

	return v1alpha1.Application{}
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
