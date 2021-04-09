package helpers

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

const (
	timeout = time.Second * 30
	applicationJSON = "application/json"
)


// SendEvent sends event to the application gateway
func SendEvent(target, eventType, eventTypeVersion string) error {
	log.Printf("Sending an event to target: %s with eventType: %s, eventTypeVersion: %s", target, eventType, eventTypeVersion)
	payload := fmt.Sprintf(
		`{"event-type":"%s","event-type-version": "%s","event-time":"2018-11-02T22:08:41+00:00","data":"foo"}`, eventType, eventTypeVersion)
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: timeout,
		}).DialContext,
	}
	client := http.Client{Transport: transport}
	res, err := client.Post(target,
		applicationJSON,
		strings.NewReader(payload))
	if err != nil {
		return errors.Wrap(err, "HTTP POST request failed in SendEvent()")
	}

	if err := verifyStatusCode(res, 200); err != nil {
		return errors.Wrap(err, "HTTP POST request returned non-2xx failed in SendEvent()")
	}

	return nil
}

// CheckEvent checks whether the subscriber has received the event or not
func CheckEvent(target string, statusCode int, retryOptions ...retry.Option) error {
	return retry.Do(func() error {
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
		}
		client := http.Client{Transport: transport}
		res, err := client.Get(target)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("HTTP GET failed in CheckEvent() for target: %v", target))
		}

		if err := verifyStatusCode(res, statusCode); err != nil {
			return errors.Wrap(err, fmt.Sprintf("HTTP GET request returned non-200 in CheckEvent() for target: %v", target))
		}

		return nil
	}, retryOptions...)
}

// Verify that the http response has the given status code and return an error if not
func verifyStatusCode(res *http.Response, expectedStatusCode int) error {
	if res.StatusCode != expectedStatusCode {
		return fmt.Errorf("status code is wrong, have: %d, want: %d", res.StatusCode, expectedStatusCode)
	}
	return nil
}
