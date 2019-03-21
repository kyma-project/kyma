package assethook

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type webhook struct {
	httpClient HttpClient
}

//go:generate mockery -name=HttpClient -output=automock -outpkg=automock -case=underscore
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

//go:generate mockery -name=Webhook -output=automock -outpkg=automock -case=underscore
type Webhook interface {
	Call(ctx context.Context, url string, request interface{}, response interface{}) error
}

func New(httpClient HttpClient) Webhook {
	return &webhook{
		httpClient: httpClient,
	}
}

func (w *webhook) Call(ctx context.Context, url string, request interface{}, response interface{}) error {
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return errors.Wrapf(err, "while converting request to JSON")
	}

	httpRequest, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.WithContext(ctx)

	httpResponse, err := w.httpClient.Do(httpRequest)
	if err != nil {
		return errors.Wrapf(err, "while sending request to webhook")
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode > 299 {
		return errors.New(httpResponse.Status)
	}

	responseBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return errors.Wrapf(err, "while reading response body")
	}

	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return errors.Wrapf(err, "while parsing response body")
	}

	return nil
}
