package request

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

func GetHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	return client
}

func DoGet(httpClient *http.Client, url string, authHeader string) (int, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, "", errors.Wrap(err, "cannot create a new HTTP request")
	}
	req.Header.Add("Authorization", authHeader)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, "", errors.Wrapf(err, "cannot send HTTP request to %s", url)
	}
	defer resp.Body.Close()
	var body bytes.Buffer
	if _, err := io.Copy(&body, resp.Body); err != nil {
		return 0, "", errors.Wrap(err, "cannot read response body")
	}
	return resp.StatusCode, body.String(), nil
}
