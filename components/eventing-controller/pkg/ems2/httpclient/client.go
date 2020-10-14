package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	auth2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/auth"
)

type Client struct {
	httpClient http.Client
}

func NewHttpClient() *Client {
	return &Client{httpClient: http.Client{}}
}

func (c Client) NewRequest(token *auth2.AccessToken, method, url string, body interface{}) (*http.Request, *Error) {
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
	req.Header.Set("Authorization", c.getBearerToken(token))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c Client) Do(req *http.Request, result interface{}) (*http.Response, *Error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if resp == nil {
			return resp, NewError(err)
		}
		return resp, NewError(err, WithStatusCode(resp.StatusCode))
	}
	defer func() { _ = resp.Body.Close() }()
	defer c.httpClient.CloseIdleConnections()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, NewError(err, WithStatusCode(resp.StatusCode))
	}
	if len(body) == 0 {
		return resp, nil
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return resp, NewError(err, WithStatusCode(resp.StatusCode), WithMessage(string(body)))
	}

	return resp, nil
}

func (c Client) getBearerToken(token *auth2.AccessToken) string {
	return fmt.Sprintf("Bearer %s", token.Value)
}
