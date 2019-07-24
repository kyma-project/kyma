package synchronization

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
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
	return v1alpha1.Application{}
}

func (c converter) apiToService(definition graphql.APIDefinition) v1alpha1.Service {

	description := ""

	if definition.Description != nil {
		description = *definition.Description
	}

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
		Description: description,
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

//Type       string `json:"type"`
//GatewayUrl string `json:"gatewayUrl"`
//// AccessLabel is not required for Events, 'omitempty' is needed because of regexp validation
//AccessLabel                 string      `json:"accessLabel,omitempty"`
//TargetUrl                   string      `json:"targetUrl"`
//SpecificationUrl            string      `json:"specificationUrl,omitempty"`
//ApiType                     string      `json:"apiType,omitempty"`
//Credentials                 Credentials `json:"credentials,omitempty"`
//RequestParametersSecretName string      `json:"requestParametersSecretName,omitempty"

func (c converter) apiToServiceEntry(definition graphql.APIDefinition) v1alpha1.Entry {
	entry := v1alpha1.Entry{
		Type: specAPIType,
		// TODO what is put here by the Application Registry
		AccessLabel: "",
		TargetUrl:   definition.TargetURL,
		// Director returns BLOB here
		SpecificationUrl: "",
		// Application Registry puts a value from payload here
		ApiType: "",
		// Application Registry puts a value from payload here
		Credentials: v1alpha1.Credentials{},
		// Application Registry puts a value from payload here
		RequestParametersSecretName: "",
	}

	return entry
}

func (c converter) eventApiToService(definition graphql.EventAPIDefinition) v1alpha1.Service {

	description := ""

	if definition.Description != nil {
		description = *definition.Description
	}

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
		Description: description,
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

func (c converter) eventAPIToServiceEntry(definition graphql.EventAPIDefinition) v1alpha1.Entry {
	entry := v1alpha1.Entry{
		Type: specEventsType,
		// TODO what is put here by the Application Registry
		AccessLabel: "",
		// Director returns BLOB here
		SpecificationUrl: "",
	}

	return entry
}

func NewConverter() Converter {
	return converter{}
}
