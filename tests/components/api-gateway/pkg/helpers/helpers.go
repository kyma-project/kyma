package helpers

import (
	"io/ioutil"
	"net/http"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

//const testToken = "ZYqT86bNtVT-QViFpKGsmlnKGpovxVCQ8cMGsQQVU8A.WQC8MchDy-uyW2iIdqW7m26yZwmGAk_I6cR-YO-IiPY"

type Helper struct {
	client *http.Client
	opts   []retry.Option
}

func NewHelper(c *http.Client, opts []retry.Option) *Helper {
	return &Helper{
		client: c,
		opts:   opts,
	}
}

func (h *Helper) CallEndpointsWithRetries(url string, predicate *StatusPredicate) error {
	return h.withRetries(func() (*http.Response, error) {
		return h.client.Get(url)
	}, predicate.TestPredicate)
}

func (h *Helper) withRetries(httpCall func() (*http.Response, error), isResponseValid func(*http.Response) bool) error {

	if err := retry.Do(func() error {

		response, callErr := httpCall()
		if callErr != nil {
			return callErr
		}

		if !isResponseValid(response) {
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return errors.Errorf("unexpected response %s. Reason unknown: unable to parse response body: %s.", response.Status, err.Error())
			}
			return errors.Errorf("unexpected response %s: %s", response.Status, string(body))
		}

		return nil
	},
		h.opts...,
	); err != nil {
		return err
	}

	return nil
}

// StatusPredicate is a struct representing desired endpoint call status code, that is beetween LowerStatusBound and UpperStatusBound
type StatusPredicate struct {
	LowerStatusBound int
	UpperStatusBound int
}

func (s *StatusPredicate) TestPredicate(response *http.Response) bool {
	return response.StatusCode >= s.LowerStatusBound && response.StatusCode <= s.UpperStatusBound
}
