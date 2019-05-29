package externalapi

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

type Service struct {
	ID          string             `json:"id"`
	Provider    string             `json:"provider"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Identifier  string             `json:"identifier,omitempty"`
	Labels      *map[string]string `json:"labels,omitempty"`
}

type ServiceDetails struct {
	Provider         string             `json:"provider" valid:"required~Provider field cannot be empty."`
	Name             string             `json:"name" valid:"required~Name field cannot be empty."`
	Description      string             `json:"description" valid:"required~Description field cannot be empty."`
	ShortDescription string             `json:"shortDescription,omitempty"`
	Identifier       string             `json:"identifier,omitempty"`
	Labels           *map[string]string `json:"labels,omitempty"`
	Api              *API               `json:"api,omitempty"`
	Events           *Events            `json:"events,omitempty"`
	Documentation    *Documentation     `json:"documentation,omitempty"`
}

type CreateServiceResponse struct {
	ID string `json:"id"`
}

type API struct {
	TargetUrl        string               `json:"targetUrl" valid:"url,required~targetUrl field cannot be empty."`
	Credentials      *Credentials         `json:"credentials,omitempty"`
	Spec             json.RawMessage      `json:"spec,omitempty"`
	SpecificationUrl string               `json:"specificationUrl,omitempty"`
	ApiType          string               `json:"apiType,omitempty"`
	Headers          *map[string][]string `json:"headers,omitempty"`
	QueryParameters  *map[string][]string `json:"queryParameters,omitempty"`
}

type Credentials struct {
	Oauth           *Oauth               `json:"oauth,omitempty"`
	Basic           *BasicAuth           `json:"basic,omitempty"`
	CertificateGen  *CertificateGen      `json:"certificateGen,omitempty"`
	Headers         *map[string][]string `json:"headers,omitempty"`
	QueryParameters *map[string][]string `json:"queryParameters,omitempty"`
}

type CSRFInfo struct {
	TokenEndpointURL string `json:"tokenEndpointURL" valid:"url,required~tokenEndpointURL field cannot be empty"`
}

type Oauth struct {
	URL          string    `json:"url" valid:"url,required~oauth url field cannot be empty"`
	ClientID     string    `json:"clientId" valid:"required~oauth clientId field cannot be empty"`
	ClientSecret string    `json:"clientSecret" valid:"required~oauth clientSecret cannot be empty"`
	CSRFInfo     *CSRFInfo `json:"csrfInfo,omitempty"`
}

type BasicAuth struct {
	Username string    `json:"username" valid:"required~basic auth username field cannot be empty"`
	Password string    `json:"password" valid:"required~basic auth password field cannot be empty"`
	CSRFInfo *CSRFInfo `json:"csrfInfo,omitempty"`
}

type CertificateGen struct {
	CommonName  string    `json:"commonName"`
	Certificate string    `json:"certificate"`
	CSRFInfo    *CSRFInfo `json:"csrfInfo,omitempty"`
}

type Events struct {
	Spec json.RawMessage `json:"spec" valid:"required~spec cannot be empty"`
}

type Documentation struct {
	DisplayName string       `json:"displayName" valid:"required~displayName field cannot be empty in documentation"`
	Description string       `json:"description" valid:"required~description field cannot be empty in documentation"`
	Type        string       `json:"type" valid:"required~type field cannot be empty in documentation"`
	Tags        []string     `json:"tags,omitempty"`
	Docs        []DocsObject `json:"docs,omitempty"`
}

type DocsObject struct {
	Title  string `json:"title"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

const stars = "********"

func serviceDefinitionToService(serviceDefinition model.ServiceDefinition) Service {
	return Service{
		ID:          serviceDefinition.ID,
		Name:        serviceDefinition.Name,
		Provider:    serviceDefinition.Provider,
		Description: serviceDefinition.Description,
		Identifier:  serviceDefinition.Identifier,
		Labels:      serviceDefinition.Labels,
	}
}

func serviceDefinitionToServiceDetails(serviceDefinition model.ServiceDefinition) (ServiceDetails, apperrors.AppError) {
	serviceDetails := ServiceDetails{
		Provider:         serviceDefinition.Provider,
		Name:             serviceDefinition.Name,
		Description:      serviceDefinition.Description,
		ShortDescription: serviceDefinition.ShortDescription,
		Identifier:       serviceDefinition.Identifier,
	}

	if serviceDefinition.Labels != nil {
		serviceDetails.Labels = serviceDefinition.Labels
	}

	if serviceDefinition.Api != nil {
		serviceDetails.Api = &API{
			TargetUrl:        serviceDefinition.Api.TargetUrl,
			Spec:             serviceDefinition.Api.Spec,
			SpecificationUrl: serviceDefinition.Api.SpecificationUrl,
			ApiType:          serviceDefinition.Api.ApiType,
			Headers:          serviceDefinition.Api.Headers,
			QueryParameters:  serviceDefinition.Api.QueryParameters,
		}

		if serviceDefinition.Api.Credentials != nil {
			serviceDetails.Api.Credentials = serviceDefinitionCredentialsToServiceDetailsCredentials(serviceDefinition.Api.Credentials)
		}
	}

	if serviceDefinition.Events != nil {
		serviceDetails.Events = &Events{
			Spec: serviceDefinition.Events.Spec,
		}
	}

	if serviceDefinition.Documentation != nil {
		err := json.Unmarshal(serviceDefinition.Documentation, &serviceDetails.Documentation)
		if err != nil {
			return serviceDetails, apperrors.Internal("Failed to unmarshal documentation, %s", err.Error())
		}

	}

	return serviceDetails, nil
}

func serviceDefinitionCredentialsToServiceDetailsCredentials(credentials *model.Credentials) *Credentials {

	csrfInfoFromModel := func(model *model.CSRFInfo) *CSRFInfo {
		if model == nil {
			return nil
		}
		return &CSRFInfo{
			TokenEndpointURL: model.TokenEndpointURL,
		}
	}

	if credentials.Oauth != nil {
		return &Credentials{
			Oauth: &Oauth{
				ClientID:     stars,
				ClientSecret: stars,
				URL:          credentials.Oauth.URL,
				CSRFInfo:     csrfInfoFromModel(credentials.Oauth.CSRFInfo),
			},
			Headers:         credentials.Headers,
			QueryParameters: credentials.QueryParameters,
		}
	}

	if credentials.Basic != nil {
		return &Credentials{
			Basic: &BasicAuth{
				Username: stars,
				Password: stars,
				CSRFInfo: csrfInfoFromModel(credentials.Basic.CSRFInfo),
			},
			Headers:         credentials.Headers,
			QueryParameters: credentials.QueryParameters,
		}
	}

	if credentials.CertificateGen != nil {
		return &Credentials{
			CertificateGen: &CertificateGen{
				CommonName:  credentials.CertificateGen.CommonName,
				Certificate: credentials.CertificateGen.Certificate,
				CSRFInfo:    csrfInfoFromModel(credentials.CertificateGen.CSRFInfo),
			},
			Headers:         credentials.Headers,
			QueryParameters: credentials.QueryParameters,
		}
	}

	return nil
}

func serviceDetailsToServiceDefinition(serviceDetails ServiceDetails) (model.ServiceDefinition, apperrors.AppError) {
	serviceDefinition := model.ServiceDefinition{
		Provider:         serviceDetails.Provider,
		Name:             serviceDetails.Name,
		Description:      serviceDetails.Description,
		ShortDescription: serviceDetails.ShortDescription,
		Identifier:       serviceDetails.Identifier,
	}

	if serviceDetails.Labels != nil {
		serviceDefinition.Labels = serviceDetails.Labels
	}

	if serviceDetails.Api != nil {
		serviceDefinition.Api = &model.API{
			TargetUrl:        serviceDetails.Api.TargetUrl,
			SpecificationUrl: serviceDetails.Api.SpecificationUrl,
			ApiType:          serviceDetails.Api.ApiType,
			Headers:          serviceDetails.Api.Headers,
			QueryParameters:  serviceDetails.Api.QueryParameters,
		}
		if serviceDetails.Api.Credentials != nil {
			serviceDefinition.Api.Credentials = serviceDetailsCredentialsToServiceDefinitionCredentials(serviceDetails.Api.Credentials)
		}
		if serviceDetails.Api.Spec != nil {
			serviceDefinition.Api.Spec = compact(serviceDetails.Api.Spec)
		}
	}

	if serviceDetails.Events != nil && serviceDetails.Events.Spec != nil {
		serviceDefinition.Events = &model.Events{
			Spec: compact(serviceDetails.Events.Spec),
		}
	}

	if serviceDetails.Documentation != nil {
		marshalled, err := json.Marshal(&serviceDetails.Documentation)
		if err != nil {
			return serviceDefinition, apperrors.WrongInput("Failed to marshal documentation, %s", err.Error())
		}
		serviceDefinition.Documentation = marshalled
	}

	return serviceDefinition, nil
}

func serviceDetailsCredentialsToServiceDefinitionCredentials(credentials *Credentials) *model.Credentials {

	csrfInfoToModel := func(api *CSRFInfo) *model.CSRFInfo {
		if api == nil {
			return nil
		}
		return &model.CSRFInfo{
			TokenEndpointURL: api.TokenEndpointURL,
		}
	}

	if credentials.Oauth != nil {
		return &model.Credentials{
			Oauth: &model.Oauth{
				ClientID:     credentials.Oauth.ClientID,
				ClientSecret: credentials.Oauth.ClientSecret,
				URL:          credentials.Oauth.URL,
				CSRFInfo:     csrfInfoToModel(credentials.Oauth.CSRFInfo),
			},
			Headers:         credentials.Headers,
			QueryParameters: credentials.QueryParameters,
		}
	}

	if credentials.Basic != nil {
		return &model.Credentials{
			Basic: &model.Basic{
				Username: credentials.Basic.Username,
				Password: credentials.Basic.Password,
				CSRFInfo: csrfInfoToModel(credentials.Basic.CSRFInfo),
			},
			Headers:         credentials.Headers,
			QueryParameters: credentials.QueryParameters,
		}
	}

	if credentials.CertificateGen != nil {
		return &model.Credentials{
			CertificateGen: &model.CertificateGen{
				CommonName: credentials.CertificateGen.CommonName,
				CSRFInfo:   csrfInfoToModel(credentials.CertificateGen.CSRFInfo),
			},
			Headers:         credentials.Headers,
			QueryParameters: credentials.QueryParameters,
		}
	}

	return nil
}

func (api API) MarshalJSON() ([]byte, error) {
	bytes, err := api.marshalWithJSONSpec()
	if err == nil {
		return bytes, nil
	}

	bytes, err = api.marshalWithNonJSONSpec()
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (api API) marshalWithJSONSpec() ([]byte, error) {
	return json.Marshal(&struct {
		TargetUrl        string               `json:"targetUrl" valid:"url,required~targetUrl field cannot be empty."`
		Credentials      *Credentials         `json:"credentials,omitempty"`
		Spec             json.RawMessage      `json:"spec,omitempty"`
		SpecificationUrl string               `json:"specificationUrl,omitempty"`
		ApiType          string               `json:"apiType,omitempty"`
		Headers          *map[string][]string `json:"headers,omitempty"`
		QueryParameters  *map[string][]string `json:"queryParameters,omitempty"`
	}{
		api.TargetUrl,
		api.Credentials,
		api.Spec,
		api.SpecificationUrl,
		api.ApiType,
		api.Headers,
		api.QueryParameters,
	})
}

func (api API) marshalWithNonJSONSpec() ([]byte, error) {
	return json.Marshal(&struct {
		TargetUrl        string               `json:"targetUrl" valid:"url,required~targetUrl field cannot be empty."`
		Credentials      *Credentials         `json:"credentials,omitempty"`
		Spec             string               `json:"spec,omitempty"`
		SpecificationUrl string               `json:"specificationUrl,omitempty"`
		ApiType          string               `json:"apiType,omitempty"`
		Headers          *map[string][]string `json:"headers,omitempty"`
		QueryParameters  *map[string][]string `json:"queryParameters,omitempty"`
	}{
		api.TargetUrl,
		api.Credentials,
		string(api.Spec),
		api.SpecificationUrl,
		api.ApiType,
		api.Headers,
		api.QueryParameters,
	})
}
