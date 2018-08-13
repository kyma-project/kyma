package remoteenv

import (
	"fmt"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	log "github.com/sirupsen/logrus"
)

func convertFromK8sType(service v1alpha1.Service) (Service, apperrors.AppError) {
	var api *ServiceAPI
	var events bool
	{
		for _, entry := range service.Entries {
			if entry.Type == specAPIType {
				api = &ServiceAPI{
					GatewayURL:            entry.GatewayUrl,
					AccessLabel:           entry.AccessLabel,
					TargetUrl:             entry.TargetUrl,
					OauthUrl:              entry.OauthUrl,
					CredentialsSecretName: entry.CredentialsSecretName,
				}
			} else if entry.Type == specEventsType {
				events = true
			} else {
				message := fmt.Sprintf("incorrect type of entry '%s' in Remote Environment Service definition", entry.Type)
				log.Error(message)
				return Service{}, apperrors.Internal(message)
			}
		}
	}

	return Service{
		ID:                  service.ID,
		DisplayName:         service.DisplayName,
		LongDescription:     service.LongDescription,
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
			Type:                  specAPIType,
			GatewayUrl:            service.API.GatewayURL,
			AccessLabel:           service.API.AccessLabel,
			TargetUrl:             service.API.TargetUrl,
			OauthUrl:              service.API.OauthUrl,
			CredentialsSecretName: service.API.CredentialsSecretName,
		}
		serviceEntries = append(serviceEntries, apiEntry)
	}

	if service.Events {
		eventsEntry := v1alpha1.Entry{Type: specEventsType}
		serviceEntries = append(serviceEntries, eventsEntry)
	}

	return v1alpha1.Service{
		ID:                  service.ID,
		DisplayName:         service.DisplayName,
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
		message := fmt.Sprintf("Service with ID %s doesn't exist.", id)

		return apperrors.NotFound(message)
	}

	return nil
}

func ensureServiceNotExists(id string, re *v1alpha1.RemoteEnvironment) apperrors.AppError {
	if serviceExists(id, re) {
		message := fmt.Sprintf("Service with ID %s already exists.", id)

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
