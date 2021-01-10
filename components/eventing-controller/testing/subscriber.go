package testing

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/avast/retry-go"
	pkgerrors "github.com/pkg/errors"
)

type Subscriber struct {
	addr   string
	t      *testing.T
	server *http.Server
}

func NewSubscriber(addr string, t *testing.T) *Subscriber {
	return &Subscriber{
		addr: addr,
		t:    t,
	}
}

func (s *Subscriber) Start() {
	store := ""
	mux := http.NewServeMux()
	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.t.Fatalf("failed to read data: %v", err)
		}
		store = string(data)
	})
	mux.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(store))
		if err != nil {
			s.t.Fatalf("failed to write to the response: %v", err)
		}
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
				s.t.Fatalf("failed to start server: %v", err)
			}
		}
	}()
}

func (s *Subscriber) Shutdown() {
	go func() {
		err := s.server.Close()
		if err != nil {
			s.t.Errorf("failed to shutdown Subscriber: %v", err)
		}
		log.Print("subscriber server shut down was successful")
	}()
}

func (s Subscriber) CheckEvent(expectedData, subscriberCheckURL string) error {
	var body []byte
	maxAttempts := uint(5)
	err := retry.Do(
		func() error {
			resp, err := http.Get(subscriberCheckURL)
			if err != nil {
				return err
			}
			defer func() { _ = resp.Body.Close() }()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			if string(body) != expectedData {
				return fmt.Errorf("subscriber did not get the event with data: \"%s\" yet...waiting", expectedData)
			}
			return nil
		},
		retry.Delay(2*time.Second),
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
