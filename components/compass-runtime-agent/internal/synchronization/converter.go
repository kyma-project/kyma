package synchronization

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/applications"
)

const (
	specAPIType                   = "API"
	specEventsType                = "Events"
	CredentialsOAuthType          = "OAuth"
	CredentialsBasicType          = "Basic"
	CredentialsCertificateGenType = "CertificateGen"
)

type Converter interface {
	Do(application Application) v1alpha1.Application
}

type converter struct {
}

func (c converter) Do(application Application) v1alpha1.Application {
	description := ""
	if application.Description != nil {
		description = *application.Description
	}

	return v1alpha1.Application{
		Spec: v1alpha1.ApplicationSpec{
			Description:      description,
			SkipInstallation: false,
			AccessLabel:      "",
			Labels:           nil,
		},
	}
}

func (c converter) toServices() []v1alpha1.Service {
	return nil
}

func (c converter) toAPIService(definition APIDefinition) v1alpha1.Service {

	newService := v1alpha1.Service{
		ID:         definition.ID,
		Identifier: "", // not available in the Director's API
		// What Application Registry puts here?
		Name: definition.Name,
		// Application Registry stores Name attribute from payload here
		DisplayName: "", // ??
		// Application Registry adds here ShortDescription from the payload
		Description: definition.Description,
		// Application Registry adds here an union of two things: labels specified in the payload and connectedApp label
		Labels: nil, // ??
		// Application Registry adds here Description from the payload
		LongDescription:     "",                // not available in the Director's API
		ProviderDisplayName: "",                // not available in the Director's API
		Tags:                make([]string, 0), // ??
		Entries: []v1alpha1.Entry{
			c.toServiceEntry(definition),
		},
	}

	return newService
}

func (c converter) toServiceEntry(definition APIDefinition) v1alpha1.Entry {
	entry := v1alpha1.Entry{
		Type: specAPIType,
		// TODO what is put here by the Application Registry
		AccessLabel: "",
		TargetUrl:   definition.TargetUrl,
		// Director returns BLOB here
		SpecificationUrl: "",
		// Application Registry puts a value from payload here
		ApiType: "",
		// Application Registry puts a value from payload here
		Credentials: c.toCredentials(definition.Credentials),
		// Application Registry puts a value from payload here
		RequestParametersSecretName: "",
	}

	return entry
}

func (c converter) toCredentials(credentials *Credentials) v1alpha1.Credentials {

	toCSRF := func(csrf *CSRFInfo) *v1alpha1.CSRFInfo {
		if csrf != nil {
			return &v1alpha1.CSRFInfo{
				TokenEndpointURL: csrf.TokenEndpointURL,
			}
		}

		return &v1alpha1.CSRFInfo{}
	}

	if credentials != nil {
		if credentials.Oauth != nil {
			return v1alpha1.Credentials{
				Type:              applications.CredentialsOAuthType,
				AuthenticationUrl: credentials.Oauth.URL,
				SecretName:        "",
				CSRFInfo:          toCSRF(credentials.CSRFInfo),
			}
		}

		if credentials.Basic != nil {
			return v1alpha1.Credentials{
				SecretName: "",
				CSRFInfo:   toCSRF(credentials.CSRFInfo),
			}
		}
		return v1alpha1.Credentials{}
	}

	return v1alpha1.Credentials{}
}

func (c converter) toEventAPIService(definition EventAPIDefinition) v1alpha1.Service {

	newService := v1alpha1.Service{
		ID: definition.ID,
		// Application Registry allows to specify Identifier field which represents external system identifier
		// What about UI?
		Identifier: "",
		// What application registry places here??
		Name: definition.Name,
		// Application Registry stores Name attribute from payload here
		DisplayName: "", // ??
		// Application Registry adds here ShortDescription from the payload
		Description: definition.Description,
		// Application Registry adds here an union of two things: labels specified in the payload and connectedApp label
		Labels: nil, // ??
		// Application Registry adds here Description from the payload
		LongDescription: "", // ??
		// Application Registry add here Provider from the payload
		ProviderDisplayName: "", // ??
		// Application Registry adds an empty slice here
		Tags:    nil, // ??
		Entries: nil, // Fill entries in
	}

	return newService
}

func (c converter) toEventServiceEntry(definition EventAPIDefinition) v1alpha1.Entry {
	entry := v1alpha1.Entry{
		Type: specEventsType,
		// TODO what is put here by the Application Registry
		AccessLabel: "",
		// Director returns BLOB here
		SpecificationUrl: "", // Use the same stuff as the code creating secrets
	}

	return entry
}
