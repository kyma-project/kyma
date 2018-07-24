package application

import (
	"log"
	"net/http"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/controllers"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/handlers"
	"github.com/kyma-project/kyma/components/event-bus/internal/publish"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
)

type PublishApplication struct {
	publisher controllers.Publisher
	tracer    trace.Tracer
	ServerMux *http.ServeMux
}

func NewPublishApplication(publishOpts *publish.Options) *PublishApplication {
	log.Println("Publish :: Initializing NATS Streaming publisher")
	publisher := controllers.GetPublisher(publishOpts.ClientID, publishOpts.NatsURL, publishOpts.NatsStreamingClusterID)
	err := publisher.Start()
	if err != nil {
		log.Fatalf("Error while initializing NATS Streaming publisher. %v", err)
	}

	log.Println("Publish :: Initializing tracer")
	traceOpts := trace.Options{
		APIURL:        publishOpts.TraceAPIURL,
		HostPort:      publishOpts.TraceHostPort,
		ServiceName:   publishOpts.ServiceName,
		OperationName: publishOpts.OperationName,
		Debug:         publishOpts.TraceDebug,
	}

	tracer := trace.StartNewTracer(&traceOpts)

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/v1/events", handlers.GetPublishHandler(&publisher, &tracer))
	serveMux.HandleFunc("/v1/status/ready", handlers.GetReadinessHandler(&publisher))

	return &PublishApplication{
		publisher: publisher,
		tracer:    tracer,
		ServerMux: serveMux,
	}
}

func (app *PublishApplication) Stop() {
	app.publisher.Stop()
	app.tracer.Stop()
}
