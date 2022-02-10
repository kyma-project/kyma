package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/signals"
)

// compile time check
var _ BaseURLAwareClient = Client{}

type BaseURLAwareClient interface {
	NewRequest(method, path string, body interface{}) (*http.Request, *Error)
	Do(req *http.Request, result interface{}) (*http.Response, *[]byte, *Error)
}

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func NewHTTPClient(baseURL string, client *http.Client) (*Client, error) {
	ctx := signals.NewReusableContext()
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		httpClient: client,
		baseURL:    url,
	}, nil
}

func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

func (c Client) NewRequest(method, path string, body interface{}) (*http.Request, *Error) {
	var jsonBody io.ReadWriter
	if body != nil {
		jsonBody = new(bytes.Buffer)
		if err := json.NewEncoder(jsonBody).Encode(body); err != nil {
			return nil, NewError(err)
		}
	}

	pu, err := url.Parse(path)
	if err != nil {
		return nil, NewError(err)
	}
	u := c.baseURL.ResolveReference(pu)
	req, err := http.NewRequest(method, u.String(), jsonBody)
	if err != nil {
		return nil, NewError(err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c Client) Do(req *http.Request, result interface{}) (*http.Response, *[]byte, *Error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if resp == nil {
			return resp, nil, NewError(err)
		}
		return resp, nil, NewError(err, WithStatusCode(resp.StatusCode))
	}
	defer func() { _ = resp.Body.Close() }()
	defer c.httpClient.CloseIdleConnections()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, NewError(err, WithStatusCode(resp.StatusCode))
	}
	if len(body) == 0 {
		return resp, nil, nil
	}

	if err := json.Unmarshal(body, result); err != nil {
		return resp, nil, NewError(err, WithStatusCode(resp.StatusCode), WithMessage(string(body)))
	}

	return resp, &body, nil
}
