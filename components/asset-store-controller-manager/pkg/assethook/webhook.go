package assethook

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"
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
	Do(ctx context.Context, contentType string, webhook v1alpha2.WebhookService, body io.Reader, response interface{}, timeout time.Duration) error
}

func New(httpClient HttpClient) Webhook {
	return &webhook{
		httpClient: httpClient,
	}
}

func (w *webhook) Do(ctx context.Context, contentType string, webhook v1alpha2.WebhookService, body io.Reader, response interface{}, timeout time.Duration) error {
	context, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequest("POST", w.getWebhookUrl(webhook), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.WithContext(context)

	rsp, err := w.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "while sending request to webhook")
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid response from %s, code: %d", req.URL, rsp.StatusCode)
	}

	responseBytes, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return errors.Wrapf(err, "while reading response body")
	}

	err = json.Unmarshal(responseBytes, response)
	if err != nil {
		return errors.Wrapf(err, "while parsing response body")
	}

	return nil
}

func (*webhook) getWebhookUrl(service v1alpha2.WebhookService) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local%s", service.Name, service.Namespace, service.Endpoint)
}
