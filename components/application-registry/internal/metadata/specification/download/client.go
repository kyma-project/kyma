package download

import (
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/csrf"
	csrfClient "github.com/kyma-project/kyma/components/application-gateway/pkg/csrf/client"
	csrfStrategy "github.com/kyma-project/kyma/components/application-gateway/pkg/csrf/strategy"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"io/ioutil"
	"net/http"
	"time"
)

type Client interface {
	Fetch(url string) ([]byte, apperrors.AppError)
}

type downloader struct {
	client               *http.Client
	authorizationFactory authorization.StrategyFactory
	csrfFactory          csrf.TokenStrategyFactory
}

func NewClient(client *http.Client) Client {
	return downloader{
		client:               client,
		authorizationFactory: authorization.NewStrategyFactory(authorization.FactoryConfiguration{}),
		csrfFactory:          csrfStrategy.NewTokenStrategyFactory(csrfClient.New(10, nil, nil)),
	}
}

func (d downloader) Fetch(url string) ([]byte, apperrors.AppError) {

	// TODO - Remove this function and use the function performing authorisation calls
	res, err := d.requestAPISpec(url)
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

func (d downloader) FetchSecured(url string, credentials *authorization.Credentials) ([]byte, apperrors.AppError) {

	factory := authorization.NewStrategyFactory(authorization.FactoryConfiguration{})
	s := factory.Create(nil)

	s.AddAuthorization(nil, nil)

	return nil, nil
}

func (d downloader) newCSRFTokenStrategy(authorizationStrategy authorization.Strategy, credentials *authorization.Credentials) csrf.TokenStrategy {
	csrfTokenEndpointURL := ""
	if credentials != nil {
		csrfTokenEndpointURL = credentials.CSRFTokenEndpointURL
	}
	return d.csrfFactory.Create(authorizationStrategy, csrfTokenEndpointURL)
}

func (d downloader) newRequest(specUrl string, credentials *authorization.Credentials) (*http.Response, apperrors.AppError) {
	req, err := http.NewRequest(http.MethodGet, specUrl, nil)
	if err != nil {
		return nil, apperrors.Internal("Creating request for fetching API spec from %s failed, %s", specUrl, err.Error())
	}

	strategy := d.authorizationFactory.Create(credentials)

	client := &http.Client{
		Timeout: time.Duration(5) * time.Second}

	ts := func(transport *http.Transport) {
		client.Transport = transport
	}

	strategy.AddAuthorization(req, ts)

	return nil, nil
}

func (d downloader) requestAPISpec(specUrl string) (*http.Response, apperrors.AppError) {
	req, err := http.NewRequest(http.MethodGet, specUrl, nil)
	if err != nil {
		return nil, apperrors.Internal("Creating request for fetching API spec from %s failed, %s", specUrl, err.Error())
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
