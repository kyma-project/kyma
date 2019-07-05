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
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		Client: http.DefaultClient,
	}
}

// HTTPClient represents remote bundle repository which is accessed via HTTP.
type HTTPClient struct {
	RepositoryURL string
	Client        interface {
		Do(req *http.Request) (*http.Response, error)
	}
}

// Set url to bundle repository wich will be fetched
func (c *HTTPClient) SetURL(URL string) {
	c.RepositoryURL = URL
}

// IndexReader acquire repository index.
func (c *HTTPClient) IndexReader() (r io.ReadCloser, err error) {
	return c.doGetCall(c.RepositoryURL)
}

// BundleReader calls repository for a specific bundle and returns means to read bundle content.
func (c *HTTPClient) BundleReader(name, version string) (r io.ReadCloser, err error) {
	return c.doGetCall(c.URLForBundle(name, version))
}

// URLForBundle returns direct URL for getting the bundle
func (c *HTTPClient) URLForBundle(name, version string) string {
	return fmt.Sprintf("%s%s-%s.tgz", c.baseOfURL(c.RepositoryURL), name, version)
}

func (c *HTTPClient) baseOfURL(fullURL string) string {
	return strings.TrimRight(fullURL, path.Base(fullURL))
}

func (c *HTTPClient) doGetCall(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing request")
	}

	resp, err := c.Client.Do(req)
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
