package remoteenv

import (
	"fmt"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
)

func convertFromK8sType(service v1alpha1.Service) (Service, apperrors.AppError) {
	var api *ServiceAPI
	var events bool
	{
		for _, entry := range service.Entries {
			if entry.Type == specAPIType {
				api = &ServiceAPI{
					GatewayURL:  entry.GatewayUrl,
					AccessLabel: entry.AccessLabel,
					TargetUrl:   entry.TargetUrl,
					Credentials: convertCredentialsFromK8sType(entry.Credentials),
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

func convertCredentialsFromK8sType(credentials v1alpha1.Credentials) *Credentials {
	emptyCredentials := v1alpha1.Credentials{}
	if credentials == emptyCredentials {
		return nil
	}

	return &Credentials{
		Type:       credentials.Type,
		SecretName: credentials.SecretName,
		Url:        credentials.AuthenticationUrl,
	}
}
