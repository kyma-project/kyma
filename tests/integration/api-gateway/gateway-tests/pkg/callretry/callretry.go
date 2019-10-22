package callretry

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"net/http"
)

type securityValidator struct {
	securedClient, unsecuredClient *http.Client
	opts                           []retry.Option
}

func NewSecurityValidator(c *http.Client, opts []retry.Option) *securityValidator {
	return &securityValidator{
		securedClient:   c,
		unsecuredClient: &http.Client{},
		opts:            opts,
	}
}

func (h *securityValidator) ValidateSecureAPI(url string) error {

	err := h.withRetries(func() (*http.Response, error) {
		return h.securedClient.Get(fmt.Sprintf("https://%s", url))
	}, httpOkPredicate)

	if err != nil {
		return err
	}

	return h.withRetries(func() (*http.Response, error) {
		return h.unsecuredClient.Get(fmt.Sprintf("https://%s", url))
	}, httpUnauthorizedPredicate)
}

func (h *securityValidator) ValidateInsecureAPI(url string) error {
	return h.withRetries(func() (*http.Response, error) {
		return h.unsecuredClient.Get(fmt.Sprintf("https://%s", url))
	}, httpOkPredicate)
}

func (h *securityValidator) withRetries(httpCall func() (*http.Response, error), shouldRetry func(*http.Response) bool) error {

	if err := retry.Do(func() error {

		var callErr error
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
