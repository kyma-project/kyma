package oauth

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

//go:generate mockery -name=Client
type Client interface {
	GetAuthorizationToken() (Token, error)
}

type oauthClient struct {
	httpClient       *http.Client
	oauthCredentials credentials
}

func NewOauthClient(client *http.Client, creds credentials) Client {
	return &oauthClient{
		httpClient:       client,
		oauthCredentials: creds,
	}
}

func (c *oauthClient) GetAuthorizationToken() (Token, error) {
	return c.getAuthorizationToken(c.oauthCredentials)
}

func (c *oauthClient) getAuthorizationToken(credentials credentials) (Token, error) {
	log.Infof("Getting authorisation token for credentials to access Director from endpoint: %s", credentials.tokensEndpoint)

	form := url.Values{}
	form.Add(grantTypeFieldName, credentialsGrantType)
	form.Add(scopeFieldName, scopes)

	request, err := http.NewRequest(http.MethodPost, credentials.tokensEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		log.Errorf("Failed to create authorisation token request")
		return Token{}, errors.Wrap(err, "Failed to create authorisation token request")
	}

	now := time.Now().Unix()

	request.SetBasicAuth(credentials.clientID, credentials.clientSecret)
	request.Header.Set(contentTypeHeader, contentTypeApplicationURLEncoded)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Token{}, errors.Wrap(err, "Failed to execute http call")
	}
	//defer util.Close(response.Body)

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		dump, err := httputil.DumpResponse(response, true)
		if err != nil {
			dump = []byte("failed to dump response body")
		}
		return Token{}, fmt.Errorf("Get token call returned unexpected status: %s. Response dump: %s", response.Status, string(dump))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return Token{}, fmt.Errorf("Failed to read token response body from '%s': %s", err.Error())
	}

	tokenResponse := Token{}

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return Token{}, fmt.Errorf("failed to unmarshal token response body: %s", err.Error())
	}

	log.Infof("Successfully unmarshal response oauth token for accessing Director")

	tokenResponse.Expiration += now

	return tokenResponse, nil
}
