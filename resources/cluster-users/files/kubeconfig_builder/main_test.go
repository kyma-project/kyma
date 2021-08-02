package kubeconfig_builder

import (
	"context"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/oauth2"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
)

const (
	oryHydraOauthFqdn     = "oauth2.kyma.example.com"
	oidcClientId          = "1b834db7-429a-4864-8932-103a8278eb04"
	oidcClientRedirectUrl = "http://testclient3.example.com"
	state                 = "dd3557bfb07ee1858f0ac8abc4a46aef"
	nonce                 = "lubiesecurityskany"
	authUrlTemplate       = "https://%s/oauth2/auth?client_id=%s&response_type=id_token&scope=openid&redirect_uri=%s&state=%s&nonce=%s"
)

var QueryParams = []interface{}{oryHydraOauthFqdn, oidcClientId, oidcClientRedirectUrl, state, nonce}

func TestOryHydraLogin(t *testing.T) {
	u, err := url.Parse(fmt.Sprintf(authUrlTemplate, QueryParams...))
	if err != nil {
		t.Fatal(err)
	}
	httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL != u {
			t.Fatalf("req url does not match: %s", req.URL)
		}
	}))
}

func TestKubeconfig(t *testing.T) {
	_ = context.Background()
	config := oauth2.Config{
		ClientID:     oidcClientId,
		ClientSecret: state,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://" + oryHydraOauthFqdn + "/oauth2/auth",
			TokenURL:  "https://" + oryHydraOauthFqdn + "/oauth2/token",
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		RedirectURL: oidcClientRedirectUrl,
		Scopes:      []string{"openid"},
	}
	authCodeURL := config.AuthCodeURL(state)
	fmt.Println(authCodeURL)
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{
		Jar: jar,
	}
	resp, err := httpClient.Get(authCodeURL)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp.Request.URL)
	fmt.Println("second redirect - posting form")
	loginForm := url.Values{}
	loginForm.Set("email", "admin@kyma.cx")
	loginForm.Set("password", "1234")
	fmt.Println("loginForm: ", loginForm)
	fmt.Println("httpClient.Jar.Cookies: ", httpClient.Jar.Cookies(resp.Request.URL))
	resp, err = httpClient.PostForm(resp.Request.URL.String(), loginForm)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
	//resp, err := httpClient.PostForm(authCodeURL, loginForm)
	//token, err := config.PasswordCredentialsToken(ctx, "admin@kyma.cx", "1234")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fmt.Println(token.AccessToken)
}
