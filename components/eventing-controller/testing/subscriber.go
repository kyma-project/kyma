package testing

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/avast/retry-go/v3"
	pkgerrors "github.com/pkg/errors"
)

type Subscriber struct {
	addr          string
	server        *http.Server
	StoreEndpoint string
	CheckEndpoint string
	Return500Endpoint string
	CheckRetriesEndpoint string
}

func NewSubscriber(addr string) *Subscriber {
	return &Subscriber{
		addr:          addr,
		StoreEndpoint: "/store",
		CheckEndpoint: "/check",
		Return500Endpoint: "/return500",
		CheckRetriesEndpoint: "/check_retries",
	}
}

func (s *Subscriber) Start() {
	store := make(chan string, 5)
	retries := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("read data failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		store <- string(data)
	})
	mux.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/return500", func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("read data failed: %v", err)
		}
		retries++
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/check_retries", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(fmt.Sprintf("%d",retries)))
		if err != nil {
			log.Printf("check_retries failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		log.Printf("start subscriber %v", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("start subscriber failed: %v", err)
		}
	}()
}

func (s *Subscriber) Shutdown() {
	go func() {
		if err := s.server.Close(); err != nil {
			log.Printf("shutdown subscriber failed: %v", err)
		}
	}()
}

func (s Subscriber) CheckEvent(expectedData, subscriberCheckURL string) error {
	var body []byte
	maxAttempts := uint(5)
	delay := time.Second
	err := retry.Do(
		func() error {
			resp, err := http.Get(subscriberCheckURL) //nolint:gosec
			if err != nil {
				return pkgerrors.Wrapf(err, "get HTTP request failed")
			}
			if !is2XXStatusCode(resp.StatusCode) {
				return fmt.Errorf("response code is not 2xx, received response code is: %d", resp.StatusCode)
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

func (s Subscriber) CheckRetries(expectedData int, subscriberCheckRetriesURL string) error {
	var body []byte
	maxAttempts := uint(5)
	delay := time.Second
	err := retry.Do(
		func() error {
			resp, err := http.Get(subscriberCheckRetriesURL) //nolint:gosec
			if err != nil {
				return pkgerrors.Wrapf(err, "get HTTP request failed")
			}
			if !is2XXStatusCode(resp.StatusCode) {
				return fmt.Errorf("response code is not 2xx, received response code is: %d", resp.StatusCode)
			}
			defer func() { _ = resp.Body.Close() }()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return pkgerrors.Wrapf(err, "read data failed")
			}
			if string(body) != fmt.Sprintf("%d", expectedData){
				return fmt.Errorf("total retries not received")
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
	return nil
}

func is2XXStatusCode(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
