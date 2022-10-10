package jwt

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const (
	state           = "dd3557bfb07ee1858f0ac8abc4a46aef"
	nonce           = "dd3557bfb07ee1858f0ac8abc4a46aef"
	authUrlTemplate = "https://%s/oauth2/auth?client_id=%s&response_type=id_token&scope=openid&redirect_uri=%s&state=%s&nonce=%s"
)

type idTokenProvider interface {
	fetchIdToken() (string, error)
}

// OidcHydraTestFlow implements OIDC flows for deployed OryHydra with test login consent endpoints.
type OidcHydraTestFlow struct {
	httpClient *http.Client
	config     oidcHydraConfig
}

func newOidcHydraTestFlow(httpClient *http.Client, config oidcHydraConfig) idTokenProvider {
	return &OidcHydraTestFlow{httpClient, config}
}

func getToken(req *http.Request) string {
	tokenWithTrailingState := strings.Split(req.URL.Fragment, "=")[1]
	t := strings.Split(tokenWithTrailingState, "&")[0]
	return t
}

func (f *OidcHydraTestFlow) fetchIdToken() (string, error) {
	return f.GetToken()
}

func (f *OidcHydraTestFlow) GetToken() (string, error) {
	loginResp, err := f.doLogin()

	if err != nil {
		return "", err
	}
	consent := f.prepareConsent(loginResp)

	return f.sentConsentToGetToken(loginResp, consent)
}

func (f *OidcHydraTestFlow) sentConsentToGetToken(response *http.Response, consentForm url.Values) (string, error) {
	redirectUrl, err := url.Parse(f.config.ClientConfig.RedirectUri)
	if err != nil {
		return "", err
	}
	var token string
	f.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req.URL.Host == redirectUrl.Host {
			token = getToken(req)
			return http.ErrUseLastResponse
		}
		return nil
	}
	resp, err := f.httpClient.PostForm(response.Request.URL.String(), consentForm)
	if resp.StatusCode>399 {
		return "", fmt.Errorf("could not fetch token, err_code=%d", resp.StatusCode)
	}
	return token, err
}

func (f *OidcHydraTestFlow) doLogin() (*http.Response, error) {
	u, err := url.Parse(fmt.Sprintf(authUrlTemplate, f.config.HydraFqdn, f.config.ClientConfig.ID, f.config.ClientConfig.RedirectUri, state, nonce))
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Get(u.String())
	if err != nil {
		return nil, errors.Wrap(err, "while performing HTTP GET on auth endpoint")
	}

	loginForm := url.Values{}
	loginForm.Set("email", f.config.UserCredentials.Username)
	loginForm.Set("password", f.config.UserCredentials.Password)
	loginForm.Set("challenge", resp.Request.URL.Query().Get("login_challenge"))
	resp, err = f.httpClient.PostForm(resp.Request.URL.String(), loginForm)
	if err != nil {
		return nil, errors.Wrap(err, "while performing HTTP POST on login endpoint")
	}

	if resp.StatusCode>399 {
		return nil, fmt.Errorf("could not do login, err_code=%d", resp.StatusCode)
	}
	
	return resp, nil
}

func (f *OidcHydraTestFlow) prepareConsent(response *http.Response) url.Values {
	consentForm := url.Values{}
	consentForm.Set("challenge", response.Request.URL.Query().Get("consent_challenge"))
	consentForm.Set("grant_scope", "openid")
	consentForm.Set("submit", "Allow+access")
	return consentForm
}
