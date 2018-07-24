package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-sv/application"
	"github.com/kyma-project/kyma/components/event-bus/internal/sv/opts"
)

func main() {
	svOpts := opts.ParseFlags()

	svApplication := application.NewSubscriptionValidatorApplication(svOpts)
	defer svApplication.Stop()

	log.Printf("HTTP server starting on port %v", svOpts.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", svOpts.Port), svApplication.ServerMux))
}
