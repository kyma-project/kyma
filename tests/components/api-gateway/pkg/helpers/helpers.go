package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// RetriableApiRule wraps any function that modifies or creates an APIRule
type RetriableApiRule func(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error)

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

// Returns error if the status code is not in between bounds of status predicate after retrying deadline is reached
func (h *Helper) CallEndpointWithRetries(url string, predicate *StatusPredicate) error {
	err := h.withRetries(func() (*http.Response, error) {
		return h.client.Get(url)
	}, predicate.TestPredicate)

	if err != nil {
		return fmt.Errorf("error calling endpoint %s err=%s", url, err)
	}

	return nil
}

// Returns error if the status code is not in between bounds of status predicate after retrying deadline is reached
func (h *Helper) CallEndpointWithHeadersWithRetries(headerValue string, headerName, url string, predicate *StatusPredicate) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set(headerName, headerValue)
	err = h.withRetries(func() (*http.Response, error) {
		return h.client.Do(req)
	}, predicate.TestPredicate)

	if err != nil {
		return fmt.Errorf("error calling endpoint %s with %s=%s err=%s", url, headerName, headerValue, err)
	}

	return nil
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

// StatusPredicate is a struct representing desired endpoint call response status code, that is between LowerStatusBound and UpperStatusBound
type StatusPredicate struct {
	LowerStatusBound int
	UpperStatusBound int
}

func (s *StatusPredicate) TestPredicate(response *http.Response) bool {
	return response.StatusCode >= s.LowerStatusBound && response.StatusCode <= s.UpperStatusBound
}

// APIRuleWithRetries tries toExecute function and retries with onRetry if APIRule status is "ERROR"
func (h *Helper) APIRuleWithRetries(toExecute RetriableApiRule, onRetry RetriableApiRule, k8sClient dynamic.Interface, resources []unstructured.Unstructured) error {

	type status struct {
		Status struct {
			APIRuleStatus struct {
				Code string `json:"code"`
			} `json:"APIRuleStatus"`
		} `json:"status"`
	}
	res, err := toExecute(k8sClient, resources...)
	if err != nil {
		return err
	}

	js, err := json.Marshal(res)
	if err != nil {
		return err
	}

	apiStatus := status{}

	json.Unmarshal(js, &apiStatus)
	if apiStatus.Status.APIRuleStatus.Code == "ERROR" {
		return retry.Do(func() error {
			res, err := onRetry(k8sClient, resources...)
			if err != nil {
				return err
			}

			js, err := json.Marshal(res)
			if err != nil {
				return err
			}
			json.Unmarshal(js, &apiStatus)
			if apiStatus.Status.APIRuleStatus.Code == "ERROR" {
				return errors.New("APIRule status not ok")
			}
			return nil
		}, h.opts...)
	}
	return nil
}
