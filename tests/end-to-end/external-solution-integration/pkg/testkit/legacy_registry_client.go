package testkit

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type LegacyRegistryClient struct {
	httpClient *http.Client
}

func NewLegacyRegistryClient(httpClient *http.Client) *LegacyRegistryClient {
	return &LegacyRegistryClient{
		httpClient: httpClient,
	}
}

func (lrc *LegacyRegistryClient) RegisterService(url string, service *ServiceDetails) (string, error) {
	body, err := json.Marshal(service)
	if err != nil {
		return "", errors.Wrap(err, "while marshalling service payload")
	}

	response, err := lrc.httpClient.Post(url, "", strings.NewReader(string(body)))
	if err != nil {
		return "", errors.Wrap(err, "while creating service instance")
	}

	if response.StatusCode != http.StatusOK {
		err := parseErrorResponse(response)
		return "", errors.Wrap(err, "error response")
	}

	registerServiceResponse := &RegisterServiceResponse{}

	err = json.NewDecoder(response.Body).Decode(registerServiceResponse)
	if err != nil {
		return "", errors.Wrap(err, "while decoding response body")
	}
	return registerServiceResponse.ID, nil
}
