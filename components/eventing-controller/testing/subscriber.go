package testing

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/testing/event/cehelper"
	"go.uber.org/atomic"

	"github.com/avast/retry-go/v3"
	pkgerrors "github.com/pkg/errors"
)

const (
	maxNoOfData           = 10
	maxAttempts           = 10
	storeEndpoint         = "/store"
	checkEndpoint         = "/check"
	internalErrorEndpoint = "/return500"
	checkRetriesEndpoint  = "/check_retries"
)

type Subscriber struct {
	server           *httptest.Server
	SinkURL          string
	checkURL         string
	InternalErrorURL string
	checkRetriesURL  string
}

type SubscriberOption func(*Subscriber)

// NewSubscriber creates a simple test Subscriber with http Endpoints to received, store end answer
// event data. NewSubscriber accepts SubscriberOpts for customization.
func NewSubscriber(opts ...SubscriberOption) *Subscriber {
	subscriber := &Subscriber{}
	for _, opt := range opts {
		opt(subscriber)
	}

	if subscriber.server == nil {
		mux := getDataServeMux()
		subscriber.server = httptest.NewServer(mux)
	}
	subscriber.SinkURL = fmt.Sprintf("%s%s", subscriber.server.URL, storeEndpoint)
	subscriber.checkURL = fmt.Sprintf("%s%s", subscriber.server.URL, checkEndpoint)
	subscriber.InternalErrorURL = fmt.Sprintf("%s%s", subscriber.server.URL, internalErrorEndpoint)
	subscriber.checkRetriesURL = fmt.Sprintf("%s%s", subscriber.server.URL, checkRetriesEndpoint)
	return subscriber
}

// WithCloudEventServeMux will make the Subscriber store all CloudEvent-related date ("ce-..." header etc) at its
// "/store" e Endpoint instead of its default behaviour of only storing the http.Request.Body data.
func WithCloudEventServeMux() SubscriberOption {
	return func(subscriber *Subscriber) {
		mux := getCloudEventServeMux()
		subscriber.server = httptest.NewServer(mux)
	}
}

func WithListener(listener net.Listener) SubscriberOption {
	return func(subscriber *Subscriber) {
		mux := getDataServeMux()
		subscriber.server = httptest.NewUnstartedServer(mux)
		subscriber.server.Listener.Close()
		subscriber.server.Listener = listener
		subscriber.server.Start()
	}
}

