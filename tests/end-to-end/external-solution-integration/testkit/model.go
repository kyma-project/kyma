package testkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var (
	ApiRawSpec = compact([]byte("{\"name\":\"api\"}"))
)

type Subject struct {
	CommonName         string
	Country            string
	Organization       string
	OrganizationalUnit string
	Locality           string
	Province           string
}

type InfoResponse struct {
	CertUrl     string   `json:"csrUrl"`
	Api         ApiInfo  `json:"api"`
	Certificate CertInfo `json:"certificate"`
}

type RegisterServiceResponse struct {
	ID string `json:"id"`
}

type RuntimeURLs struct {
	MetadataUrl string `json:"metadataUrl"`
	EventsUrl   string `json:"eventsUrl"`
}

type ApiInfo struct {
	*RuntimeURLs
	ManagementInfoURL string `json:"infoUrl"`
	CertificatesUrl   string `json:"certificatesUrl"`
}

type CertInfo struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}

type CsrRequest struct {
	Csr string `json:"csr"`
}

type CrtResponse struct {
	CRTChain  string `json:"crt"`
	ClientCRT string `json:"clientCrt"`
	CaCRT     string `json:"caCrt"`
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

type API struct {
	TargetUrl        string          `json:"targetUrl"`
	Credentials      *Credentials    `json:"credentials,omitempty"`
	Spec             json.RawMessage `json:"spec,omitempty"`
	SpecificationUrl string          `json:"specificationUrl,omitempty"`
	ApiType          string          `json:"apiType"`
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
	URL          string    `json:"url"`
	ClientID     string    `json:"clientId"`
	ClientSecret string    `json:"clientSecret"`
	CSRFInfo     *CSRFInfo `json:"csrfInfo,omitempty"`
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

type ErrorResponse struct {
	Code     int    `json:"code"`
	ErrorMsg string `json:"error"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%v: %s", e.Code, e.ErrorMsg)
}

type ExampleEvent struct {
	EventType        string    `json:"event-type"`
	EventTypeVersion string    `json:"event-type-version"`
	EventID          string    `json:"event-id"`
	EventTime        time.Time `json:"event-time"`
	Data             string    `json:"data"`
}

func parseErrorResponse(response *http.Response) error {
	errorResponse := &ErrorResponse{}
	err := json.NewDecoder(response.Body).Decode(errorResponse)
	if err != nil {
		return err
	}

	return errorResponse
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
