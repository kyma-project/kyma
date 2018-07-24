package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/application"
	"github.com/kyma-project/kyma/components/event-bus/internal/publish"
)

const (
	allowedIDChars = `^[a-zA-Z0-9_\-]+$`
)

var (
	isValidID = regexp.MustCompile(allowedIDChars).MatchString
	sema      chan struct{}
)

func main() {
	log.Println("Publish :: Starting up")
	publishOpts := publish.ParseFlags()

	startPublish(publishOpts)
}

func startPublish(publishOpts *publish.Options) {
	if !isValidID(publishOpts.ClientID) {
		log.Fatal("invalid client_id ", publishOpts.ClientID)
	}
	// enforce a limit of concurrent requests processed in parallel.
	sema = make(chan struct{}, publishOpts.NoOfConcurrentRequests)

	publishApplication := application.NewPublishApplication(publishOpts)
	defer publishApplication.Stop()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	rtr := limitParallelRequests(publishApplication.ServerMux)
	if publishOpts.TraceRequests {
		logger(rtr)
	}

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(publishOpts.Port),
		Handler: rtr,
	}
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Println("Got os interrupt...")
	case syscall.SIGTERM:
		log.Println("Got SIGTERM")
	}

	log.Println("The service is shutting down....")
	srv.Shutdown(context.Background())
	log.Println("Done..")
}

func limitParallelRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{} // acquire a semaphore
		defer func() { <-sema }()
		h.ServeHTTP(w, r)
	})
}

func logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s requested %s", r.RemoteAddr, r.URL)
		dump, err := httputil.DumpRequest(r, true)
		log.Printf("%q", dump)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		h.ServeHTTP(w, r)
	})
}
