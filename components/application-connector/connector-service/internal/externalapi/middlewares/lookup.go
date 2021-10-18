package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/clientcontext"
	"github.com/tidwall/gjson"
)

const (
	timeout  = 30 * time.Second
	filename = "config.json"
)

type LookupService interface {
	Fetch(context clientcontext.ApplicationContext, configFilePath string) (string, error)
}

type GraphQLLookupService struct{}

func NewGraphQLLookupService() *GraphQLLookupService {
	return &GraphQLLookupService{}
}

type LookUpConfig struct {
	URL     string
	Headers Headers
}

type Headers map[string]string

func (ls *GraphQLLookupService) Fetch(context clientcontext.ApplicationContext, configFilePath string) (string, error) {
	lookUpConfig, e := readConfig(configFilePath + filename)

	if e != nil {
		return "", apperrors.Internal("Error while reading config file: %s", e)
	}

	request, err := createRequest(context, lookUpConfig)

	if err != nil {
		return "", apperrors.Internal("Error while creating request: %s", e)
	}

	return sendRequest(request)
}

func readConfig(configFilePath string) (LookUpConfig, error) {
	jsonFile, e := os.Open(configFilePath)

	if e != nil {
		return LookUpConfig{}, e
	}

	defer jsonFile.Close()

	bytesValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		return LookUpConfig{}, err
	}

	var config LookUpConfig

	unmarshalErr := json.Unmarshal(bytesValue, &config)

	if unmarshalErr != nil {
		return LookUpConfig{}, unmarshalErr
	}

	return config, nil
}

func createRequest(context clientcontext.ApplicationContext, config LookUpConfig) (*http.Request, error) {
	query := `{"query":"{ applications(where: {accountId: \"%s\", groupName: \"%s\", appName: \"%s\"}) {name account { id } groups { id name clusters { id name endpoints { gateway } } } } }"}`

	body := fmt.Sprintf(query, context.Tenant, context.Group, context.Application)

	byteBody := []byte(body)

	request, e := http.NewRequest("POST", config.URL, bytes.NewBuffer(byteBody))

	if e != nil {
		return nil, e
	}

	for k, v := range config.Headers {
		request.Header.Set(k, v)
	}

	logrus.Info("Request:", request)

	return request, nil
}

func sendRequest(request *http.Request) (string, error) {
	client := &http.Client{}
	client.Timeout = timeout

	response, err := client.Do(request)
	if err != nil {
		return "", apperrors.Internal("Error sending request: %s", err)
	}
	defer response.Body.Close()

	body, e := ioutil.ReadAll(response.Body)

	if e != nil {
		return "", apperrors.Internal("Error reading response body: %s", e)
	}

	return getGatewayUrl(body).Str, nil
}

func getGatewayUrl(body []byte) gjson.Result {
	stringBody := string(body)
	logrus.Info(stringBody)
	gatewayUrl := gjson.Get(stringBody, "data.applications.0.groups.0.clusters.0.endpoints.gateway")
	return gatewayUrl
}
