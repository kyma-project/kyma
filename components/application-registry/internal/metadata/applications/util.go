package applications

import (
	"fmt"

	"crypto/sha1"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	log "github.com/sirupsen/logrus"
)

func convertFromK8sType(service v1alpha1.Service) (Service, apperrors.AppError) {
	var api *ServiceAPI

	fromK8sCSRFInfo := func(csrfInfo *v1alpha1.CSRFInfo) *CSRFInfo {
		if csrfInfo == nil {
			return nil
		}

		return &CSRFInfo{
			TokenEndpointURL: csrfInfo.TokenEndpointURL,
		}
	}

	var events bool
	{
		for _, entry := range service.Entries {
			if entry.Type == specAPIType {
				api = &ServiceAPI{
					GatewayURL:                  entry.GatewayUrl,
					AccessLabel:                 entry.AccessLabel,
					TargetUrl:                   entry.TargetUrl,
					SpecificationUrl:            entry.SpecificationUrl,
					ApiType:                     entry.ApiType,
					RequestParametersSecretName: entry.RequestParametersSecretName,
					Credentials: Credentials{
						AuthenticationUrl: entry.Credentials.AuthenticationUrl,
						CSRFInfo:          fromK8sCSRFInfo(entry.Credentials.CSRFInfo),
						SecretName:        entry.Credentials.SecretName,
						Type:              entry.Credentials.Type,
					},
				}
			} else if entry.Type == specEventsType {
				events = true
			} else {
				message := fmt.Sprintf("Entry %s in Application Service definition has incorrect type. Type %s needed", entry.Type, specEventsType)
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

		toK8sCSRFInfo := func(model *CSRFInfo) *v1alpha1.CSRFInfo {
			if model == nil {
				return nil
			}

			return &v1alpha1.CSRFInfo{
				TokenEndpointURL: model.TokenEndpointURL,
			}
		}

		apiEntry := v1alpha1.Entry{
			Name:                        service.Name,
			Type:                        specAPIType,
			GatewayUrl:                  service.API.GatewayURL,
			AccessLabel:                 service.API.AccessLabel,
			TargetUrl:                   service.API.TargetUrl,
			SpecificationUrl:            service.API.SpecificationUrl,
			ApiType:                     service.API.ApiType,
			RequestParametersSecretName: service.API.RequestParametersSecretName,
			Credentials: v1alpha1.Credentials{
				AuthenticationUrl: service.API.Credentials.AuthenticationUrl,
				CSRFInfo:          toK8sCSRFInfo(service.API.Credentials.CSRFInfo),
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

func removeService(id string, app *v1alpha1.Application) {
	serviceIndex := getServiceIndex(id, app)

	if serviceIndex != -1 {
		copy(app.Spec.Services[serviceIndex:], app.Spec.Services[serviceIndex+1:])
		size := len(app.Spec.Services)
		app.Spec.Services = app.Spec.Services[:size-1]
	}
}

func replaceService(id string, app *v1alpha1.Application, service v1alpha1.Service) {
	serviceIndex := getServiceIndex(id, app)

	if serviceIndex != -1 {
		app.Spec.Services[serviceIndex] = service
	}
}

func ensureServiceExists(id string, app *v1alpha1.Application) apperrors.AppError {
	if !serviceExists(id, app) {
		message := fmt.Sprintf("Service with ID %s does not exist", id)

		return apperrors.NotFound(message)
	}

	return nil
}

func ensureServiceNotExists(id string, app *v1alpha1.Application) apperrors.AppError {
	if serviceExists(id, app) {
		message := fmt.Sprintf("Service with ID %s already exists", id)

		return apperrors.AlreadyExists(message)
	}

	return nil
}

func serviceExists(id string, app *v1alpha1.Application) bool {
	return getServiceIndex(id, app) != -1
}

func getServiceIndex(id string, app *v1alpha1.Application) int {
	for i, service := range app.Spec.Services {
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
