package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
)

const (
	AuthorizationHeader = "Authorization"
	ContentTypeHeader   = "Content-Type"

	GrantTypeFieldName   = "grant_type"
	CredentialsGrantType = "client_credentials"

	ScopeFieldName = "scope"
	RWScope        = "read write"
)

type Client struct {
	clientsEndpoint string
	tokensEndpoint  string
	httpClient      http.Client
}

type oauthClient struct {
	ClientID      string   `json:"client_id,omitempty"`
	Secret        string   `json:"client_secret,omitempty"`
	GrantTypes    []string `json:"grant_types"`
	ResponseTypes []string `json:"response_types,omitempty"`
	Scope         string   `json:"scope"`
	Owner         string   `json:"owner"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Expiration  int    `json:"expires_in"`
}

type credentials struct {
	clientID     string
	clientSecret string
}

func NewOauthClient(hydraPublicURL, hydraAdminURL string) *Client {
	return &Client{
		clientsEndpoint: fmt.Sprintf("%s/clients", hydraAdminURL),
		tokensEndpoint:  fmt.Sprintf("%s/oauth2/token", hydraPublicURL),
		httpClient:      http.Client{},
	}
}

func (c *Client) GetAuthorizationToken() (string, error) {
	clientCredentials, e := c.createOAuth2Client()

	if e != nil {
		return "", e
	}

	credentials, e := buildCredentialsString(clientCredentials)

	if e != nil {
		return "", e
	}

	return c.getOAuthToken(credentials)
}

func (c *Client) createOAuth2Client() (credentials, error) {
	body := oauthClient{
		GrantTypes: []string{CredentialsGrantType},
		Scope:      RWScope,
	}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return credentials{}, err
	}

	request, err := http.NewRequest(http.MethodPost, c.clientsEndpoint, buf)

	if err != nil {
		return credentials{}, err
	}

	var oauthResp oauthClient

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return credentials{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return credentials{}, fmt.Errorf("create OAuth2 client call returned unexpected status code, %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&oauthResp)

	if err != nil {
		return credentials{}, err
	}

	return credentials{
		clientID:     oauthResp.ClientID,
		clientSecret: oauthResp.Secret,
	}, nil
}

func buildCredentialsString(credentials credentials) (string, error) {
	credentialsEncoded := []byte(fmt.Sprintf("%s:%s", credentials.clientID, credentials.clientSecret))
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(credentialsEncoded)), nil
}

func (c *Client) getOAuthToken(credentials string) (string, error) {
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)

	e := setRequiredFields(w)

	if e != nil {
		return "", e
	}

	request, e := http.NewRequest(http.MethodPost, c.tokensEndpoint, b)

	if e != nil {
		return "", e
	}

	request.Header.Set(AuthorizationHeader, credentials)
	request.Header.Set(ContentTypeHeader, w.FormDataContentType())

	response, e := c.httpClient.Do(request)

	if e != nil {
		return "", e
	}

	defer response.Body.Close()

	var tokenResponse tokenResponse

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get token call returned unexpected status code, %d", response.StatusCode)
	}

	e = json.NewDecoder(response.Body).Decode(&tokenResponse)

	if e != nil {
		return "", e
	}

	return tokenResponse.AccessToken, nil
}

func setRequiredFields(w *multipart.Writer) error {
	defer w.Close()

	err := w.WriteField(GrantTypeFieldName, CredentialsGrantType)
	if err != nil {
		return err
	}
	err = w.WriteField(ScopeFieldName, RWScope)
	if err != nil {
		return err
	}
	return nil
}
