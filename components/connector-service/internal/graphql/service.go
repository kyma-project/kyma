package graphql

import (
	"bytes"
	"encoding/json"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type GraphQLService interface {
	ReadConfig(configStream io.Reader) (Config, error)
	SendRequest(query string, config Config, timeout time.Duration) (*http.Response, error)
}

type graphQLService struct{}

type Config struct {
	URL     string
	Headers Headers
}

type Headers map[string]string

func NewGraphQLService() GraphQLService {
	return &graphQLService{}
}

func (gs *graphQLService) ReadConfig(configStream io.Reader) (Config, error) {
	bytesValue, err := ioutil.ReadAll(configStream)

	if err != nil {
		return Config{}, apperrors.Internal("Error while reading config: %s", err.Error())
	}

	var config Config

	unmarshalErr := json.Unmarshal(bytesValue, &config)

	if unmarshalErr != nil {
		return Config{}, apperrors.Internal("Error while reading config: %s", unmarshalErr.Error())
	}

	return config, nil
}

func (gs *graphQLService) SendRequest(query string, config Config, timeout time.Duration) (*http.Response, error) {
	byteBody := []byte(query)

	request, e := http.NewRequest("POST", config.URL, bytes.NewBuffer(byteBody))

	if e != nil {
		return nil, apperrors.Internal("Error creating request: %s", e.Error())
	}

	for k, v := range config.Headers {
		request.Header.Set(k, v)
	}

	logrus.Info("GraphQL Request:", request)

	client := &http.Client{Timeout: timeout}

	response, err := client.Do(request)
	if err != nil {
		return nil, apperrors.Internal("Error sending request: %s", err)
	}

	return response, nil
}
