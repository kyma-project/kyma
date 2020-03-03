package applications

import (
	"strings"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"kyma-project.io/compass-runtime-agent/internal/k8sconsts"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SpecAPIType          = "API"
	SpecEventsType       = "Events"
	CredentialsOAuthType = "OAuth"
	CredentialsBasicType = "Basic"
)

//go:generate mockery -name=Converter
type Converter interface {
	Do(application model.Application) v1alpha1.Application
}

type converter struct {
	nameResolver k8sconsts.NameResolver
}

func NewConverter(nameResolver k8sconsts.NameResolver) Converter {
	return converter{
		nameResolver: nameResolver,
	}
}

// TODO: consider differences in Director's and Application CRD
// 1. Director provides application labels in a form of map[string]string[] whereas application CRD expects map[string]string
// 2. Service object being a part of Application CRD contains some fields which are not returned by the Director:
// 	 1) ProviderDisplayName field ; Application Registry takes this value from the payload passed on service registration.
//	 2) LongDescription field ; Application Registry takes this value from the payload passed on service registration.
//   3) Identifier field ; Application Registry takes this value from the payload passed on service registration. The field represent an external identifier defined in the system exposing API/Events.
//   4) Labels for api definition ; Application Registry allows to specify labels to be added to Service object
func (c converter) Do(application model.Application) v1alpha1.Application {

	convertLabels := func(directorLabels model.Labels) map[string]string {
		labels := make(map[string]string)

		for key, value := range directorLabels {
			switch value.(type) {
			case string:
				labels[key] = value.(string)
				break
			case []string:
				newVal := strings.Join(value.([]string), ",")
				labels[key] = newVal
				break
			}
		}

		return labels
	}

	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: application.Name,
		},
		Spec: v1alpha1.ApplicationSpec{
			Description:      application.Description,
			SkipInstallation: false,
			AccessLabel:      application.Name,
			Labels:           convertLabels(application.Labels),
			Services:         c.toServices(application.Name, application.ProviderDisplayName, application.APIs, application.EventAPIs),
			CompassMetadata:  c.toCompassMetadata(application.ID, application.SystemAuthsIDs),
		},
	}
}

const (
	connectedApp = "connected-app"
)

func (c converter) toServices(applicationName, appProvider string, apis []model.APIDefinitionWithAuth, eventAPIs []model.EventAPIDefinition) []v1alpha1.Service {
	services := make([]v1alpha1.Service, 0, len(apis)+len(eventAPIs))

	for _, apiDefinition := range apis {
		services = append(services, c.toAPIService(applicationName, appProvider, apiDefinition))
	}

	for _, eventsDefinition := range eventAPIs {
		services = append(services, c.toEventAPIService(applicationName, appProvider, eventsDefinition))
	}

	return services
}

func (c converter) toAPIService(applicationName, appProvider string, apiDefinition model.APIDefinitionWithAuth) v1alpha1.Service {

	description := apiDefinition.Description
	if description == "" {
		description = "Description not provided"
	}

	return v1alpha1.Service{
		ID:                  apiDefinition.ID,
		Identifier:          "", // not available in the Director's API
		Name:                createServiceName(apiDefinition.Name, apiDefinition.ID),
		DisplayName:         apiDefinition.Name,
		Description:         description,
		Labels:              map[string]string{connectedApp: applicationName}, // not available in the Director's API
		LongDescription:     "",                                               // not available in the Director's API
		ProviderDisplayName: appProvider,
		Tags:                make([]string, 0),
		Entries: []v1alpha1.Entry{
			c.toServiceEntry(applicationName, apiDefinition),
		},
	}
}

func (c converter) toServiceEntry(applicationName string, apiDefinition model.APIDefinitionWithAuth) v1alpha1.Entry {

	getRequestParamsSecretName := func() string {
		if apiDefinition.RequestParameters.Headers != nil || apiDefinition.RequestParameters.QueryParameters != nil {
			return c.nameResolver.GetRequestParamsSecretName(applicationName, apiDefinition.ID)
		}

		return ""
	}

	getApiType := func() string {
		if apiDefinition.APISpec != nil {
			return string(apiDefinition.APISpec.Type)
		}

		return ""
	}

	entry := v1alpha1.Entry{
		Type:                        SpecAPIType,
		AccessLabel:                 c.nameResolver.GetResourceName(applicationName, apiDefinition.ID),
		ApiType:                     getApiType(),
		TargetUrl:                   apiDefinition.TargetUrl,
		GatewayUrl:                  c.nameResolver.GetGatewayUrl(applicationName, apiDefinition.ID),
		SpecificationUrl:            "", // Director returns BLOB here
		Credentials:                 c.toCredentials(applicationName, apiDefinition.ID, apiDefinition.Credentials),
		RequestParametersSecretName: getRequestParamsSecretName(),
	}

	return entry
}

func (c converter) toCredentials(applicationName string, apiDefinitionID string, credentials *model.Credentials) v1alpha1.Credentials {

	toCSRF := func(csrf *model.CSRFInfo) *v1alpha1.CSRFInfo {
		if csrf != nil {
			return &v1alpha1.CSRFInfo{
				TokenEndpointURL: csrf.TokenEndpointURL,
			}
		}

		return nil
	}

	if credentials != nil {
		if credentials.Oauth != nil {
			return v1alpha1.Credentials{
				Type:              CredentialsOAuthType,
				AuthenticationUrl: credentials.Oauth.URL,
				SecretName:        c.nameResolver.GetCredentialsSecretName(applicationName, apiDefinitionID),
				CSRFInfo:          toCSRF(credentials.CSRFInfo),
			}
		}

		if credentials.Basic != nil {
			return v1alpha1.Credentials{
				Type:       CredentialsBasicType,
				SecretName: c.nameResolver.GetCredentialsSecretName(applicationName, apiDefinitionID),
				CSRFInfo:   toCSRF(credentials.CSRFInfo),
			}
		}
	}

	return v1alpha1.Credentials{}
}

func (c converter) toEventAPIService(applicationName, appProvider string, eventsDefinition model.EventAPIDefinition) v1alpha1.Service {
	description := eventsDefinition.Description
	if description == "" {
		description = "Description not provided"
	}

	return v1alpha1.Service{
		ID:                  eventsDefinition.ID,
		Identifier:          "", // not available in the Director's API
		Name:                createServiceName(eventsDefinition.Name, eventsDefinition.ID),
		DisplayName:         eventsDefinition.Name,
		Description:         description,
		Labels:              map[string]string{connectedApp: applicationName}, // Application Registry adds here an union of two things: labels specified in the payload and connectedApp label
		LongDescription:     "",                                               // not available in the Director's API
		ProviderDisplayName: appProvider,
		Tags:                make([]string, 0),
		Entries:             []v1alpha1.Entry{c.toEventServiceEntry(applicationName, eventsDefinition)},
	}
}

func (c converter) toEventServiceEntry(applicationName string, eventsDefinition model.EventAPIDefinition) v1alpha1.Entry {
	return v1alpha1.Entry{
		Type:             SpecEventsType,
		AccessLabel:      c.nameResolver.GetResourceName(applicationName, eventsDefinition.ID),
		SpecificationUrl: "", // Director returns BLOB here
	}
}

func (c converter) toCompassMetadata(applicationID string, systemAuthsIDs []string) *v1alpha1.CompassMetadata {
	return &v1alpha1.CompassMetadata{
		ApplicationID: applicationID,
		Authentication: v1alpha1.Authentication{
			ClientIds: systemAuthsIDs,
		},
	}
}
