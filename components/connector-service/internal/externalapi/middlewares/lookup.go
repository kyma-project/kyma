package middlewares

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"io/ioutil"
	"net/http"
	"os"
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
	return "", nil
}

func createRequest(context clientcontext.ApplicationContext, config LookUpConfig) {
	request := &http.Request{}

	setHeaders(request, config)
}

func setHeaders(request *http.Request, config LookUpConfig) {
	for k, v := range config.Headers {
		request.Header.Set(k, v)
	}
}

func readConfigFromFile(configFilePath string) (LookUpConfig, error) {
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

	json.Unmarshal(bytesValue, &config)

	return config, nil
}
