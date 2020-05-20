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
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	counter     uint32
	receivedCEs map[string]cloudevents.Event
	mu          sync.Mutex
)

type counterResponse struct {
	Counter int `json:"counter"`
}

func main() {
	port := flag.Int("port", 9000, "tcp port on which to listen for http requests")
	flag.Parse()

	receivedCEs = make(map[string]cloudevents.Event)

	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	cehandler, err := cloudevents.NewHTTPReceiveHandler(ctx, p, receiveCE)
	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}

	// Use a gorilla mux implementation for the overall http handler.
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	router.Handle("/ce", cehandler).Methods("POST")
	router.HandleFunc("/", increaseCounter).Methods("POST")

	router.HandleFunc("/ce/{source}/{type}/{version}", checkCEBySourceTypeVersion).Methods("GET")
	router.HandleFunc("/ce/by-uuid/{uuid}", checkCEbyUUID).Methods("GET")
	router.HandleFunc("/ce", getAllCE).Methods("GET")
	router.HandleFunc("/", checkCounter).Methods("GET")

	router.HandleFunc("/", reset).Methods("DELETE")

	log.Printf("will listen on :%v\n", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", *port), router); err != nil {
		log.Fatalf("unable to start http server, %s", err)
	}
}

func receiveCE(_ context.Context, event cloudevents.Event) {
	mu.Lock()
	defer mu.Unlock()
	id := event.Context.GetID()
	log.Infof("Received CE: %v", id)
	receivedCEs[id] = event

}

func increaseCounter(_ http.ResponseWriter, _ *http.Request) {
	atomic.AddUint32(&counter, 1)
	log.Infof("Received Request: counter = %v", counter)
}

func checkCEBySourceTypeVersion(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	vars := mux.Vars(r)
	eventsource := vars["source"]
	eventtype := vars["type"]
	eventversion := vars["version"]

	events := make([]cloudevents.Event, 0)
	for _, event := range receivedCEs {
		if event.Source() == eventsource &&
			event.Type() == eventtype &&
			event.Extensions()["eventtypeversion"] == eventversion {
			events = append(events, event)
		}
	}
	log.Infof("Checking for source: %v, type: %v, version: %v  :: found: %v", eventsource, eventtype, eventversion, events)
	if len(events) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Errorf("Error during checkCEbySourceTypeVersion: %v", err)
	}
}

func getAllCE(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	events := make([]cloudevents.Event, 0)
	for _, event := range receivedCEs {
		events = append(events, event)
	}
	log.Infof("Getting all CE: %v", events)

	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Errorf("Error during getAllCE: %v", err)
	}
}

func checkCEbyUUID(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	ce, exists := receivedCEs[uuid]
	if !exists {
		log.Infof("Checking for uuid: %v. found: []", uuid)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	log.Infof("Checking for uuid: %v. found: %v", uuid, ce)
	if err := json.NewEncoder(w).Encode(ce); err != nil {
		log.Errorf("Error during checkCEbyUUID: %v", err)
	}
}

func checkCounter(w http.ResponseWriter, r *http.Request) {
	response := counterResponse{Counter: int(counter)}
	log.Infof("Checking counter: %v", response)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Errorf("Error during checkCounter: %v", err)
	}
}

func reset(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	log.Info("Reset")
	receivedCEs = make(map[string]cloudevents.Event)
	atomic.StoreUint32(&counter, 0)
}
