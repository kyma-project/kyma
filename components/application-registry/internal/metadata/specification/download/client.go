package download

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

const timeout = 5

type Client interface {
	Fetch(url string, credentials *authorization.Credentials, parameters *model.RequestParameters) ([]byte, apperrors.AppError)
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

func (d downloader) Fetch(url string, credentials *authorization.Credentials, parameters *model.RequestParameters) ([]byte, apperrors.AppError) {
	res, err := d.requestAPISpec(url, credentials, parameters)
	if err != nil {
		return nil, err
	}

	{
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, apperrors.Internal("Failed to read response body from %s.", url)
		}

		return bytes, nil
	}
}

func (d downloader) requestAPISpec(specUrl string, credentials *authorization.Credentials, parameters *model.RequestParameters) (*http.Response, apperrors.AppError) {
	req, err := http.NewRequest(http.MethodGet, specUrl, nil)
	if err != nil {
		return nil, apperrors.Internal("Creating request for fetching API spec from %s failed, %s", specUrl, err.Error())
	}

	if credentials != nil {
		err := d.addAuthorization(req, credentials)
		if err != nil {
			return nil, apperrors.UpstreamServerCallFailed("Adding authorization for fetching API spec from %s failed, %s", specUrl, err.Error())
		}
	}

	if parameters != nil {
		setCustomQueryParameters(req.URL, parameters.QueryParameters)
		setCustomHeaders(req.Header, parameters.Headers)
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

	ts := func(transport *http.Transport) {
		d.client.Transport = transport
	}

	strategy := d.authorizationFactory.Create(credentials)

	err := strategy.AddAuthorization(r, ts)

	if err != nil {
		return apperrors.Internal(err.Error())
	}

	return nil
}

func setCustomQueryParameters(url *url.URL, queryParams *map[string][]string) {
	if queryParams == nil {
		return
	}

	reqQueryValues := url.Query()

	for customQueryParam, values := range *queryParams {
		if _, ok := reqQueryValues[customQueryParam]; ok {
			continue
		}

		reqQueryValues[customQueryParam] = values
	}

	url.RawQuery = reqQueryValues.Encode()
}

func setCustomHeaders(reqHeaders http.Header, customHeaders *map[string][]string) {
	if _, ok := reqHeaders[httpconsts.HeaderUserAgent]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		reqHeaders.Set(httpconsts.HeaderUserAgent, "")
	}

	setHeaders(reqHeaders, customHeaders)
}

func setHeaders(reqHeaders http.Header, customHeaders *map[string][]string) {
	if customHeaders == nil {
		return
	}

	for header, values := range *customHeaders {
		if _, ok := reqHeaders[header]; ok {
			// if header is already specified we do not interfere with it
			continue
		}

		reqHeaders[header] = values
	}
}
