package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// NewSubscriberServerWithPort ...
func NewSubscriberServerWithPort(port int, stop chan bool) *http.Server {
	subscriberMux := http.NewServeMux()
	subscriberMux.Handle("/events", getEventsHandler())
	subscriberMux.Handle("/status", getStatusHandler())
	subscriberMux.Handle("/results", getResultsHandler())
	subscriberMux.Handle("/shutdown", getShutdownHandler(stop))

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: Logger(subscriberMux)}

	// start listener and serve requests
	go func() {
		log.Printf("Subscriber HTTP server starting on port %d", port)
		log.Fatal(server.ListenAndServe())
	}()

	return server
}

func getEventsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var result string
		_ = json.NewDecoder(r.Body).Decode(&result)
	})
}

func getStatusHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func getResultsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var result string
		_ = json.NewEncoder(w).Encode(result)
	})
}

func getShutdownHandler(stop chan bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		stop <- true
	})
}
