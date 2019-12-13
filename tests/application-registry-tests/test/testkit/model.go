package testkit

import (
	"bytes"
	"encoding/json"
)

var (
	ApiRawSpec    = Compact([]byte("{\"name\":\"api\"}"))
	EventsRawSpec = Compact([]byte("{\"asyncapi\":\"2.0.0\",\"info\":{\"title\":\"OneOf example\",\"version\":\"1.0.0\"},\"channels\":{\"test\":{\"publish\":{\"message\":{\"$ref\":\"#/components/messages/testMessages\"}}}},\"components\":{\"messages\":{\"testMessages\":{\"description\":\"test\"}}}}"))

	SwaggerApiSpec = Compact([]byte("{\"swagger\":\"2.0\"}"))
)

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
	TargetUrl         string               `json:"targetUrl"`
	Credentials       *Credentials         `json:"credentials,omitempty"`
	Spec              json.RawMessage      `json:"spec,omitempty"`
	SpecificationUrl  string               `json:"specificationUrl,omitempty"`
	ApiType           string               `json:"apiType"`
	RequestParameters *RequestParameters   `json:"requestParameters,omitempty"`
	Headers           *map[string][]string `json:"headers,omitempty"`
	QueryParameters   *map[string][]string `json:"queryParameters,omitempty"`
}

type RequestParameters struct {
	Headers         *map[string][]string `json:"headers,omitempty"`
	QueryParameters *map[string][]string `json:"queryParameters,omitempty"`
}

type Credentials struct {
	Oauth          *Oauth          `json:"oauth,omitempty"`
	Basic          *Basic          `json:"basic,omitempty"`
	CertificateGen *CertificateGen `json:"certificateGen,omitempty"`
}

type CSRFInfo struct {
	TokenEndpointURL string `json:"tokenEndpointURL"`
}

type Oauth struct {
	URL               string             `json:"url"`
	ClientID          string             `json:"clientId"`
	ClientSecret      string             `json:"clientSecret"`
	CSRFInfo          *CSRFInfo          `json:"csrfInfo,omitempty"`
	RequestParameters *RequestParameters `json:"requestParameters,omitempty"`
}

type Basic struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	CSRFInfo *CSRFInfo `json:"csrfInfo,omitempty"`
}

type CertificateGen struct {
	CommonName  string    `json:"commonName"`
	Certificate string    `json:"certificate"`
	CSRFInfo    *CSRFInfo `json:"csrfInfo,omitempty"`
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

func Compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}

func (sd ServiceDetails) WithAPI(api *API) ServiceDetails {
	sd.Api = api
	return sd
}

func (api *API) WithCSRFInOAuth(csrfInfo *CSRFInfo) *API {
	if api.Credentials != nil && api.Credentials.Oauth != nil {
		api.Credentials.Oauth.CSRFInfo = csrfInfo
	}
	return api
}

func (api *API) WithRequestParametersInOAuth(requestParameters *RequestParameters) *API {
	if api.Credentials != nil && api.Credentials.Oauth != nil {
		api.Credentials.Oauth.RequestParameters = requestParameters
	}
	return api
}

func (api *API) WithCSRFInBasic(csrfInfo *CSRFInfo) *API {
	if api.Credentials != nil && api.Credentials.Basic != nil {
		api.Credentials.Basic.CSRFInfo = csrfInfo
	}
	return api
}

func (api *API) WithCSRFInCertificateGen(csrfInfo *CSRFInfo) *API {
	if api.Credentials != nil && api.Credentials.CertificateGen != nil {
		api.Credentials.CertificateGen.CSRFInfo = csrfInfo
	}
	return api
}

func (api *API) WithRequestParameters(requestParameters *RequestParameters) *API {
	api.RequestParameters = requestParameters
	return api
}

func (api *API) WithHeadersAndQueryParameters(headers, queryParameters *map[string][]string) *API {
	api.Headers = headers
	api.QueryParameters = queryParameters
	return api
}
