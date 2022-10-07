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
	fmt.Print("-->vladimir, Here1")
	loginResp, err := f.doLogin()

	fmt.Print("-->vladimir, Here2")
	if err != nil {
		return "", err
	}
	fmt.Print("-->vladimir, Here3")
	consent := f.prepareConsent(loginResp)
	fmt.Print("-->vladimir, Here4")

	return f.sentConsentToGetToken(loginResp, consent)
}

func (f *OidcHydraTestFlow) sentConsentToGetToken(response *http.Response, consentForm url.Values) (string, error) {
	fmt.Print("-->vladimir, Here5.1")
	redirectUrl, err := url.Parse(f.config.ClientConfig.RedirectUri)
	if err != nil {
		return "", err
	}
	fmt.Print("-->vladimir, Here5.2")
	var token string
	f.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		fmt.Print("-->vladimir, Here5.3")
		if req.URL.Host == redirectUrl.Host {
			fmt.Print("-->vladimir, Here5.4")
			token = getToken(req)
			fmt.Printf("-->vladimir, GOT token: %s", token)
			return http.ErrUseLastResponse
		}
		fmt.Print("-->vladimir, Here5.5")
		return nil
	}
	fmt.Print("-->vladimir, Here5.6")
	fmt.Printf("-->vladimir, Here5, URL: %s", response.Request.URL.String())
	fmt.Printf("-->vladimir, Here5, consent: %T", consentForm)
	_, err = f.httpClient.PostForm(response.Request.URL.String(), consentForm)
	fmt.Print("-->vladimir, Here5.7")
	return token, err
}

func (f *OidcHydraTestFlow) doLogin() (*http.Response, error) {
	u, err := url.Parse(fmt.Sprintf(authUrlTemplate, f.config.HydraFqdn, f.config.ClientConfig.ID, f.config.ClientConfig.RedirectUri, state, nonce))
	if err != nil {
		return nil, err
	}
	fmt.Print("-->vladimir, Here6")
	resp, err := f.httpClient.Get(u.String())
	if err != nil {
		return nil, errors.Wrap(err, "while performing HTTP GET on auth endpoint")
	}
	fmt.Print("-->vladimir, Here7")
	loginForm := url.Values{}
	loginForm.Set("email", f.config.UserCredentials.Username)
	loginForm.Set("password", f.config.UserCredentials.Password)
	loginForm.Set("challenge", resp.Request.URL.Query().Get("login_challenge"))
	resp, err = f.httpClient.PostForm(resp.Request.URL.String(), loginForm)
	if err != nil {
		return nil, errors.Wrap(err, "while performing HTTP POST on login endpoint")
	}
	fmt.Print("-->vladimir, Here8")
	return resp, nil
}

func (f *OidcHydraTestFlow) prepareConsent(response *http.Response) url.Values {
	consentForm := url.Values{}
	consentForm.Set("challenge", response.Request.URL.Query().Get("consent_challenge"))
	consentForm.Set("grant_scope", "openid")
	consentForm.Set("submit", "Allow+access")
	return consentForm
}
