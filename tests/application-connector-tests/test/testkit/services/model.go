package services

import "encoding/json"

type PublishRequest struct {
	EventType        string   `json:"event-type,omitempty"`
	EventTypeVersion string   `json:"event-type-version,omitempty"`
	EventID          string   `json:"event-id,omitempty"`
	EventTime        string   `json:"event-time,omitempty"`
	Data             AnyValue `json:"data,omitempty"`
}

type PublishResponse struct {
	EventID string `json:"event-id,omitempty"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

type AnyValue interface {
}

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
	TargetUrl        string               `json:"targetUrl"`
	Credentials      *Credentials         `json:"credentials,omitempty"`
	Spec             json.RawMessage      `json:"spec,omitempty"`
	SpecificationUrl string               `json:"specificationUrl,omitempty"`
	ApiType          string               `json:"apiType"`
	Headers          *map[string][]string `json:"headers,omitempty"`
	QueryParameters  *map[string][]string `json:"queryParameters,omitempty"`
}

type Credentials struct {
	Oauth           *Oauth               `json:"oauth,omitempty"`
	Basic           *Basic               `json:"basic,omitempty"`
	CertificateGen  *CertificateGen      `json:"certificateGen,omitempty"`
	Headers         *map[string][]string `json:"headers,omitempty"`
	QueryParameters *map[string][]string `json:"queryParameters,omitempty"`
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

func (api *API) WithBasicAuth(username, password string) *API {
	if api.Credentials == nil {
		api.Credentials = &Credentials{}
	}

	api.Credentials.Basic = &Basic{
		Username: username,
		Password: password,
	}

	return api
}

func (api *API) WithOAuth(url, clientID, clientSecret string) *API {
	if api.Credentials == nil {
		api.Credentials = &Credentials{}
	}

	api.Credentials.Oauth = &Oauth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		URL:          url,
	}

	return api
}

func (api *API) WithCustomHeaders(headers *map[string][]string) *API {
	api.Headers = headers

	return api
}

func (api *API) WithCustomQueryParams(queryParams *map[string][]string) *API {
	api.QueryParameters = queryParams

	return api
}
