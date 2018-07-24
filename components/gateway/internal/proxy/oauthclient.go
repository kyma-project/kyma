package proxy

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/gateway/internal/httpconsts"
)

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type OAuthClient interface {
	GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
}

type oauthClient struct {
	timeoutDuration int
}

func NewOauthClient(timeoutDuration int) OAuthClient {
	return &oauthClient{timeoutDuration: timeoutDuration}
}

func (oc *oauthClient) GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	form := url.Values{}
	form.Add("client_id", clientID)
	form.Add("client_secret", clientSecret)
	form.Add("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, authURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", apperrors.Internal("failed to create token request: %s", err.Error())
	}

	req.Header.Add(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationURLEncoded)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(oc.timeoutDuration)*time.Second)
	defer cancel()
	requestWithContext := req.WithContext(ctx)

	response, err := client.Do(requestWithContext)
	if err != nil {
		return "", apperrors.UpstreamServerCallFailed("failed to make a request to '%s': %s", authURL, err.Error())
	}

	if response.StatusCode != http.StatusOK {
		return "", apperrors.UpstreamServerCallFailed("incorrect response code '%s' while getting token from %s", response.StatusCode, authURL)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return "", apperrors.UpstreamServerCallFailed("failed to read token response body from '%s': %s", authURL, err.Error())
	}

	tokenResponse := oauthResponse{}

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", apperrors.UpstreamServerCallFailed("failed to unmarshal token response body: %s", err.Error())
	}

	return "Bearer " + tokenResponse.AccessToken, nil
}
