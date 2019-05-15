package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
)

// Subscribers Servers API paths
const (
	SubServer1EventsPath  = "/v1/events"
	SubServer1StatusPath  = "/v1/status"
	SubServer1ResultsPath = "/v1/results"
	SubServer2EventsPath  = "/v2/events"
	SubServer2StatusPath  = "/v2/status"
	SubServer2ResultsPath = "/v2/results"
	SubServer3EventsPath  = "/v3/events"
	SubServer3StatusPath  = "/v3/status"
	SubServer3ResultsPath = "/v3/results"
)

// NewSubscriberServerV1 ...
func NewSubscriberServerV1() *httptest.Server {
	subscriberMux := http.NewServeMux()
	subscriberMux.HandleFunc(SubServer1EventsPath, eventsHandlerV1)
	subscriberMux.HandleFunc(SubServer1StatusPath, statusHandler)
	subscriberMux.HandleFunc(SubServer1ResultsPath, resultsHandlerV1)
	return httptest.NewServer(Logger(subscriberMux))
}

// NewSubscriberServerV2 ...
func NewSubscriberServerV2() *httptest.Server {
	subscriberMux := http.NewServeMux()
	subscriberMux.HandleFunc(SubServer2EventsPath, eventsHandlerV2)
	subscriberMux.HandleFunc(SubServer2StatusPath, statusHandler)
	subscriberMux.HandleFunc(SubServer2ResultsPath, resultsHandlerV2)
	return httptest.NewServer(Logger(subscriberMux))
}

// NewSubscriberServerWithPort ...
func NewSubscriberServerWithPort(port int, stop chan bool) *http.Server {
	subscriberMux := http.NewServeMux()
	subscriberMux.HandleFunc("/v1/events", eventsHandlerV1)
	subscriberMux.HandleFunc("/v1/status", statusHandler)
	subscriberMux.HandleFunc("/v1/results", resultsHandlerV1)
	subscriberMux.HandleFunc(SubServer3EventsPath, eventsHandlerV3)
	subscriberMux.HandleFunc(SubServer3StatusPath, statusHandler)
	subscriberMux.HandleFunc(SubServer3ResultsPath, resultsHandlerV3)
	subscriberMux.Handle("/shutdown", shutdownHandler(stop))

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: Logger(subscriberMux)}

	// start listener and serve requests
	go func() {
		log.Printf("Subscriber HTTP server starting on port %d", port)
		log.Fatal(srv.ListenAndServe())
	}()
	return srv
}

var (
	subscriberV1Result string
	subscriberV2Result string
	subscriberV3Result map[string][]string
	mu                 sync.Mutex
)

func resultsHandlerV1(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(subscriberV1Result)
}

func eventsHandlerV1(_ http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewDecoder(r.Body).Decode(&subscriberV1Result)
}

func resultsHandlerV2(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(subscriberV2Result)
}

func eventsHandlerV2(_ http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewDecoder(r.Body).Decode(&subscriberV2Result)
}

func resultsHandlerV3(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(subscriberV3Result)
}

func eventsHandlerV3(_ http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	subscriberV3Result = r.Header
}

func statusHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func shutdownHandler(stop chan bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		stop <- true
	})
}
