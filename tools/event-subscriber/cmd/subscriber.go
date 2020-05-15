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

	router.HandleFunc("/ce/{uuid}", checkCE).Methods("GET")
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

func checkCE(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	log.Infof("Checking for uuid: %v. found: %v", uuid, receivedCEs[uuid])
	cloudevents.Client
	if err := json.NewEncoder(w).Encode(receivedCEs[uuid]); err != nil {
		log.Errorf("Error during checkCE: %v", err)
	}
}

func checkCounter(w http.ResponseWriter, r *http.Request) {
	log.Infof("Checking counter = %v", counter)
	if err := json.NewEncoder(w).Encode(counter); err != nil {
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
