package resilient

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	retry "github.com/avast/retry-go"
)

// WrappedHttpClient is a wrapper around HttpClient which retries requests in case of errors. It implements all of public
// methods available in http.Client
type WrappedHttpClient struct {
	underlying HttpClient
	opts       []retry.Option
}

// HttpClient is a simplified version of http.Client interface
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHttpClient returns new WrappedHttpClient with default http.Client as upstream. If you want to configure it in
// anyway use WrapHttpClient instead.
func NewHttpClient(opts ...retry.Option) *WrappedHttpClient {
	return WrapHttpClient(&http.Client{}, opts...)
}

// WrapHttpClient returns new WrappedHttpClient with given HttpClient as upstream
func WrapHttpClient(client HttpClient, opts ...retry.Option) *WrappedHttpClient {
	return &WrappedHttpClient{
		underlying: client,
		opts:       opts,
	}
}

// Do calls Do method of underlying HttpClient and retries according to given options when an error occurs
func (c *WrappedHttpClient) Do(req *http.Request) (resp *http.Response, err error) {
	err = retry.Do(func() error {
		resp, err = c.underlying.Do(req)
		if err != nil {
			return err
		}
		return nil
	}, c.opts...)
	return resp, err
}

// Get is copied from http.Client to be compliant with its interface. For more documentation see http.Client.Get
func (c *WrappedHttpClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post is copied from http.Client to be compliant with its interface. For more documentation see http.Client.Post
func (c *WrappedHttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// PostForm is copied from http.Client to be compliant with its interface. For more documentation see http.Client.PostForm
func (c *WrappedHttpClient) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

// Head is copied from http.Client to be compliant with its interface. For more documentation see http.Client.Head
func (c *WrappedHttpClient) Head(url string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
