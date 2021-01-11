package testing

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	pkgerrors "github.com/pkg/errors"
)

type Subscriber struct {
	addr   string
	server *http.Server
}

func NewSubscriber(addr string) *Subscriber {
	return &Subscriber{
		addr: addr,
	}
}

func (s *Subscriber) Start() {
	store := ""
	mux := http.NewServeMux()
	mux.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			msg := fmt.Sprintf("failed to read data: %v", err)
			log.Printf(msg)
			_, writeErr := w.Write([]byte(msg))
			if writeErr != nil {
				log.Printf("failed to write data for /store: %v", writeErr)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		store = string(data)
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(store))
		if err != nil {
			_, writeErr := w.Write([]byte("failed to write to the response"))
			if writeErr != nil {
				log.Printf("failed to write data for /check: %v", writeErr)
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		log.Printf("subscriber server is starting at %v", s.server.Addr)
		err := s.server.ListenAndServe()
		if err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("failed to start server: %v", err)
			}
		}
	}()
}

func (s *Subscriber) Shutdown() {
	go func() {
		err := s.server.Close()
		if err != nil {
			log.Printf("failed to shutdown Subscriber: %v", err)
		}
		log.Print("subscriber server shut down was successful")
	}()
}

func (s Subscriber) CheckEvent(expectedData, subscriberCheckURL string) error {
	var body []byte
	maxAttempts := uint(5)
	delay := time.Second
	err := retry.Do(
		func() error {
			resp, err := http.Get(subscriberCheckURL)
			if err != nil {
				return pkgerrors.Wrapf(err, "failed to HTTP GET")
			}
			if !is2XXStatusCode(resp.StatusCode) {
				return fmt.Errorf("response code is not 2xx, received response code is: %d", resp.StatusCode)
			}
			defer func() { _ = resp.Body.Close() }()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return pkgerrors.Wrapf(err, "failed to read data")
			}

			if string(body) != expectedData {
				return fmt.Errorf("subscriber did not get the event with data: \"%s\" yet...waiting", expectedData)
			}
			return nil
		},
		retry.Delay(delay),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(maxAttempts),
		retry.OnRetry(func(n uint, err error) { log.Printf("[%v] try failed: %s", n, err) }),
	)
	if err != nil {
		return pkgerrors.Wrapf(err, "failed to check the event after retries")
	}

	log.Printf("event :%s received successfully", expectedData)

	return nil
}

func is2XXStatusCode(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
