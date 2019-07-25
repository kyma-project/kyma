package synchronization

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	specAPIType          = "API"
	specEventsType       = "Events"
	CredentialsOAuthType = "OAuth"
	CredentialsBasicType = "Basic"
)

type Converter interface {
	Do(application Application) v1alpha1.Application
}

type converter struct {
	namespace    string
	nameResolver k8sconsts.NameResolver
}

func NewConverter(nameResolver k8sconsts.NameResolver) Converter {
	return converter{
		nameResolver: nameResolver,
	}
}

func (c converter) Do(application Application) v1alpha1.Application {
	description := ""
	if application.Description != nil {
		description = *application.Description
	}

	convertLabels := func(directorLabels Labels) map[string]string {
		labels := make(map[string]string)

		for key, value := range directorLabels {
			newVal := strings.Join(value, ",")
			labels[key] = newVal
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
			Description:      description,
			SkipInstallation: false,
			AccessLabel:      application.Name,
			Labels:           convertLabels(application.Labels),
			Services:         c.toServices(application.Name, application.APIs, application.EventAPIs),
		},
	}
}

const (
	connectedApp = "connected-app"
)

func (c converter) toServices(application string, apis []APIDefinition, eventAPIs []EventAPIDefinition) []v1alpha1.Service {
	services := make([]v1alpha1.Service, 0, len(apis)+len(eventAPIs))

	for _, apiDefinition := range apis {
		services = append(services, c.toAPIService(application, apiDefinition))
	}

	for _, eventDefinition := range eventAPIs {
		services = append(services, c.toEventAPIService(application, eventDefinition))
	}

	return services
}

func (c converter) toAPIService(application string, definition APIDefinition) v1alpha1.Service {

	newService := v1alpha1.Service{
		ID:                  definition.ID,
		Identifier:          "", // not available in the Director's API
		Name:                createServiceName(definition.Name, definition.ID),
		DisplayName:         definition.Name,
		Description:         definition.Description,
		Labels:              map[string]string{connectedApp: application}, // not available in the Director's API
		LongDescription:     "",                                           // not available in the Director's API
		ProviderDisplayName: "",                                           // not available in the Director's API
		Tags:                make([]string, 0),
		Entries: []v1alpha1.Entry{
			c.toServiceEntry(application, definition),
		},
	}

	return newService
}

func (c converter) toServiceEntry(application string, definition APIDefinition) v1alpha1.Entry {

	getRequestParamsSecretName := func() string {
		if definition.RequestParameters.Headers != nil || definition.RequestParameters.QueryParameters != nil {
			return c.nameResolver.GetRequestParamsSecretName(application, definition.ID)
		}

		return ""
	}

	entry := v1alpha1.Entry{
		Type:                        specAPIType,
		AccessLabel:                 c.nameResolver.GetResourceName(application, definition.ID),
		TargetUrl:                   definition.TargetUrl,
		GatewayUrl:                  c.nameResolver.GetGatewayUrl(application, definition.ID),
		SpecificationUrl:            "", // Director returns BLOB here
		Credentials:                 c.toCredentials(application, definition.ID, definition.Credentials),
		RequestParametersSecretName: getRequestParamsSecretName(),
	}

	return entry
}

var nonAlphaNumeric = regexp.MustCompile("[^A-Za-z0-9]+")

func (c converter) toCredentials(application string, serviceID string, credentials *Credentials) v1alpha1.Credentials {

	toCSRF := func(csrf *CSRFInfo) *v1alpha1.CSRFInfo {
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
				SecretName:        c.nameResolver.GetCredentialsSecretName(application, serviceID),
				CSRFInfo:          toCSRF(credentials.CSRFInfo),
			}
		}

		if credentials.Basic != nil {
			return v1alpha1.Credentials{
				Type:       CredentialsBasicType,
				SecretName: c.nameResolver.GetCredentialsSecretName(application, serviceID),
				CSRFInfo:   toCSRF(credentials.CSRFInfo),
			}
		}
		return v1alpha1.Credentials{}
	}

	return v1alpha1.Credentials{}
}

func (c converter) toEventAPIService(application string, definition EventAPIDefinition) v1alpha1.Service {

	newService := v1alpha1.Service{
		ID:                  definition.ID,
		Identifier:          "", // not available in the Director's API
		Name:                createServiceName(definition.Name, definition.ID),
		DisplayName:         definition.Name,
		Description:         definition.Description,
		Labels:              map[string]string{connectedApp: application}, // Application Registry adds here an union of two things: labels specified in the payload and connectedApp label
		LongDescription:     "",                                           // not available in the Director's API
		ProviderDisplayName: "",                                           // not available in the Director's API
		Tags:                make([]string, 0),
		Entries:             []v1alpha1.Entry{c.toEventServiceEntry(application, definition)},
	}

	return newService
}

func (c converter) toEventServiceEntry(application string, definition EventAPIDefinition) v1alpha1.Entry {
	entry := v1alpha1.Entry{
		Type:             specEventsType,
		AccessLabel:      c.nameResolver.GetResourceName(application, definition.ID),
		SpecificationUrl: "", // Director returns BLOB here
	}

	return entry
}

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
