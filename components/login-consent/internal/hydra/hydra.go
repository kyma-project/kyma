package hydra

import (
	"bytes"
	"encoding/json"
	"fmt"
	httpheaders "github.com/go-http-utils/headers"
	hydraAPI "github.com/ory/hydra-client-go/models"
	"io"
	"net/http"
	"net/url"
	"path"
)

const (
	loginFlow   = "login"
	consentFlow = "consent"

	actionAccept = "accept"
	actionReject = "reject"

	requestsEndpoint = "/oauth2/auth/requests"

	fmtChallenge = "%s_challenge"
)

type LoginConsentClient struct {
	hydraURL       url.URL
	httpClient     *http.Client
	forwardedProto string
}

func NewClient(httpClient *http.Client, url url.URL, forwardedProto string) LoginConsentClient {
	return LoginConsentClient{
		hydraURL:       url,
		httpClient:     httpClient,
		forwardedProto: forwardedProto,
	}
}

func (c *LoginConsentClient) GetLoginRequest(challenge string) (*hydraAPI.LoginRequest, error) {
	output := new(hydraAPI.LoginRequest)

	resp, err := c.get(loginFlow, challenge, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching login request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *LoginConsentClient) AcceptLoginRequest(challenge string, body *hydraAPI.AcceptLoginRequest) (*hydraAPI.CompletedRequest, error) {
	output := new(hydraAPI.CompletedRequest)

	resp, err := c.put(loginFlow, actionAccept, challenge, body, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accept login request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *LoginConsentClient) RejectLoginRequest(challenge string, body *hydraAPI.RejectRequest) (*hydraAPI.CompletedRequest, error) {
	output := new(hydraAPI.CompletedRequest)

	resp, err := c.put(loginFlow, actionReject, challenge, body, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reject login request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *LoginConsentClient) GetConsentRequest(challenge string) (*hydraAPI.ConsentRequest, error) {
	output := new(hydraAPI.ConsentRequest)

	resp, err := c.get(consentFlow, challenge, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get consent request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *LoginConsentClient) AcceptConsentRequest(challenge string, body *hydraAPI.AcceptConsentRequest) (*hydraAPI.CompletedRequest, error) {
	output := new(hydraAPI.CompletedRequest)

	resp, err := c.put(consentFlow, actionAccept, challenge, body, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accept consent request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *LoginConsentClient) RejectConsentRequest(challenge string, rejectRequest *hydraAPI.RejectRequest) (*hydraAPI.CompletedRequest, error) {
	output := new(hydraAPI.CompletedRequest)

	resp, err := c.put(consentFlow, actionReject, challenge, rejectRequest, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reject consent request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *LoginConsentClient) get(flow, challenge string, output interface{}) (*http.Response, error) {

	relPath := path.Join(requestsEndpoint, flow)

	params := map[string]string{
		fmt.Sprintf(fmtChallenge, flow): challenge,
	}

	req, err := c.newRequest(http.MethodGet, relPath, params, nil)
	if err != nil {
		return nil, err
	}

	return c.do(req, output)
}

func (c *LoginConsentClient) put(flow, action, challenge string, body, output interface{}) (*http.Response, error) {

	relPath := path.Join(requestsEndpoint, flow, action)

	params := map[string]string{
		fmt.Sprintf(fmtChallenge, flow): challenge,
	}

	req, err := c.newRequest(http.MethodPost, relPath, params, body)
	if err != nil {
		return nil, err
	}

	return c.do(req, output)
}

func (c *LoginConsentClient) newRequest(method, relativePath string, params map[string]string, body interface{}) (*http.Request, error) {

	headers := map[string]string{
		httpheaders.Accept: "application/json",
	}

	if c.forwardedProto != "" {
		headers[httpheaders.XForwardedProto] = c.forwardedProto
	}

	var buf io.ReadWriter
	if body != nil {
		headers[httpheaders.ContentType] = "application/json"
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	u := c.hydraURL
	u.Path = path.Join(u.Path, relativePath)

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	for h, v := range headers {
		req.Header.Set(h, v)
	}

	q := req.URL.Query()
	for p, v := range params {
		q.Add(p, v)
	}

	req.URL.RawQuery = q.Encode()

	return req, nil
}

func (c *LoginConsentClient) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if v != nil && resp.StatusCode < 300 {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}
