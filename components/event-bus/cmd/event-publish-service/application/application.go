package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	cloudeventstransport "github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/handlers"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/publisher"
	constants "github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/util"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/publish/opts"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
)

const (
	//APIV1 API V1 pattern
	APIV1 = "/v1/events"
	//APIV2 API V2 pattern
	APIV2 = "/v2/events"
)

// KnativePublishApplication represents a Knative PublishApplication
type KnativePublishApplication struct {
	started          bool
	options          *opts.Options
	serveMux         *http.ServeMux
	tracer           *trace.Tracer
	knativePublisher *publisher.KnativePublisher
	knativeLib       *knative.KnativeLib
}

// StartKnativePublishApplication starts a new KnativePublishApplication instance.
func StartKnativePublishApplication(options *opts.Options, knativeLib *knative.KnativeLib, knativePublisher *publisher.KnativePublisher, tracer *trace.Tracer) *KnativePublishApplication {
	// init and start the knative publish application
	application := newKnativePublishApplication(options, knativeLib, knativePublisher, tracer)
	application.start()
	return application
}

func newKnativePublishApplication(options *opts.Options, knativeLib *knative.KnativeLib, knativePublisher *publisher.KnativePublisher, tracer *trace.Tracer) *KnativePublishApplication {
	application := &KnativePublishApplication{
		started:          false,
		options:          options,
		serveMux:         http.NewServeMux(),
		tracer:           tracer,
		knativePublisher: knativePublisher,
		knativeLib:       knativeLib,
	}
	return application
}

func (app *KnativePublishApplication) start() {
	// the app is already started before
	if app.started {
		return
	}

	// mark the app as started and register the readiness and the publish handlers
	app.started = true
	app.registerReadinessProbe()
	app.registerPublishV1Handler()
	app.registerPublishV2Handler()
}

// ServeMux encapsulates an http.ServeMux
func (app *KnativePublishApplication) ServeMux() *http.ServeMux {
	return app.serveMux
}

func (app *KnativePublishApplication) registerReadinessProbe() {
	app.serveMux.HandleFunc("/v1/status/ready", handlers.ReadinessProbeHandler())
}

func (app *KnativePublishApplication) registerPublishV1Handler() {
	knativePublishHandler := handlers.KnativePublishHandler(constants.EventAPIV1, app.knativeLib, app.knativePublisher, app.tracer, app.options)
	requestSizeLimitHandler := handlers.WithRequestSizeLimiting(knativePublishHandler, app.options.MaxRequestSize)
	app.serveMux.HandleFunc(APIV1, requestSizeLimitHandler)
}

func (app *KnativePublishApplication) registerPublishV2Handler() {
	t, err := cloudevents.NewHTTPTransport()
	// TODO(nachtmaar):
	if err != nil {
		return
	}
	//TODO: set the logic here
	t.SetReceiver(cloudeventstransport.ReceiveFunc(app.HandleEvent))

	requestSizeLimitHandler := handlers.WithRequestSizeLimiting(t.ServeHTTP, app.options.MaxRequestSize)
	app.serveMux.HandleFunc(APIV2, requestSizeLimitHandler)
}

// Receive finally handles the decoded event
func (app *KnativePublishApplication) HandleEvent(ctx context.Context, event cloudevents.Event, eventResponse *cloudevents.EventResponse) error {
	fmt.Printf("received event %+v", event)

	//traceHeaders := getTraceHeaders(ctx)

	//bus.SendEventV2(event, *traceHeaders)
	//downgradedEvent := cloudevents.Event{
	//	// TODO(nachtmaar) dont' downgrade to CE v0.3 anymore if knative supports CE v1.0
	//	Context:     event.Context.AsV03(),
	//	Data:        event.Data,
	//	DataEncoded: event.DataEncoded,
	//	DataBinary:  event.DataBinary,
	//}
	if _, err := event.Context.GetExtension("event-type-version"); err != nil {
		// TODO(nachtmaar): set proper status code
		//eventResponse.Error(400, "we need a valid günther")
		return errors.New("günther")
	}

	return nil
}
