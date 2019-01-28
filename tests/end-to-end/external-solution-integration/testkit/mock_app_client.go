package testkit

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
)

type MockApplicationClient interface {
	ConnectToKyma(tokenURL string, isLocalKyma, shouldRegister bool) (connectResponse *ConnectResponse, err error)
	GetConnectionInfo() (connectResponse *ConnectResponse, err error)
}

type mockApplicationClient struct {
	mockBaseURL string
	httpClient  *http.Client
}

func NewMockApplicationClient() MockApplicationClient {
	mockBaseURL := os.Getenv("MOCKBASEURL")
	httpClient := NewHttpClient(true)

	return &mockApplicationClient{mockBaseURL: mockBaseURL, httpClient: httpClient}
}

func NewHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}

func (c *mockApplicationClient) ConnectToKyma(tokenURL string, isLocalKyma, shouldRegister bool) (connectResponse *ConnectResponse, err error) {
	connectRequest := &ConnectRequest{
		IsLocalKyma:        isLocalKyma,
		URL:                tokenURL,
		ShouldRegisterAPIs: shouldRegister,
		MockHostname:       c.mockBaseURL,
	}

	body, err := json.Marshal(connectRequest)
	if err != nil {
		return nil, err
	}

	url := c.mockBaseURL + "/connection"
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	connectResponse = &ConnectResponse{}

	err = json.NewDecoder(response.Body).Decode(&connectResponse)
	if err != nil {
		return nil, err
	}

	return connectResponse, nil
}

func (c *mockApplicationClient) GetConnectionInfo() (connectResponse *ConnectResponse, err error) {
	url := c.mockBaseURL + "/connection"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	connectResponse = &ConnectResponse{}

	err = json.NewDecoder(response.Body).Decode(&connectResponse)
	if err != nil {
		return nil, err
	}

	return connectResponse, nil
}
