package download

import (
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"io/ioutil"
	"net/http"
	"time"
)

const timeout = 5

type Client interface {
	Fetch(url string, credentials *authorization.Credentials) ([]byte, apperrors.AppError)
}

type downloader struct {
	client               *http.Client
	authorizationFactory authorization.StrategyFactory
}

func NewClient(client *http.Client, authFactory authorization.StrategyFactory) Client {
	return downloader{
		client:               client,
		authorizationFactory: authFactory,
	}
}

func (d downloader) Fetch(url string, credentials *authorization.Credentials) ([]byte, apperrors.AppError) {
	res, err := d.requestAPISpec(url, credentials)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, apperrors.Internal("Failed to fetch from Asset Store.")
	}

	{
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, apperrors.Internal("Failed to read response body from Asset Store.")
		}

		return bytes, nil
	}
}

func (d downloader) requestAPISpec(specUrl string, credentials *authorization.Credentials) (*http.Response, apperrors.AppError) {
	req, err := http.NewRequest(http.MethodGet, specUrl, nil)
	if err != nil {
		return nil, apperrors.Internal("Creating request for fetching API spec from %s failed, %s", specUrl, err.Error())
	}

	if credentials != nil {
		err := d.addAuthorization(req, credentials)
		if err != nil {
			return nil, apperrors.Internal("Adding authorization failed, %s", err.Error())
		}
	}

	response, err := d.client.Do(req)
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("Fetching API spec from %s failed, %s", specUrl, err.Error())
	}

	if response.StatusCode != http.StatusOK {
		return nil, apperrors.UpstreamServerCallFailed("Fetching API spec from %s failed with status %s", specUrl, response.Status)
	}

	return response, nil
}

func (d downloader) addAuthorization(r *http.Request, credentials *authorization.Credentials) apperrors.AppError {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second}

	ts := func(transport *http.Transport) {
		client.Transport = transport
	}

	strategy := d.authorizationFactory.Create(credentials)

	err := strategy.AddAuthorization(r, ts)

	if err != nil {
		return apperrors.Internal(err.Error())
	}

	return nil
}