// getCloudEventServeMux sets the Subscriber up to handle all CloudEvent related data (headers etc.) to test
// against CloudEvents. Use the WithCloudEventServeMux opt to set the Subscriber with this ServeMux.
func getCloudEventServeMux() *http.ServeMux {
	store := make(chan string, maxNoOfData)
	retries := atomic.Int32{}
	mux := http.NewServeMux()

	// this Endpoint stores the CloudEvent as a string.
	mux.HandleFunc(storeEndpoint, func(w http.ResponseWriter, r *http.Request) {
		eventString, err := cehelper.RequestToEventString(r)
		if err != nil {
			log.Printf("read data failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		store <- eventString
	})
	// this Endpoint returns the last stored string. If, after a time out of 0.5 sec, nothing was found in the store
	// it will return an empty string.
	mux.HandleFunc(checkEndpoint, func(w http.ResponseWriter, r *http.Request) {
		var msg string
		select {
		case m := <-store:
			msg = m
		case <-time.After(500 * time.Millisecond):
			msg = ""
		}
		_, err := w.Write([]byte(msg))
		if err != nil {
			log.Printf("write data failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	// this Endpoint stores the CloudEvent as a string while returning an "Internal Server Error" (code: 500).
	mux.HandleFunc(internalErrorEndpoint, func(w http.ResponseWriter, r *http.Request) {
		if eventString, err := cehelper.RequestToEventString(r); err != nil {
			log.Printf("read data failed: %v", err)
		} else {
			store <- eventString
		}
		retries.Inc()
		w.WriteHeader(http.StatusInternalServerError)
	})
	// this Endpoint returns the number of attempted retries.
	mux.HandleFunc(checkRetriesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(fmt.Sprintf("%d", retries.Load())))
		if err != nil {
			log.Printf("check_retries failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return mux
}

// getDataServeMux sets the Subscriber up to handle the request body data for tests. This is the default ServeMux for
// the Subscriber. Use this to test against only the event data without any metadata.
func getDataServeMux() *http.ServeMux {
	store := make(chan string, maxNoOfData)
	retries := atomic.Int32{}
	mux := http.NewServeMux()

	// this Endpoint stores the data of the request body.
	mux.HandleFunc(storeEndpoint, func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("read data failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		store <- string(data)
	})
	// this Endpoint returns the last stored string. If, after a time out of 0.5 sec, nothing was found in the store
	// it will return an empty string.
	mux.HandleFunc(checkEndpoint, func(w http.ResponseWriter, r *http.Request) {
		var msg string
		select {
		case m := <-store:
			msg = m
		case <-time.After(500 * time.Millisecond):
			msg = ""
		}
		_, err := w.Write([]byte(msg))
		if err != nil {
			log.Printf("write data failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	// this Endpoint stores the request body as a string while returning an "Internal Server Error" (code: 500).
	mux.HandleFunc(internalErrorEndpoint, func(w http.ResponseWriter, r *http.Request) {
		if data, err := ioutil.ReadAll(r.Body); err != nil {
			log.Printf("read data failed: %v", err)
		} else {
			store <- string(data)
		}
		retries.Inc()
		w.WriteHeader(http.StatusInternalServerError)
	})
	// this Endpoint returns the number of attempted retries.
	mux.HandleFunc(checkRetriesEndpoint, func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(fmt.Sprintf("%d", retries.Load())))
		if err != nil {
			log.Printf("check_retries failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	return mux
}

func (s *Subscriber) Shutdown() {
	s.server.Close()
}

func (s *Subscriber) GetSubscriberListener() net.Listener {
	return s.server.Listener
}

func (s *Subscriber) IsRunning() bool {
	return s.CheckEvent("") == nil
}

// CheckEvent checks if the Subscriber received the expected string and will return an error if this is not the case.
func (s Subscriber) CheckEvent(expectedData string) error {
	var body []byte
	delay := time.Second
	err := retry.Do(
		func() error {
			// check if a response was received and that it's code is in 2xx-range
			resp, err := http.Get(s.checkURL)
			if err != nil {
				return pkgerrors.Wrapf(err, "get HTTP request failed")
			}
			if !is2XXStatusCode(resp.StatusCode) {
				return fmt.Errorf("expected resonse code 2xx, actual response code: %d", resp.StatusCode)
			}

			// try to read the response body
			defer func() { _ = resp.Body.Close() }()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return pkgerrors.Wrapf(err, "read data failed")
			}

			// compare response body with expectations
			if expectedData != string(body) {
				return fmt.Errorf("event not received")
			}
			return nil
		},
		retry.Delay(delay),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(maxAttempts),
		retry.OnRetry(func(n uint, err error) { log.Printf("[%v] try failed: %s", n, err) }),
	)
	if err != nil {
		return pkgerrors.Wrapf(err, "check event after retries failed")
	}

	log.Print("event received")
	return nil
}

// CheckRetries checks if the number of retries specified by expectedData was done and if the data sent on each retry
// was correctly received.
func (s Subscriber) CheckRetries(expectedNoOfRetries int, expectedData string) error {
	var body []byte
	delay := time.Second
	err := retry.Do(
		func() error {
			resp, err := http.Get(s.checkRetriesURL)
			if err != nil {
				return pkgerrors.Wrapf(err, "get HTTP request failed")
			}
			if !is2XXStatusCode(resp.StatusCode) {
				return fmt.Errorf("expected resonse code 2xx, actual response code: %d", resp.StatusCode)
			}
			defer func() { _ = resp.Body.Close() }()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return pkgerrors.Wrapf(err, "read data failed")
			}
			actualRetires, err := strconv.Atoi(string(body))
			if err != nil {
				return pkgerrors.Wrapf(err, "read data failed")
			}
			if actualRetires < expectedNoOfRetries {
				return fmt.Errorf("number of retries do not match (actualRetires=%d, expectedRetries=%d)", actualRetires, expectedNoOfRetries)
			}
			return nil
		},
		retry.Delay(delay),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(maxAttempts),
		retry.OnRetry(func(n uint, err error) { log.Printf("[%v] try failed: %s", n, err) }),
	)
	if err != nil {
		return pkgerrors.Wrapf(err, "check event after retries failed")
	}
	// test if 'expectedData' was received exactly 'expectedNoOfRetries' times
	for i := 1; i < expectedNoOfRetries; i++ {
		if err := s.CheckEvent(expectedData); err != nil {
			return pkgerrors.Wrapf(err, "check received data after retries failed")
		}
	}
	// OK
	return nil
}

func is2XXStatusCode(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
