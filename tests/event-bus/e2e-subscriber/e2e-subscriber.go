package main

import (
	"context"
	"flag"
	"log"

	"github.com/kyma-project/kyma/components/event-bus/test/util"
)

func main() {

	port := flag.Int("port", 9000, "tcp port on which to listen for http requests")
	flag.Parse()

	// creates the subscription server
	stopSubscriber := make(chan bool)
	log.Println("Creates the subscription server")
	subscriberServer := util.NewSubscriberServerWithPort(*port, stopSubscriber)

	<-stopSubscriber
	log.Println("Shutting down server...")
	subscriberServer.Shutdown(context.Background())
}
