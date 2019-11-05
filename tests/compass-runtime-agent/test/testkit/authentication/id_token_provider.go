package authentication

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

	"github.com/avast/retry-go"
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

	var result map[string]string = nil

	happyRun := true
	err := retry.Do(
		func() error {

			loginEndpoint, err := p.initializeImplicitFlow()
			if err != nil {
				return err
			}

			approvalEndpoint, err := p.login(loginEndpoint)
			if err != nil {
				return err
			}

			finalRedirectURL, err := p.receiveToken(approvalEndpoint)
			if err != nil {
				return err
			}

			result = p.parseTokenResponse(finalRedirectURL)
			return nil
		},
		retry.Attempts(p.config.RetryConfig.MaxAttempts),
		retry.Delay(p.config.RetryConfig.Delay),
		retry.DelayType(retry.FixedDelay),
		retry.OnRetry(func(retryNo uint, err error) {
			happyRun = false
			log.Printf("Retry: [%d / %d], error: %s", retryNo, p.config.RetryConfig.MaxAttempts, err)
		}),
	)

	if err != nil {
		return nil, err
	}

	if happyRun {
		log.Println("Flow finished flawlessly")
	}

	return result, nil
}

//Performs call to the <authorize> endpoint.
func (p *dexIdTokenProvider) initializeImplicitFlow() (string, error) {

	authorizeResp, err := p.httpClient.PostForm(p.config.DexConfig.AuthorizeEndpoint, url.Values{
		"response_type": {"id_token token"},
		"client_id":     {p.config.ClientConfig.ID},
		"redirect_uri":  {p.config.ClientConfig.RedirectUri},
		"scope":         {"openid profile email groups"},
		"nonce":         {"vF7FAQlqq41CObeUFYY0ggv1qEELvfHaXQ0ER4XM"},
	})
	if err != nil {
		return "", err
	}
	defer closeRespBody(authorizeResp)
	authorizeRespBody := readRespBody(authorizeResp)

	switch authorizeResp.StatusCode {
	case http.StatusFound:
	case http.StatusOK:
	default:
		return "", fmt.Errorf("got unexpected response on authorize: %d - %s", authorizeResp.StatusCode, authorizeRespBody)
	}

	var loginEndpoint string
	if authorizeResp.StatusCode == http.StatusOK {
		b := bytes.NewBufferString(authorizeRespBody)
		loginEndpoint, err = getLocalAuthEndpoint(b)
		if err != nil {
			return "", errors.Wrapf(err, "while fetching link to static authentication")
		}
	} else {
		loginEndpoint = authorizeResp.Header.Get("location")
		if strings.Contains(loginEndpoint, "#.*error") {
			return "", fmt.Errorf("login - Redirected with error: '%s'", loginEndpoint)
		}
	}

	return loginEndpoint, nil
}

//Handles redirect to login endpoint
func (p *dexIdTokenProvider) login(loginEndpoint string) (string, error) {

	if _, err := p.httpClient.Get(p.config.DexConfig.BaseUrl + loginEndpoint); err != nil {
		return "", errors.Wrap(err, "while performing HTTP GET on login endpoint")
	}

	loginResp, err := p.httpClient.PostForm(p.config.DexConfig.BaseUrl+loginEndpoint, url.Values{
		"login":    {p.config.UserCredentials.Username},
		"password": {p.config.UserCredentials.Password},
	})
	if err != nil {
		return "", errors.Wrap(err, "while performing HTTP POST on login endpoint")
	}
	defer closeRespBody(loginResp)
	loginRespBody := readRespBody(loginResp)

	if loginResp.StatusCode < 300 || loginResp.StatusCode > 399 {
		return "", fmt.Errorf("login - response error: '%s' - %s", loginResp.Status, loginRespBody)
	}

	approvalEndpoint := loginResp.Header.Get("location")
	if strings.Contains(approvalEndpoint, "#.*error") {
		return "", fmt.Errorf("approval - Redirected with error: '%s'", approvalEndpoint)
	}

	return approvalEndpoint, nil
}

//Handles redirect to approval endpoint to get the token (end of flow)
func (p *dexIdTokenProvider) receiveToken(approvalEndpoint string) (*url.URL, error) {

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

	return parsedUrl, nil
}

func (p *dexIdTokenProvider) parseTokenResponse(parsedUrl *url.URL) map[string]string {

	result := make(map[string]string)
	fragmentParams := strings.Split(parsedUrl.Fragment, "&")
	for _, param := range fragmentParams {
		keyAndValue := strings.Split(param, "=")
		result[keyAndValue[0]] = keyAndValue[1]
	}

	return result
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
