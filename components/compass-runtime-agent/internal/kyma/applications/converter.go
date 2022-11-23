package applications

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/normalization"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

const defaultDescription = "Description not provided"

const (
	connectedAppLabelKey = "connected-app"
	managedByLabelKey    = "applicationconnector.kyma-project.io/managed-by"
	managedByLabelValue  = "compass-runtime-agent"
)

const (
	SpecAPIType    = "API"
	SpecEventsType = "Events"
)

//go:generate mockery --name=Converter
type Converter interface {
	Do(application model.Application) v1alpha1.Application
}

type converter struct {
	nameResolver             k8sconsts.NameResolver
	centralGatewayServiceUrl string
	appSkipTLSVerify         bool
}

func NewConverter(nameResolver k8sconsts.NameResolver, centralGatewayServiceUrl string, skipVerify bool) Converter {
	return converter{nameResolver: nameResolver,
		centralGatewayServiceUrl: centralGatewayServiceUrl,
		appSkipTLSVerify:         skipVerify,
	}
}

func (c converter) Do(application model.Application) v1alpha1.Application {
	prepareLabels := func(directorLabels model.Labels) map[string]string {
		labels := make(map[string]string)
		labels[connectedAppLabelKey] = application.Name
		return labels
	}

	description := application.Description
	if description == "" {
		description = defaultDescription
	}

	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   application.Name,
			Labels: map[string]string{managedByLabelKey: managedByLabelValue},
		},
		Spec: v1alpha1.ApplicationSpec{
			Description:      description,
			SkipInstallation: false,
			SkipVerify:       c.appSkipTLSVerify, // Taken from config. Maybe later we use value from labels of the Director's app
			Labels:           prepareLabels(application.Labels),
			Services:         c.toServices(application.Name, application.ApiBundles),
			CompassMetadata:  c.toCompassMetadata(application.ID, application.SystemAuthsIDs),
		},
	}
}

func (c converter) toServices(applicationName string, bundles []model.APIBundle) []v1alpha1.Service {
	services := make([]v1alpha1.Service, 0, len(bundles))

	for _, p := range bundles {
		services = append(services, c.toService(applicationName, p))
	}

	return services
}

func (c converter) toService(applicationName string, apiBundle model.APIBundle) v1alpha1.Service {
	description := defaultDescription
	if apiBundle.Description != nil && *apiBundle.Description != "" {
		description = *apiBundle.Description
	}

	return v1alpha1.Service{
		ID:                        apiBundle.ID,
		Identifier:                "", // not available in the Director's API
		Name:                      normalization.NormalizeServiceNameWithId(apiBundle.Name, apiBundle.ID),
		AuthCreateParameterSchema: apiBundle.InstanceAuthRequestInputSchema,
		DisplayName:               apiBundle.Name,
		Description:               description,
		Entries:                   c.toServiceEntries(applicationName, apiBundle),
	}
}

func (c converter) toServiceEntries(applicationName string, apiBundle model.APIBundle) []v1alpha1.Entry {
	entries := make([]v1alpha1.Entry, 0, len(apiBundle.APIDefinitions)+len(apiBundle.EventDefinitions))

	for _, apiDefinition := range apiBundle.APIDefinitions {
		entries = append(entries, c.toAPIEntry(applicationName, apiBundle, apiDefinition))
	}

	for _, eventAPIDefinition := range apiBundle.EventDefinitions {
		entries = append(entries, c.toEventServiceEntry(eventAPIDefinition))
	}

	return entries
}

func (c converter) toAPIEntry(applicationName string, apiBundle model.APIBundle, apiDefinition model.APIDefinition) v1alpha1.Entry {

	entry := v1alpha1.Entry{
		ID:                          apiDefinition.ID,
		Name:                        apiDefinition.Name,
		Type:                        SpecAPIType,
		TargetUrl:                   apiDefinition.TargetUrl,
		CentralGatewayUrl:           c.toCentralGatewayURL(applicationName, apiBundle.Name, apiDefinition.Name),
		SpecificationUrl:            "", // Director returns BLOB here
		Credentials:                 c.toCredential(applicationName, apiBundle),
		RequestParametersSecretName: c.toRequestParametersSecretName(applicationName, apiBundle),
	}

	return entry
}

func (c converter) toRequestParametersSecretName(applicationName string, apiBundle model.APIBundle) string {
	if apiBundle.DefaultInstanceAuth != nil && apiBundle.DefaultInstanceAuth.RequestParameters != nil && !apiBundle.DefaultInstanceAuth.RequestParameters.IsEmpty() {
		return c.nameResolver.GetRequestParametersSecretName(applicationName, apiBundle.ID)
	}
	return ""
}

func (c converter) toCentralGatewayURL(applicationName string, apiBundleName string, apiDefinitionName string) string {
	return c.centralGatewayServiceUrl + "/" + applicationName +
		"/" + normalization.NormalizeName(apiBundleName) +
		"/" + normalization.NormalizeName(apiDefinitionName)
}

func (c converter) toCredential(applicationName string, apiBundle model.APIBundle) v1alpha1.Credentials {
	result := v1alpha1.Credentials{}

	if apiBundle.DefaultInstanceAuth != nil && apiBundle.DefaultInstanceAuth.Credentials != nil {
		csrfInfo := func(csrfInfo *model.CSRFInfo) *v1alpha1.CSRFInfo {
			if csrfInfo != nil {
				return &v1alpha1.CSRFInfo{TokenEndpointURL: csrfInfo.TokenEndpointURL}
			}
			return nil
		}
		if apiBundle.DefaultInstanceAuth.Credentials.Oauth != nil {
			return v1alpha1.Credentials{
				Type:              CredentialsOAuthType,
				SecretName:        c.nameResolver.GetCredentialsSecretName(applicationName, apiBundle.ID),
				AuthenticationUrl: apiBundle.DefaultInstanceAuth.Credentials.Oauth.URL,
				CSRFInfo:          csrfInfo(apiBundle.DefaultInstanceAuth.Credentials.CSRFInfo),
			}
		} else if apiBundle.DefaultInstanceAuth.Credentials.Basic != nil {
			return v1alpha1.Credentials{
				Type:       CredentialsBasicType,
				SecretName: c.nameResolver.GetCredentialsSecretName(applicationName, apiBundle.ID),
				CSRFInfo:   csrfInfo(apiBundle.DefaultInstanceAuth.Credentials.CSRFInfo),
			}
		}
	}
	return result
}

func (c converter) toEventServiceEntry(eventsDefinition model.EventAPIDefinition) v1alpha1.Entry {
	return v1alpha1.Entry{
		ID:               eventsDefinition.ID,
		Name:             eventsDefinition.Name,
		Type:             SpecEventsType,
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
