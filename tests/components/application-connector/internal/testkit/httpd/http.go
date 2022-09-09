package httpd

import (
	"io"
	"net/http"
	"testing"
)

type LogHttp struct {
	t       *testing.T
	httpCli *http.Client
}

func NewCli(t *testing.T) LogHttp {
	return LogHttp{
		t: t,
		httpCli: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c LogHttp) Get(url string) (resp *http.Response, body []byte, err error) {
	c.t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	return c.Do(req)

}

func (c LogHttp) Do(req *http.Request) (res *http.Response, body []byte, err error) {
	c.t.Helper()
	c.t.Logf("%s %s", req.Method, req.URL)

	res, err = c.httpCli.Do(req)
	if err != nil {
		return
	}

	body, err = io.ReadAll(res.Body)
	if err == nil && len(body) > 0 {
		c.t.Logf("Body: %s", body)
	}

	return
}
