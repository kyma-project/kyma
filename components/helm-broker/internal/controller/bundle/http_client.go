package bundle

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// NewHTTPClient creates new instance of HTTPClient.
func NewHTTPClient(URL string) *HTTPClient {
	return &HTTPClient{
		RepositoryURL: URL,
		Client:        http.DefaultClient,
	}
}

// HTTPClient represents remote bundle repository which is accessed via HTTP.
type HTTPClient struct {
	RepositoryURL string
	Client        interface {
		Do(req *http.Request) (*http.Response, error)
	}
}

// IndexReader acquire repository index.
func (p *HTTPClient) IndexReader() (r io.ReadCloser, err error) {
	return p.doGetCall(p.RepositoryURL)
}

// BundleReader calls repository for a specific bundle and returns means to read bundle content.
func (p *HTTPClient) BundleReader(name, version string) (r io.ReadCloser, err error) {
	return p.doGetCall(p.URLForBundle(name, version))
}

// URLForBundle returns direct URL for getting the bundle
func (p *HTTPClient) URLForBundle(name, version string) string {
	return fmt.Sprintf("%s%s-%s.tgz", p.baseOfURL(p.RepositoryURL), name, version)
}

func (p *HTTPClient) baseOfURL(fullURL string) string {
	return strings.TrimRight(fullURL, path.Base(fullURL))
}

func (p *HTTPClient) doGetCall(url string) (io.ReadCloser, error) {
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
