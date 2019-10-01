package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
)

const (
	ContentTypeHeader = "Content-Type"

	GrantTypeFieldName   = "grant_type"
	CredentialsGrantType = "client_credentials"

	ScopeFieldName = "scope"
	Scopes         = "application:read application:write runtime:read runtime:write label_definition:read label_definition:write health_checks:read"
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

func (c *Client) GetAccessToken() (string, error) {
	clientCredentials, err := c.createOAuth2Client()

	if err != nil {
		return "", err
	}

	return c.getOAuthToken(clientCredentials)
}

func (c *Client) createOAuth2Client() (credentials, error) {
	body := oauthClient{
		GrantTypes: []string{CredentialsGrantType},
		Scope:      Scopes,
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

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return credentials{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return credentials{}, fmt.Errorf("create OAuth2 client call returned unexpected status code, %d", resp.StatusCode)
	}

	var oauthResp oauthClient
	err = json.NewDecoder(resp.Body).Decode(&oauthResp)

	if err != nil {
		return credentials{}, err
	}

	return credentials{
		clientID:     oauthResp.ClientID,
		clientSecret: oauthResp.Secret,
	}, nil
}

func (c *Client) getOAuthToken(credentials credentials) (string, error) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)

	err := setRequiredFields(writer)

	if err != nil {
		return "", err
	}

	request, err := http.NewRequest(http.MethodPost, c.tokensEndpoint, buffer)

	if err != nil {
		return "", err
	}

	request.SetBasicAuth(credentials.clientID, credentials.clientSecret)

	request.Header.Set(ContentTypeHeader, writer.FormDataContentType())

	response, err := c.httpClient.Do(request)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	var tokenResponse tokenResponse

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get token call returned unexpected status code, %d", response.StatusCode)
	}

	err = json.NewDecoder(response.Body).Decode(&tokenResponse)

	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func setRequiredFields(w *multipart.Writer) error {
	defer w.Close()

	err := w.WriteField(GrantTypeFieldName, CredentialsGrantType)
	if err != nil {
		return err
	}
	err = w.WriteField(ScopeFieldName, Scopes)
	if err != nil {
		return err
	}
	return nil
}
