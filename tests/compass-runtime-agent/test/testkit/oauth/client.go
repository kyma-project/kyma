package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
)

type Client struct {
	HydraURL       string
	TokensEndpoint string
	HttpClient     http.Client
}

type oauthClient struct {
	ClientID      string   `json:"client_id,omitempty"`
	Secret        string   `json:"client_secret,omitempty"`
	GrantTypes    []string `json:"grant_types"`
	ResponseTypes []string `json:"response_types,omitempty"`
	Scope         string   `json:"scope"`
	Owner         string   `json:"owner"`
}

func (c *Client) CreateOAuth2Client() (string error) {
	grantTypes := []string{"client_credentials"}
	body := oauthClient{
		GrantTypes: grantTypes,
		Scope:      "read write",
	}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, c.HydraURL, buf)

	var oauthResp oauthClient

	resp, err := c.HttpClient.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		err = json.NewDecoder(resp.Body).Decode(oauthResp)
	}

	clientID := oauthResp.ClientID
	clientSecret := oauthResp.ClientID

	buildCredentialsString(clientID, clientSecret)

	return err
}

func buildCredentialsString(clientID string, clientSecret string) (string, error) {
	clientIDBytes, e := base64.StdEncoding.DecodeString(clientID)

	if e != nil {
		return "", e
	}

	secretValueBytes, e := base64.StdEncoding.DecodeString(clientSecret)

	if e != nil {
		return "", e
	}

	clientID = string(clientIDBytes)
	clientSecret = string(secretValueBytes)

	credentials := []byte(fmt.Sprintf("Basic %s:%s", clientID, clientSecret))
	return base64.StdEncoding.EncodeToString(credentials), nil
}

func (c *Client) getOAuthToken(credentials string) (string, error) {
	request, e := http.NewRequest(http.MethodPost, c.TokensEndpoint, nil)

	if e != nil {
		return "", e
	}

	request.MultipartForm = &multipart.Form{Value: map[string][]string{"grant_type": {"client_credentials"}, "scope": {"scope-a scope-b"}}}
	request.Header.Set("Authorization", credentials)

	response, e := c.HttpClient.Do(request)

	if e != nil {
		return "", e
	}

}
