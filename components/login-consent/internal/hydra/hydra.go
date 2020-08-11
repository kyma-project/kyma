package hydra

import (
	"bytes"
	"encoding/json"
	"fmt"
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
)

type client struct {
	hydraURL           url.URL
	mockTLSTermination bool
	httpClient         *http.Client
	ForwardedProto     string
}

func NewClient(httpClient *http.Client, url url.URL, mockTLSTermination bool) client {
	return client{
		hydraURL:           url,
		mockTLSTermination: mockTLSTermination,
		httpClient:         httpClient,
	}
}

func (c *client) GetLoginRequest(challenge string) (*hydraAPI.LoginRequest, error) {
	output := &hydraAPI.LoginRequest{}

	resp, err := c.get(loginFlow, challenge, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching login request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *client) AcceptLoginRequest(challenge string, body io.ReadCloser) (*hydraAPI.AcceptLoginRequest, error) {
	output := &hydraAPI.AcceptLoginRequest{}

	resp, err := c.put(loginFlow, actionAccept, challenge, body, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accept login request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *client) RejectLoginRequest(challenge string, body io.ReadCloser) (*hydraAPI.RejectRequest, error) {
	output := &hydraAPI.RejectRequest{}

	resp, err := c.put(loginFlow, actionReject, challenge, body, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accept login request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *client) GetConsentRequest(challenge string) (*hydraAPI.ConsentRequest, error) {
	output := &hydraAPI.ConsentRequest{}

	resp, err := c.get(consentFlow, challenge, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accept login request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *client) AcceptConsentRequest(challenge string, body io.ReadCloser) (*hydraAPI.AcceptConsentRequest, error) {
	output := new(hydraAPI.AcceptConsentRequest)

	resp, err := c.put(consentFlow, actionAccept, challenge, body, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accept login request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *client) RejectConsentRequest(challenge string, body io.ReadCloser) (*hydraAPI.RejectRequest, error) {
	output := new(hydraAPI.RejectRequest)

	resp, err := c.put(consentFlow, actionReject, challenge, body, output)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accept login request failed: %s", resp.Status)
	}

	return output, nil
}

func (c *client) get(flow, challenge string, output interface{}) (*http.Response, error) {

}

func (c *client) put(flow, action, challenge string, body io.ReadCloser, output interface{}) (*http.Response, error) {

}

func (c *client) newRequest(method, relativePath string, body interface{}) (*http.Request, error) {

	var buf io.ReadWriter
	if body != nil {
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

	if c.ForwardedProto != "" {
		req.Header.Add("X-Forwarded-Proto", c.ForwardedProto)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *client) do(req *http.Request, v interface{}) (*http.Response, error) {
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
