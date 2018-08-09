package dex

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"log"
)

type idTokenProvider interface {
	fetchIdToken() (string, error)
}

type dexIdTokenProvider struct {
	httpClient *http.Client
	config     idProviderConfig
}

func newDexIdTokenProvider(httpClient *http.Client, config idProviderConfig) idTokenProvider {
	return &dexIdTokenProvider{httpClient, config}
}

func (p *dexIdTokenProvider) fetchIdToken() (string, error) {

	flowResult, err := p.implicitFlow()
	if err != nil {
		return "", err
	}
	return flowResult["id_token"], nil
}

func (p *dexIdTokenProvider) implicitFlow() (map[string]string, error) {

	authorizeResp, err := p.httpClient.PostForm(p.config.dexConfig.authorizeEndpoint, url.Values{
		"response_type": {"id_token token"},
		"client_id":     {p.config.clientConfig.id},
		"redirect_uri":  {p.config.clientConfig.redirectUri},
		"scope":         {"openid profile email groups"},
		"nonce":         {"vF7FAQlqq41CObeUFYY0ggv1qEELvfHaXQ0ER4XM"},
	})
	if err != nil {
		return nil, err
	}
	if authorizeResp.StatusCode < 300 || authorizeResp.StatusCode > 399 {
		return nil, fmt.Errorf("Authorize - response error: '%s' - %s", authorizeResp.Status, readRespBody(authorizeResp))
	}

	// /auth/local?req=qruhpy2cqjvv4hcrbuu44mf4v
	loginEndpoint := authorizeResp.Header.Get("location")
	if strings.Contains(loginEndpoint, "#.*error") {
		return nil, fmt.Errorf("Login - Redirected with error: '%s'", loginEndpoint)
	}

	_, err = p.httpClient.Get(p.config.dexConfig.baseUrl + loginEndpoint)
	if err != nil {
		return nil, err
	}

	loginResp, err := p.httpClient.PostForm(p.config.dexConfig.baseUrl+loginEndpoint, url.Values{
		"login":    {p.config.userCredentials.username},
		"password": {p.config.userCredentials.password},
	})
	if err != nil {
		return nil, err
	}
	if loginResp.StatusCode < 300 || loginResp.StatusCode > 399 {
		return nil, fmt.Errorf("Login - response error: '%s' - %s", loginResp.Status, readRespBody(loginResp))
	}

	// /approval?req=qruhpy2cqjvv4hcrbuu44mf4v
	approvalEndpoint := loginResp.Header.Get("location")
	if strings.Contains(approvalEndpoint, "#.*error") {
		return nil, fmt.Errorf("Approval - Redirected with error: '%s'", approvalEndpoint)
	}
	approvalResp, err := p.httpClient.Get(p.config.dexConfig.baseUrl + approvalEndpoint)
	if err != nil {
		return nil, err
	}
	if approvalResp.StatusCode < 300 || approvalResp.StatusCode > 399 {
		return nil, fmt.Errorf("Approval - response error: '%s' - %s", approvalResp.Status, readRespBody(approvalResp))
	}

	clientEndpoint := approvalResp.Header.Get("location")
	if strings.Contains(clientEndpoint, "#.*error") {
		return nil, fmt.Errorf("Client - Redirected with error: '%s'", clientEndpoint)
	}

	parsedUrl, parseErr := url.Parse(clientEndpoint)
	if parseErr != nil {
		return nil, parseErr
	}

	result := make(map[string]string)
	fragmentParams := strings.Split(parsedUrl.Fragment, "&")
	for _, param := range fragmentParams {
		keyAndValue := strings.Split(param, "=")
		result[keyAndValue[0]] = keyAndValue[1]
	}

	return result, nil
}

func readRespBody(resp *http.Response) string {
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("WARNING: Unable to read response body (status: '%s'). Root cause: %v", resp.Status, err)
		return "<<Error reading response body>>"
	}
	return string(b)
}
