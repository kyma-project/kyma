/*
 * [y] hybris Platform
 *
 * Copyright (c) 2000-2018 hybris AG
 * All rights reserved.
 *
 * This software is the confidential and proprietary information of hybris
 * ("Confidential Information"). You shall not disclose such Confidential
 * Information and shall use it only in accordance with the terms of the
 * license agreement you entered into with hybris.
 */
package oauth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authentication/oauth/tokencache"
)

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type OAuthClient interface {
	GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
	InvalidateAndRetry(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
	InvalidateTokenCache(clientID string)
}

type oauthClient struct {
	timeoutDuration int
	tokenCache      tokencache.TokenCache
}

func NewOauthClient(timeoutDuration int, tokenCache tokencache.TokenCache) OAuthClient {
	return &oauthClient{
		timeoutDuration: timeoutDuration,
		tokenCache:      tokenCache,
	}
}

func (oc *oauthClient) GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError) {
	token, found := oc.tokenCache.Get(clientID)
	if found {
		return "Bearer " + token, nil
	}

	tokenResponse, err := oc.requestToken(clientID, clientSecret, authURL)
	if err != nil {
		return "", err
	}

	oc.tokenCache.Add(clientID, tokenResponse.AccessToken, tokenResponse.ExpiresIn)

	return "Bearer " + tokenResponse.AccessToken, nil
}

func (oc *oauthClient) InvalidateAndRetry(clientID string, clientSecret string, authURL string) (string, apperrors.AppError) {
	oc.tokenCache.Remove(clientID)

	tokenResponse, err := oc.requestToken(clientID, clientSecret, authURL)
	if err != nil {
		return "", err
	}

	oc.tokenCache.Add(clientID, tokenResponse.AccessToken, tokenResponse.ExpiresIn)

	return "Bearer " + tokenResponse.AccessToken, nil
}

func (oc *oauthClient) InvalidateTokenCache(clientID string) {
	oc.tokenCache.Remove(clientID)
}

func (oc *oauthClient) requestToken(clientID string, clientSecret string, authURL string) (*oauthResponse, apperrors.AppError) {
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
		return nil, apperrors.Internal("failed to create token request: %s", err.Error())
	}

	req.Header.Add(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationURLEncoded)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(oc.timeoutDuration)*time.Second)
	defer cancel()
	requestWithContext := req.WithContext(ctx)

	response, err := client.Do(requestWithContext)
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("failed to make a request to '%s': %s", authURL, err.Error())
	}

	if response.StatusCode != http.StatusOK {
		return nil, apperrors.UpstreamServerCallFailed("incorrect response code '%d' while getting token from %s", response.StatusCode, authURL)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("failed to read token response body from '%s': %s", authURL, err.Error())
	}

	tokenResponse := &oauthResponse{}

	err = json.Unmarshal(body, tokenResponse)
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("failed to unmarshal token response body: %s", err.Error())
	}

	return tokenResponse, nil
}
