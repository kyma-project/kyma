package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Subscribers Servers API paths
const (
	// v1 endpoints
	SubServer1EventsPath  = "/v1/events"
	SubServer1StatusPath  = "/v1/status"
	SubServer1ResultsPath = "/v1/results"

	// v2 endpoints
	SubServer2EventsPath  = "/v2/events"
	SubServer2StatusPath  = "/v2/status"
	SubServer2ResultsPath = "/v2/results"

	// v3 endpoints
	SubServer3EventsPath  = "/v3/events"
	SubServer3StatusPath  = "/v3/status"
	SubServer3ResultsPath = "/v3/results"
)

// NewSubscriberServerWithPort creates a new HTTP server with multiple endpoints acting as an event subscriber
func NewSubscriberServerWithPort(port int, stop chan bool) *http.Server {
	// multiplexer
	serveMux := http.NewServeMux()

	// v1 handlers
	serveMux.HandleFunc(SubServer1EventsPath, eventsHandlerV1)
	serveMux.HandleFunc(SubServer1StatusPath, statusHandler)
	serveMux.HandleFunc(SubServer1ResultsPath, resultsHandlerV1)
	// v2 handlers
	serveMux.HandleFunc(SubServer2EventsPath, eventsHandlerV2)
	serveMux.HandleFunc(SubServer2StatusPath, statusHandler)
	serveMux.HandleFunc(SubServer2ResultsPath, resultsHandlerV2)
	// v3 handlers
	serveMux.HandleFunc(SubServer3EventsPath, eventsHandlerV3)
	serveMux.HandleFunc(SubServer3StatusPath, statusHandler)
	serveMux.HandleFunc(SubServer3ResultsPath, resultsHandlerV3)
	// shutdown handler
	serveMux.Handle("/shutdown", shutdownHandler(stop))

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: Logger(serveMux)}

	// start listener and serve requests
	go func() {
		log.Printf("Subscriber HTTP server starting on port %d", port)
		log.Fatal(server.ListenAndServe())
	}()

	return server
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
	_ = json.NewEncoder(w).Encode(subscriberV1Result)
}

func eventsHandlerV1(_ http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	_ = json.NewDecoder(r.Body).Decode(&subscriberV1Result)
}

func resultsHandlerV2(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	_ = json.NewEncoder(w).Encode(subscriberV2Result)
}

func eventsHandlerV2(_ http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	_ = json.NewDecoder(r.Body).Decode(&subscriberV2Result)
}

func resultsHandlerV3(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	_ = json.NewEncoder(w).Encode(subscriberV3Result)
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
