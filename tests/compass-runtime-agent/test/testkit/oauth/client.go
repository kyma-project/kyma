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
	AuthorizationHeader  = "Authorization"
	CredentialsGrantType = "client_credentials"
	RWScope              = "read write"
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

type credentials struct {
	clientID     string
	clientSecret string
}

func NewOauthClient(hydraURL string) *Client {
	return &Client{
		clientsEndpoint: fmt.Sprintf("%s/clients", hydraURL),
		tokensEndpoint:  fmt.Sprintf("%s/oauth2/tokens", hydraURL),
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

	buf := new(bytes.Buffer)
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

	if resp.StatusCode == http.StatusCreated {
		err = json.NewDecoder(resp.Body).Decode(oauthResp)
	}

	if err != nil {
		return credentials{}, err
	}

	return credentials{
		clientID:     oauthResp.ClientID,
		clientSecret: oauthResp.Secret,
	}, nil
}

func buildCredentialsString(credentials credentials) (string, error) {
	clientIDBytes, e := base64.StdEncoding.DecodeString(credentials.clientID)

	if e != nil {
		return "", e
	}

	secretValueBytes, e := base64.StdEncoding.DecodeString(credentials.clientSecret)

	if e != nil {
		return "", e
	}

	clientID := string(clientIDBytes)
	clientSecret := string(secretValueBytes)

	credentialsEncoded := []byte(fmt.Sprintf("Basic %s:%s", clientID, clientSecret))
	return base64.StdEncoding.EncodeToString(credentialsEncoded), nil
}

func (c *Client) getOAuthToken(credentials string) (string, error) {
	request, e := http.NewRequest(http.MethodPost, c.tokensEndpoint, nil)

	if e != nil {
		return "", e
	}

	request.MultipartForm = &multipart.Form{Value: map[string][]string{"grant_type": {CredentialsGrantType}, "scope": {RWScope}}}
	request.Header.Set(AuthorizationHeader, credentials)

	response, e := c.httpClient.Do(request)

	if e != nil {
		return "", e
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusCreated {
		e = json.NewDecoder(response.Body).Decode()
	}

	if e != nil {
		return "", e
	}

	return "", nil
}
