package testkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/rand"
	"net/http"
)

type RegistryClient interface {
	RegisterService(service *ServiceDetails) (string, error)
	GetService(id string) (*ServiceDetails, error)
	DeleteService(id string) error
}

type registryClient struct {
	url        string
	httpClient *http.Client
	logger     logrus.FieldLogger
}

func NewRegistryClient(url string, skipVerify bool, logger logrus.FieldLogger) RegistryClient {
	return &registryClient{
		url:        url,
		httpClient: newHttpClient(skipVerify),
		logger:     logger,
	}
}

func (rc *registryClient) RegisterService(service *ServiceDetails) (string, error) {
	body, err := json.Marshal(service)
	if err != nil {
		rc.logger.Error(err)
		return "", err
	}

	request, err := http.NewRequest(http.MethodPost, rc.url, bytes.NewReader(body))
	if err != nil {
		rc.logger.Error(err)
		return "", err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := rc.httpClient.Do(request)
	if err != nil {
		rc.logger.Error(err)
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		err := parseErrorResponse(response)
		rc.logger.Error(err)
		return "", err
	}

	registerServiceResponse := &RegisterServiceResponse{}

	err = json.NewDecoder(response.Body).Decode(registerServiceResponse)
	if err != nil {
		rc.logger.Error(err)
		return "", err
	}

	return registerServiceResponse.ID, nil
}

func (rc *registryClient) GetService(id string) (*ServiceDetails, error) {
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", rc.url, id), nil)
	if err != nil {
		rc.logger.Error(err)
		return nil, err
	}

	response, err := rc.httpClient.Do(request)
	if err != nil {
		rc.logger.Error(err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err := parseErrorResponse(response)
		rc.logger.Error(err)
		return nil, err
	}

	serviceDetails := &ServiceDetails{}

	err = json.NewDecoder(response.Body).Decode(serviceDetails)
	if err != nil {
		rc.logger.Error(err)
		return nil, err
	}

	return serviceDetails, nil
}

func (rc *registryClient) DeleteService(id string) error {
	request, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", rc.url, id), nil)
	if err != nil {
		rc.logger.Error(err)
		return err
	}

	response, err := rc.httpClient.Do(request)
	if err != nil {
		rc.logger.Error(err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		err := parseErrorResponse(response)
		rc.logger.Error(err)
		return err
	}

	return nil
}

func PrepareAppService(url string, labels map[string]string) *ServiceDetails {
	return &ServiceDetails{
		Name:        "test service",
		Provider:    "service provider",
		Description: "service description",
		Identifier:  generateIdentifier(),
		Labels:      labels,
		Api: &API{
			Spec:      ApiRawSpec,
			TargetUrl: url,
		},
		Documentation: &Documentation{
			DisplayName: "documentation name",
			Description: "documentation description",
			Type:        "documentation type",
			Tags:        []string{"tag1", "tag2"},
			Docs:        []DocsObject{{Title: "docs title", Type: "docs type", Source: "docs source"}},
		},
	}
}

func generateIdentifier() string {
	return fmt.Sprintf("identifier-%s", rand.String(8))
}
