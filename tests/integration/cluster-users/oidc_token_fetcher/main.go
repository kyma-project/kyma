package main

import (
	"flag"
	"fmt"
	l "log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

const (
	state           = "dd3557bfb07ee1858f0ac8abc4a46aef"
	nonce           = "lubiesecurityskany"
	authUrlTemplate = "https://%s/oauth2/auth?client_id=%s&response_type=id_token&scope=openid&redirect_uri=%s&state=%s&nonce=%s"
)

var log = l.Default()

func main() {
	oidcClientIdPtr := flag.String("oidcClientId", "", "OIDC client ID")
	oidcClientRedirectUrlPtr := flag.String("oidcClientRedirectUrl", "", "OIDC client redirect URL")
	oidcProviderFqdnPtr := flag.String("oidcProviderFqdn", "", "FQDN of OIDC provider")
	emailPtr := flag.String("email", "admin@kyma.cx", "email")
	passwordPtr := flag.String("password", "1234", "password")
	flag.Parse()
	if *oidcClientIdPtr == "" || *oidcClientRedirectUrlPtr == "" || *oidcProviderFqdnPtr == "" {
		log.Fatal("oidcClientId, oidcClientRedirectURL, oidcProviderFqdn are required")
	}

	client, err := buildClient()
	if err != nil {
		l.Fatal(err)
	}
	config := oidcConfig{
		email:                 *emailPtr,
		password:              *passwordPtr,
		oidcClientId:          *oidcClientIdPtr,
		oidcClientRedirectUrl: *oidcClientRedirectUrlPtr,
	}
	h := OidcHydraTestFlow{
		hydraFqdn:  *oidcProviderFqdnPtr,
		httpClient: client,
		config:     config}
	token, err := h.GetToken()
	if err != nil {
		l.Fatal(err)
	}
	if token == "" {
		l.Fatal("token is empty")
	}
	fmt.Print(token)
}

func buildClient() (*http.Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	httpClient := &http.Client{
		Jar: jar,
	}
	return httpClient, err
}

func getToken(req *http.Request) string {
	urlFragments := strings.Split(req.URL.Fragment, "&")
	for i := 0; i < len(urlFragments); i++ {
		if strings.HasPrefix(urlFragments[i], "id_token") {
			return strings.Split(urlFragments[i], "=")[1]
		}
	}
	return ""
}

type oidcConfig struct {
	email                 string
	password              string
	oidcClientId          string
	oidcClientRedirectUrl string
}

// OidcHydraTestFlow implements OIDC flows for deployed OryHydra with test login consent endpoints.
type OidcHydraTestFlow struct {
	httpClient *http.Client
	config     oidcConfig
	hydraFqdn  string
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
	redirectUrl, err := url.Parse(f.config.oidcClientRedirectUrl)
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
	_, _ = f.httpClient.PostForm(response.Request.URL.String(), consentForm)
	return token, nil
}

func (f *OidcHydraTestFlow) doLogin() (*http.Response, error) {
	u, err := url.Parse(fmt.Sprintf(authUrlTemplate, f.hydraFqdn, f.config.oidcClientId, f.config.oidcClientRedirectUrl, state, nonce))
	if err != nil {
		return nil, err
	}
	resp, err := f.httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	loginForm := url.Values{}
	loginForm.Set("email", f.config.email)
	loginForm.Set("password", f.config.password)
	loginForm.Set("challenge", resp.Request.URL.Query().Get("login_challenge"))
	resp, err = f.httpClient.PostForm(resp.Request.URL.String(), loginForm)
	if err != nil {
		return nil, err
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
