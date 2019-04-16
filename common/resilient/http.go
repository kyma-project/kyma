package resilient

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	retry "github.com/avast/retry-go"
)

type WrappedHttpClient struct {
	underlying HttpClient
	opts       []retry.Option
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewHttpClient(opts ...retry.Option) *WrappedHttpClient {
	return WrapHttpClient(&http.Client{}, opts...)
}

func WrapHttpClient(client HttpClient, opts ...retry.Option) *WrappedHttpClient {
	return &WrappedHttpClient{
		underlying: client,
		opts:       opts,
	}
}

func (c *WrappedHttpClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	err = retry.Do(func() error {
		resp, err = c.underlying.Do(req)
		if err != nil {
			return err
		}
		return nil
	}, c.opts...)
	return resp, err
}

func (c *WrappedHttpClient) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *WrappedHttpClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func (c *WrappedHttpClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *WrappedHttpClient) Head(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
