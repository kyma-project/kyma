package download

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"io/ioutil"
	"net/http"
)

type Client interface {
	Fetch(url string) ([]byte, apperrors.AppError)
}

type downloader struct {
	client *http.Client
}

func NewClient(client *http.Client) Client {
	return downloader{
		client: client,
	}
}

func (d downloader) Fetch(url string) ([]byte, apperrors.AppError) {
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
