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
	nameResolver k8sconsts.NameResolver
}

func NewConverter(nameResolver k8sconsts.NameResolver) Converter {
	return converter{
		nameResolver: nameResolver,
	}
}

// TODO: consider differences in Director's and Application CRD
// 1. Director provides Application Name and Application ID but we cannot store both. Application ID is used as application name in the CRD.
// 2. Director provides application labels in a form of map[string]string[] whereas application CRD expects map[string]string
// 3. Service object being a part of Application CRD contains some fields which are not returned by the Director:
// 	 1) ProviderDisplayName field ; Application Registry takes this value from the payload passed on service registration.
//	 2) LongDescription field ; Application Registry takes this value from the payload passed on service registration.
//   3) Identifier field ; Application Registry takes this value from the payload passed on service registration. The field represent an external identifier defined in the system exposing API/Events.
//   4) Labels for api definition ; Application Registry allows to specify labels to be added to Service object
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
			Name: application.ID,
		},
		Spec: v1alpha1.ApplicationSpec{
			Description:      description,
			SkipInstallation: false,
			AccessLabel:      application.ID,
			Labels:           convertLabels(application.Labels),
			Services:         c.toServices(application.ID, application.APIs, application.EventAPIs),
		},
	}
}

const (
	connectedApp = "connected-app"
)

func (c converter) toServices(applicationID string, apis []APIDefinition, eventAPIs []EventAPIDefinition) []v1alpha1.Service {
	services := make([]v1alpha1.Service, 0, len(apis)+len(eventAPIs))

	for _, apiDefinition := range apis {
		services = append(services, c.toAPIService(applicationID, apiDefinition))
	}

	for _, eventsDefinition := range eventAPIs {
		services = append(services, c.toEventAPIService(applicationID, eventsDefinition))
	}

	return services
}

func (c converter) toAPIService(applicationID string, apiDefinition APIDefinition) v1alpha1.Service {

	newService := v1alpha1.Service{
		ID:                  apiDefinition.ID,
		Identifier:          "", // not available in the Director's API
		Name:                createServiceName(apiDefinition.Name, apiDefinition.ID),
		DisplayName:         apiDefinition.Name,
		Description:         apiDefinition.Description,
		Labels:              map[string]string{connectedApp: applicationID}, // not available in the Director's API
		LongDescription:     "",                                             // not available in the Director's API
		ProviderDisplayName: "",                                             // not available in the Director's API
		Tags:                make([]string, 0),
		Entries: []v1alpha1.Entry{
			c.toServiceEntry(applicationID, apiDefinition),
		},
	}

	return newService
}

func (c converter) toServiceEntry(applicationID string, apiDefinition APIDefinition) v1alpha1.Entry {

	getRequestParamsSecretName := func() string {
		if apiDefinition.RequestParameters.Headers != nil || apiDefinition.RequestParameters.QueryParameters != nil {
			return c.nameResolver.GetRequestParamsSecretName(applicationID, apiDefinition.ID)
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
		Type:                        specAPIType,
		AccessLabel:                 c.nameResolver.GetResourceName(applicationID, apiDefinition.ID),
		ApiType:                     getApiType(),
		TargetUrl:                   apiDefinition.TargetUrl,
		GatewayUrl:                  c.nameResolver.GetGatewayUrl(applicationID, apiDefinition.ID),
		SpecificationUrl:            "", // Director returns BLOB here
		Credentials:                 c.toCredentials(applicationID, apiDefinition.ID, apiDefinition.Credentials),
		RequestParametersSecretName: getRequestParamsSecretName(),
	}

	return entry
}

func (c converter) toCredentials(applicationID string, apiDefinitionID string, credentials *Credentials) v1alpha1.Credentials {

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
				SecretName:        c.nameResolver.GetCredentialsSecretName(applicationID, apiDefinitionID),
				CSRFInfo:          toCSRF(credentials.CSRFInfo),
			}
		}

		if credentials.Basic != nil {
			return v1alpha1.Credentials{
				Type:       CredentialsBasicType,
				SecretName: c.nameResolver.GetCredentialsSecretName(applicationID, apiDefinitionID),
				CSRFInfo:   toCSRF(credentials.CSRFInfo),
			}
		}
	}

	return v1alpha1.Credentials{}
}

func (c converter) toEventAPIService(applicationID string, eventsDefinition EventAPIDefinition) v1alpha1.Service {

	newService := v1alpha1.Service{
		ID:                  eventsDefinition.ID,
		Identifier:          "", // not available in the Director's API
		Name:                createServiceName(eventsDefinition.Name, eventsDefinition.ID),
		DisplayName:         eventsDefinition.Name,
		Description:         eventsDefinition.Description,
		Labels:              map[string]string{connectedApp: applicationID}, // Application Registry adds here an union of two things: labels specified in the payload and connectedApp label
		LongDescription:     "",                                             // not available in the Director's API
		ProviderDisplayName: "",                                             // not available in the Director's API
		Tags:                make([]string, 0),
		Entries:             []v1alpha1.Entry{c.toEventServiceEntry(applicationID, eventsDefinition)},
	}

	return newService
}

func (c converter) toEventServiceEntry(applicationID string, eventsDefinition EventAPIDefinition) v1alpha1.Entry {
	entry := v1alpha1.Entry{
		Type:             specEventsType,
		AccessLabel:      c.nameResolver.GetResourceName(applicationID, eventsDefinition.ID),
		SpecificationUrl: "", // Director returns BLOB here
	}

	return entry
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
