package ybundle

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// RepositoryConfig provides configuration for HTTP Repository.
type RepositoryConfig struct {
	BaseURL string `json:"baseUrl" valid:"required"`
}

// NewHTTPRepository creates new instance of HTTPRepository.
func NewHTTPRepository(cfg RepositoryConfig) *HTTPRepository {
	return &HTTPRepository{
		BaseURL: cfg.BaseURL,
		Client:  http.DefaultClient,
	}
}

// HTTPRepository represents remote bundle repository which is accessed via HTTP.
type HTTPRepository struct {
	BaseURL string
	Client  interface {
		Do(req *http.Request) (*http.Response, error)
	}
}

// IndexReader acquire repository index.
func (p *HTTPRepository) IndexReader() (r io.Reader, closer func(), err error) {
	return p.doGetCall("index.yaml")
}

// BundleReader calls repository for a specific bundle and returns means to read bundle content.
func (p *HTTPRepository) BundleReader(name BundleName, version BundleVersion) (r io.Reader, closer func(), err error) {
	bundleFileName := func(n BundleName, v BundleVersion) string {
		return fmt.Sprintf("%s-%s.tgz", n, v)
	}

	return p.doGetCall(bundleFileName(name, version))
}

func (p *HTTPRepository) doGetCall(urlPart string) (r io.Reader, closer func(), err error) {
	req, err := http.NewRequest(http.MethodGet, p.fullPath(urlPart), http.NoBody)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while preparing request")
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while calling")
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("got http error: status=%d body unavailable due to error %s", resp.StatusCode, err.Error())
		}
		return nil, nil, fmt.Errorf("got http error: status=%d body='%s'", resp.StatusCode, string(body))
	}

	return resp.Body, func() { resp.Body.Close() }, nil
}

func (p *HTTPRepository) fullPath(part string) string {
	normalisedBaseURL := strings.TrimRight(p.BaseURL, "/")
	urlParts := []string{normalisedBaseURL, part}
	return strings.Join(urlParts, "/")
}
