package bundle

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// NewHTTPRepository creates new instance of HTTPRepository.
func NewHTTPRepository(cfg RepositoryConfig) *HTTPRepository {
	return &HTTPRepository{
		IndexFile: cfg.IndexFileName(),
		BaseURL:   cfg.BaseURL(),
		Client:    http.DefaultClient,
	}
}

// HTTPRepository represents remote bundle repository which is accessed via HTTP.
type HTTPRepository struct {
	IndexFile string
	BaseURL   string
	Client    interface {
		Do(req *http.Request) (*http.Response, error)
	}
}

// IndexReader acquire repository index.
func (p *HTTPRepository) IndexReader() (r io.ReadCloser, err error) {
	return p.doGetCall(p.BaseURL + p.IndexFile)
}

// BundleReader calls repository for a specific bundle and returns means to read bundle content.
func (p *HTTPRepository) BundleReader(name Name, version Version) (r io.ReadCloser, err error) {
	return p.doGetCall(p.URLForBundle(name, version))
}

// URLForBundle returns direct URL for getting the bundle
func (p *HTTPRepository) URLForBundle(name Name, version Version) string {
	return fmt.Sprintf("%s%s-%s.tgz", p.BaseURL, name, version)
}

func (p *HTTPRepository) doGetCall(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing request")
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "while calling")
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("got http error: status=%d body unavailable due to error %s", resp.StatusCode, err.Error())
		}
		return nil, fmt.Errorf("got http error: status=%d body='%s'", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}
