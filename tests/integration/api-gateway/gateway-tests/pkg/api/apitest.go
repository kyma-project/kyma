package api

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"net/http"
)

type Tester struct {
	client *http.Client
	opts   []retry.Option
}

func NewTester(c *http.Client, opts []retry.Option) *Tester {
	return &Tester{
		client: c,
		opts:   opts,
	}
}

func (h *Tester) TestSecuredAPI(url, token string) error {

	err := h.withRetries(func() (*http.Response, error) {
		return h.client.Get(url)
	}, httpUnauthorizedPredicate)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	return h.withRetries(func() (*http.Response, error) {
		return h.client.Do(req)
	}, httpOkPredicate)
}

func (h *Tester) TestUnsecuredAPI(url string) error {
	return h.withRetries(func() (*http.Response, error) {
		return h.client.Get(url)
	}, httpOkPredicate)
}

func (h *Tester) TestDeletedAPI(url string) error {
	return h.withRetries(func()(*http.Response,error ) {
		return h.client.Get(url)
	}, NotFoundPredicate)
}

func (h *Tester) withRetries(httpCall func() (*http.Response, error), shouldRetry func(*http.Response) bool) error {

	if err := retry.Do(func() error {

		response, callErr := httpCall()
		if callErr != nil {
			return callErr
		}

		if shouldRetry(response) {
			return errors.Errorf("unexpected response: %s", response.Status)
		}

		return nil
	},
		h.opts...
	); err != nil {
		return err
	}

	return nil
}

func httpOkPredicate(response *http.Response) bool {
	return response.StatusCode < 200 || response.StatusCode > 299
}

func httpUnauthorizedPredicate(response *http.Response) bool {
	return response.StatusCode != 401

}

func NotFoundPredicate(response *http.Response) bool {
	return response.StatusCode == http.StatusNotFound
}