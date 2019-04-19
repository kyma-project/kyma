package fetch_dex_token

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"log"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

type idTokenProvider interface {
	fetchIdToken() (string, error)
}

type dexIdTokenProvider struct {
	httpClient *http.Client
	config     IdProviderConfig
}

func newDexIdTokenProvider(httpClient *http.Client, config IdProviderConfig) idTokenProvider {
	// turn off follow-redirects
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
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

	authorizeResp, err := p.httpClient.PostForm(p.config.DexConfig.AuthorizeEndpoint, url.Values{
		"response_type": {"id_token token"},
		"client_id":     {p.config.ClientConfig.ID},
		"redirect_uri":  {p.config.ClientConfig.RedirectUri},
		"scope":         {"openid profile email groups"},
		"nonce":         {"vF7FAQlqq41CObeUFYY0ggv1qEELvfHaXQ0ER4XM"},
	})
	if err != nil {
		return nil, err
	}
	defer closeRespBody(authorizeResp)
	authorizeRespBody := readRespBody(authorizeResp)

	switch authorizeResp.StatusCode {
	case http.StatusFound:
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("got unexpected response on authorize: %d - %s", authorizeResp.StatusCode, authorizeRespBody)
	}

	var loginEndpoint string
	if authorizeResp.StatusCode == http.StatusOK {
		b := bytes.NewBufferString(authorizeRespBody)
		loginEndpoint, err = getLocalAuthEndpoint(b)
		if err != nil {
			return nil, errors.Wrapf(err, "while fetching link to static authentication")
		}
	} else {
		loginEndpoint = authorizeResp.Header.Get("location")
		if strings.Contains(loginEndpoint, "#.*error") {
			return nil, fmt.Errorf("login - Redirected with error: '%s'", loginEndpoint)
		}
	}

	if _, err := p.httpClient.Get(p.config.DexConfig.BaseUrl + loginEndpoint); err != nil {
		return nil, errors.Wrap(err, "while performing HTTP GET on login endpoint")
	}

	loginResp, err := p.httpClient.PostForm(p.config.DexConfig.BaseUrl+loginEndpoint, url.Values{
		"login":    {p.config.UserCredentials.Username},
		"password": {p.config.UserCredentials.Password},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while performing HTTP POST on login endpoint")
	}
	defer closeRespBody(loginResp)
	loginRespBody := readRespBody(loginResp)

	if loginResp.StatusCode < 300 || loginResp.StatusCode > 399 {
		return nil, fmt.Errorf("login - response error: '%s' - %s", loginResp.Status, loginRespBody)
	}

	approvalEndpoint := loginResp.Header.Get("location")
	if strings.Contains(approvalEndpoint, "#.*error") {
		return nil, fmt.Errorf("approval - Redirected with error: '%s'", approvalEndpoint)
	}
	approvalResp, err := p.httpClient.Get(p.config.DexConfig.BaseUrl + approvalEndpoint)
	if err != nil {
		return nil, err
	}
	defer closeRespBody(approvalResp)
	approvalRespBody := readRespBody(approvalResp)

	if approvalResp.StatusCode < 300 || approvalResp.StatusCode > 399 {
		return nil, errors.New(fmt.Sprintf("Approval - response error: '%s' - %s", approvalResp.Status, approvalRespBody))
	}

	clientEndpoint := approvalResp.Header.Get("location")
	if strings.Contains(clientEndpoint, "#.*error") {
		return nil, fmt.Errorf("client - Redirected with error: '%s'", clientEndpoint)
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

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("WARNING: Unable to read response body (status: '%s'). Root cause: %v", resp.Status, err)
		return "<<Error reading response body>>"
	}
	return string(b)
}

func closeRespBody(resp *http.Response) {
	err := resp.Body.Close()
	if err != nil {
		log.Printf("WARNING: Unable to close response body. Cause: %v", err)
	}
}

func getLocalAuthEndpoint(body io.Reader) (string, error) {
	z := html.NewTokenizer(body)
	for {
		nt := z.Next()
		if nt == html.ErrorToken {
			return "", errors.New("got HTML error token")
		}

		token := z.Token()
		if "a" != token.Data {
			continue
		}
		for _, attr := range token.Attr {
			if attr.Key != "href" {
				continue
			}
			match, err := regexp.MatchString("/auth/local.*", attr.Val)
			if err != nil {
				log.Printf("WARNING: Unable to match string. Cause: %v", err)
				return "", err
			}
			if match {
				return attr.Val, nil
			}
		}
	}
}
