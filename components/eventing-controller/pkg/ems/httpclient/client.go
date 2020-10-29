package httpclient

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/signals"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"io"
	"io/ioutil"
	"net/http"
)

type Client struct {
	httpClient *http.Client
}

func NewHttpClient(cfg *clientcredentials.Config) *Client {
	ctx := signals.NewContext()
	httpClient := newOauth2Client(ctx, cfg)
	return &Client{httpClient: httpClient}
}

// NewClient returns a new HTTP client which have nested transports for handling oauth2 security, HTTP connection pooling, and tracing.
func newOauth2Client(ctx context.Context, cfg *clientcredentials.Config) *http.Client {
	// create and configure oauth2 client
	client := cfg.Client(ctx)

	var base = http.DefaultTransport.(*http.Transport).Clone()
	client.Transport.(*oauth2.Transport).Base = base

	// TODO: Support tracing in eventing-controller #9767: https://github.com/kyma-project/kyma/issues/9767
	return client
}

func (c *Client) GetHttpClient() *http.Client {
	return c.httpClient
}

func (c Client) NewRequest(method, url string, body interface{}) (*http.Request, *Error) {
	var jsonBody io.ReadWriter
	if body != nil {
		jsonBody = new(bytes.Buffer)
		if err := json.NewEncoder(jsonBody).Encode(body); err != nil {
			return nil, NewError(err)
		}
	}

	req, err := http.NewRequest(method, url, jsonBody)
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
