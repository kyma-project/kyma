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

	"go.uber.org/atomic"

	"github.com/avast/retry-go/v3"
	pkgerrors "github.com/pkg/errors"
)

type Subscriber struct {
	server           *httptest.Server
	SinkURL          string
	checkURL         string
	InternalErrorURL string
	checkRetriesURL  string
}

type SubscriberOption func(*Subscriber)

const (
	maxNoOfData           = 10
	maxAttempts           = 10
	storeEndpoint         = "/store"
	checkEndpoint         = "/check"
	internalErrorEndpoint = "/return500"
	checkRetriesEndpoint  = "/check_retries"
)

func getServeMux() *http.ServeMux {
	store := make(chan string, maxNoOfData)
	retries := atomic.Int32{}
	mux := http.NewServeMux()

	mux.HandleFunc(storeEndpoint, func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("read data failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		store <- string(data)
	})
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
	mux.HandleFunc(internalErrorEndpoint, func(w http.ResponseWriter, r *http.Request) {
		if data, err := ioutil.ReadAll(r.Body); err != nil {
			log.Printf("read data failed: %v", err)
		} else {
			store <- string(data)
		}
		retries.Inc()
		w.WriteHeader(http.StatusInternalServerError)
	})
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

func NewSubscriber(opts ...SubscriberOption) *Subscriber {
	subscriber := &Subscriber{}
	for _, opt := range opts {
		opt(subscriber)
	}

	if subscriber.server == nil {
		mux := getServeMux()
		subscriber.server = httptest.NewServer(mux)
	}

	subscriber.SinkURL = fmt.Sprintf("%s%s", subscriber.server.URL, storeEndpoint)
	subscriber.checkURL = fmt.Sprintf("%s%s", subscriber.server.URL, checkEndpoint)
	subscriber.InternalErrorURL = fmt.Sprintf("%s%s", subscriber.server.URL, internalErrorEndpoint)
	subscriber.checkRetriesURL = fmt.Sprintf("%s%s", subscriber.server.URL, checkRetriesEndpoint)
	return subscriber
}

func WithListener(listener net.Listener) SubscriberOption {
	return func(subscriber *Subscriber) {
		mux := getServeMux()
		subscriber.server = httptest.NewUnstartedServer(mux)
		subscriber.server.Listener.Close()
		subscriber.server.Listener = listener
		subscriber.server.Start()
	}
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

func (s Subscriber) CheckEvent(expectedData string) error {
	var body []byte
	delay := time.Second
	err := retry.Do(
		func() error {
			resp, err := http.Get(s.checkURL)
			if err != nil {
				return pkgerrors.Wrapf(err, "get HTTP request failed")
			}
			if !is2XXStatusCode(resp.StatusCode) {
				return fmt.Errorf("Response code is not 2xx, received Response code is: %d", resp.StatusCode)
			}
			defer func() { _ = resp.Body.Close() }()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return pkgerrors.Wrapf(err, "read data failed")
			}
			if string(body) != expectedData {
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

// CheckRetries checks if the number of retries specified by expectedData was done and that data sent on each retry was correctly received.
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
				return fmt.Errorf("Response code is not 2xx, received Response code is: %d", resp.StatusCode)
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
