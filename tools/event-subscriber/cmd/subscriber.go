package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type SubscriberState struct {
	counter     uint32
	receivedCEs map[string]cloudevents.Event
	mu          sync.Mutex
}

type counterResponse struct {
	Counter int `json:"counter"`
}

func main() {
	port := flag.Int("port", 9000, "tcp port on which to listen for http requests")
	flag.Parse()

	state := SubscriberState{}
	router := state.setupRouter()

	log.Printf("will listen on :%v\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", *port), router); err != nil {
		log.Fatalf("unable to start http server, %s", err)
	}
}

func (s *SubscriberState) setupRouter() *mux.Router {
	s.receivedCEs = make(map[string]cloudevents.Event)
	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	cehandler, err := cloudevents.NewHTTPReceiveHandler(ctx, p, s.receiveCE)
	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	router.Handle("/ce", cehandler).Methods("POST")
	router.HandleFunc("/", s.increaseCounter).Methods("POST")

	router.HandleFunc("/ce/{source}/{type}/{version}", s.checkCEBySourceTypeVersion).Methods("GET")
	router.HandleFunc("/ce/by-uuid/{uuid}", s.checkCEbyUUID).Methods("GET")
	router.HandleFunc("/ce", s.getAllCE).Methods("GET")
	router.HandleFunc("/", s.checkCounter).Methods("GET")

	router.HandleFunc("/", s.reset).Methods("DELETE")
	return router
}

func (s *SubscriberState) receiveCE(ctx context.Context, event cloudevents.Event) protocol.Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := event.Context.GetID()
	if id == "" {
		return cehttp.NewResult(http.StatusBadRequest, "ID missing")
	}
	if event.Extensions()["eventtypeversion"] == nil {
		return cehttp.NewResult(http.StatusBadRequest, "event-type-version missing")
	}
	log.Infof("Received CE: %v", id)
	s.receivedCEs[id] = event
	return nil
}

func (s *SubscriberState) increaseCounter(_ http.ResponseWriter, _ *http.Request) {
	atomic.AddUint32(&s.counter, 1)
	log.Infof("Received Request: counter = %v", s.counter)
}

func (s *SubscriberState) checkCEBySourceTypeVersion(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vars := mux.Vars(r)
	eventsource := vars["source"]
	eventtype := vars["type"]
	eventversion := vars["version"]

	events := make([]cloudevents.Event, 0)
	for _, event := range s.receivedCEs {
		if event.Source() == eventsource &&
			event.Type() == eventtype &&
			event.Extensions()["eventtypeversion"] == eventversion {
			events = append(events, event)
		}
	}
	log.Infof("Checking for source: %v, type: %v, version: %v :: found: %v", eventsource, eventtype, eventversion, events)
	if len(events) == 0 {
		w.WriteHeader(http.StatusNoContent)
	}
	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Errorf("Error during checkCEbySourceTypeVersion: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *SubscriberState) getAllCE(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	events := make([]cloudevents.Event, 0)
	for _, event := range s.receivedCEs {
		events = append(events, event)
	}
	log.Infof("Getting all CE: %v", events)

	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Errorf("error during getAllCE: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *SubscriberState) checkCEbyUUID(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	ce, exists := s.receivedCEs[uuid]
	if !exists {
		log.Infof("Checking for uuid: %v. Not found", uuid)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	log.Infof("Checking for uuid: %v. found: %v", uuid, ce)
	if err := json.NewEncoder(w).Encode(ce); err != nil {
		log.Errorf("error during checkCEbyUUID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *SubscriberState) checkCounter(w http.ResponseWriter, r *http.Request) {
	response := counterResponse{Counter: int(s.counter)}
	log.Infof("Checking counter: %v", response)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Errorf("error during checkCounter: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *SubscriberState) reset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Info("Reset")
	s.receivedCEs = make(map[string]cloudevents.Event)
	atomic.StoreUint32(&s.counter, 0)
}
