package remoteenv

import (
	"fmt"

	"crypto/sha1"
	"encoding/hex"
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
	"unicode"
)

func convertFromK8sType(service v1alpha1.Service) (Service, apperrors.AppError) {
	var api *ServiceAPI
	var events bool
	{
		for _, entry := range service.Entries {
			if entry.Type == specAPIType {
				api = &ServiceAPI{
					GatewayURL:       entry.GatewayUrl,
					AccessLabel:      entry.AccessLabel,
					TargetUrl:        entry.TargetUrl,
					SpecificationUrl: entry.SpecificationUrl,
					ApiType:          entry.ApiType,
					Credentials: Credentials{
						AuthenticationUrl: entry.Credentials.AuthenticationUrl,
						SecretName:        entry.Credentials.SecretName,
						Type:              entry.Credentials.Type,
					},
				}
			} else if entry.Type == specEventsType {
				events = true
			} else {
				message := fmt.Sprintf("Entry %s in Remote Environment Service definition has incorrect type. Type %s needed", entry.Type, specEventsType)
				log.Error(message)
				return Service{}, apperrors.Internal(message)
			}
		}
	}

	return Service{
		ID:                  service.ID,
		DisplayName:         service.DisplayName,
		LongDescription:     service.LongDescription,
		ShortDescription:    service.Description,
		Labels:              service.Labels,
		Identifier:          service.Identifier,
		ProviderDisplayName: service.ProviderDisplayName,
		Tags:                service.Tags,
		API:                 api,
		Events:              events,
	}, nil
}

func convertToK8sType(service Service) v1alpha1.Service {
	var serviceEntries = make([]v1alpha1.Entry, 0, 2)
	if service.API != nil {
		apiEntry := v1alpha1.Entry{
			Type:             specAPIType,
			GatewayUrl:       service.API.GatewayURL,
			AccessLabel:      service.API.AccessLabel,
			TargetUrl:        service.API.TargetUrl,
			SpecificationUrl: service.API.SpecificationUrl,
			ApiType:          service.API.ApiType,
			Credentials: v1alpha1.Credentials{
				AuthenticationUrl: service.API.Credentials.AuthenticationUrl,
				SecretName:        service.API.Credentials.SecretName,
				Type:              service.API.Credentials.Type,
			},
		}
		serviceEntries = append(serviceEntries, apiEntry)
	}

	if service.Events {
		eventsEntry := v1alpha1.Entry{Type: specEventsType}
		serviceEntries = append(serviceEntries, eventsEntry)
	}

	return v1alpha1.Service{
		ID:                  service.ID,
		Name:                createServiceName(service.DisplayName, service.ID),
		DisplayName:         service.DisplayName,
		Labels:              service.Labels,
		Identifier:          service.Identifier,
		Description:         service.ShortDescription,
		LongDescription:     service.LongDescription,
		ProviderDisplayName: service.ProviderDisplayName,
		Tags:                service.Tags,
		Entries:             serviceEntries,
	}
}

func removeService(id string, re *v1alpha1.RemoteEnvironment) {
	serviceIndex := getServiceIndex(id, re)

	if serviceIndex != -1 {
		copy(re.Spec.Services[serviceIndex:], re.Spec.Services[serviceIndex+1:])
		size := len(re.Spec.Services)
		re.Spec.Services = re.Spec.Services[:size-1]
	}
}

func replaceService(id string, re *v1alpha1.RemoteEnvironment, service v1alpha1.Service) {
	serviceIndex := getServiceIndex(id, re)

	if serviceIndex != -1 {
		re.Spec.Services[serviceIndex] = service
	}
}

func ensureServiceExists(id string, re *v1alpha1.RemoteEnvironment) apperrors.AppError {
	if !serviceExists(id, re) {
		message := fmt.Sprintf("Service with ID %s does not exist", id)

		return apperrors.NotFound(message)
	}

	return nil
}

func ensureServiceNotExists(id string, re *v1alpha1.RemoteEnvironment) apperrors.AppError {
	if serviceExists(id, re) {
		message := fmt.Sprintf("Service with ID %s already exists", id)

		return apperrors.AlreadyExists(message)
	}

	return nil
}

func serviceExists(id string, re *v1alpha1.RemoteEnvironment) bool {
	return getServiceIndex(id, re) != -1
}

func getServiceIndex(id string, re *v1alpha1.RemoteEnvironment) int {
	for i, service := range re.Spec.Services {
		if service.ID == id {
			return i
		}
	}

	return -1
}

var nonAlphaNumeric = regexp.MustCompile("[^A-Za-z0-9]+")

// createServiceName creates the OSB Service Name for given RemoteEnvironment Service.
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
