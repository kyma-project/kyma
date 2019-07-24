package registry

import "encoding/json"

type Service struct {
	ID               string            `json:"id"`
	Provider         string            `json:"provider"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	ShortDescription string            `json:"shortDescription,omitempty"`
	Identifier       string            `json:"identifier,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
}

type ServiceDetails struct {
	Provider         string            `json:"provider"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	ShortDescription string            `json:"shortDescription,omitempty"`
	Identifier       string            `json:"identifier,omitempty"`
	Labels           map[string]string `json:"labels,omitempty"`
	Api              *API              `json:"api,omitempty"`
	Events           *Events           `json:"events,omitempty"`
	Documentation    *Documentation    `json:"documentation,omitempty"`
}

type PostServiceResponse struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type API struct {
	TargetUrl                      string               `json:"targetUrl"`
	Credentials                    *CredentialsWithCSRF `json:"credentials,omitempty"`
	Spec                           json.RawMessage      `json:"spec,omitempty"`
	SpecificationUrl               string               `json:"specificationUrl,omitempty"`
	ApiType                        string               `json:"apiType"`
	RequestParameters              *RequestParameters   `json:"requestParameters"`
	SpecificationCredentials       *Credentials         `json:"specificationCredentials"`
	SpecificationRequestParameters *RequestParameters   `json:"specificationRequestParameters"`
}

type RequestParameters struct {
	Headers         *map[string][]string `json:"headers,omitempty"`
	QueryParameters *map[string][]string `json:"queryParameters,omitempty"`
}

type CredentialsWithCSRF struct {
	Oauth          *Oauth          `json:"oauth,omitempty"`
	Basic          *Basic          `json:"basic,omitempty"`
	CertificateGen *CertificateGen `json:"certificateGen,omitempty"`
}

type Credentials struct {
	Oauth *Oauth `json:"oauth,omitempty"`
	Basic *Basic `json:"basic,omitempty"`
}

type Oauth struct {
	URL          string `json:"url"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type Basic struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CertificateGen struct {
	CommonName string `json:"commonName"`
}

type Events struct {
	Spec json.RawMessage `json:"spec,omitempty"`
}

type Documentation struct {
	DisplayName string       `json:"displayName"`
	Description string       `json:"description"`
	Type        string       `json:"type"`
	Tags        []string     `json:"tags,omitempty"`
	Docs        []DocsObject `json:"docs,omitempty"`
}

type DocsObject struct {
	Title  string `json:"title"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

func (api *API) WithAPISpecURL(specURL string) *API {
	api.SpecificationUrl = specURL

	return api
}

func (api *API) WithBasicAuth(username, password string) *API {
	if api.Credentials == nil {
		api.Credentials = &CredentialsWithCSRF{}
	}

	api.Credentials.Basic = &Basic{
		Username: username,
		Password: password,
	}

	return api
}

func (api *API) WithOAuth(url, clientID, clientSecret string) *API {
	if api.Credentials == nil {
		api.Credentials = &CredentialsWithCSRF{}
	}

	api.Credentials.Oauth = &Oauth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		URL:          url,
	}

	return api
}

func (api *API) WithCustomHeaders(headers *map[string][]string) *API {
	api.RequestParameters = &RequestParameters{
		Headers: headers,
	}

	return api
}

func (api *API) WithOAuthSecuredSpec(oauthURL, clientID, clientSecret string) *API {
	if api.SpecificationCredentials == nil {
		api.SpecificationCredentials = &Credentials{}
	}

	api.SpecificationCredentials.Oauth = &Oauth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		URL:          oauthURL,
	}

	return api
}

func (api *API) WithCustomHeadersSpec(headers *map[string][]string) *API {
	api.SpecificationRequestParameters = &RequestParameters{
		Headers: headers,
	}

	return api
}

func (api *API) WithCustomQueryParams(queryParams *map[string][]string) *API {
	api.RequestParameters = &RequestParameters{
		QueryParameters: queryParams,
	}

	return api
}

func (api *API) WithCustomQueryParamsSpec(queryParams *map[string][]string) *API {
	api.SpecificationRequestParameters = &RequestParameters{
		QueryParameters: queryParams,
	}

	return api
}

func (api *API) WithBasicAuthSecuredSpec(username, password string) *API {
	if api.SpecificationCredentials == nil {
		api.SpecificationCredentials = &Credentials{}
	}

	api.SpecificationCredentials.Basic = &Basic{
		Username: username,
		Password: password,
	}

	return api
}
