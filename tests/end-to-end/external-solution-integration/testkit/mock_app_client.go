package testkit

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
)

type MockApplicationClient interface {
	ConnectToKyma(tokenURL string, shouldRegister bool) (connectResponse *ConnectResponse, err error)
	GetConnectionInfo() (connectResponse *ConnectResponse, err error)
	GetAPIs() (apis *[]API, err error)
}

type mockApplicationClient struct {
	isLocalKyma bool
	mockBaseURL string
	httpClient  *http.Client
}

func NewMockApplicationClient() (MockApplicationClient, error) {
	mockBaseURL := os.Getenv("MOCKBASEURL")
	isLocal := os.Getenv("ISLOCALKYMA")
	isLocalKyma, err := strconv.ParseBool(isLocal)
	if err != nil {
		return nil, err
	}

	httpClient := NewHttpClient(true)

	return &mockApplicationClient{mockBaseURL: mockBaseURL, httpClient: httpClient, isLocalKyma: isLocalKyma}, nil
}

func NewHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}

func (c *mockApplicationClient) ConnectToKyma(tokenURL string, shouldRegister bool) (connectResponse *ConnectResponse, err error) {
	connectRequest := &ConnectRequest{
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

	request.ParseForm()
	request.Form.Add("localKyma", strconv.FormatBool(c.isLocalKyma))

	request.Header.Add("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	connectResponse = &ConnectResponse{}

	if response.StatusCode != 200 {
		panic(response.StatusCode)
	}

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

	request.ParseForm()
	request.Form.Set("localKyma", strconv.FormatBool(c.isLocalKyma))

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

func (c *mockApplicationClient) GetAPIs() (apis *[]API, err error) {
	url := c.mockBaseURL + "/apis"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	apisResponse := make([]API, 0)

	err = json.NewDecoder(response.Body).Decode(&apisResponse)
	if err != nil {
		return nil, err
	}

	return &apisResponse, nil
}
