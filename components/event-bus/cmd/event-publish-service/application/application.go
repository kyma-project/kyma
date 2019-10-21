package application

import (
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/handlers"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/publisher"
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
	// TODO(nachtmaar): move to v1 package, e.g. v1handlers.KnativePublishHandler
	knativePublishHandler := handlers.KnativePublishHandler(app.knativeLib, app.knativePublisher, app.tracer, app.options)
	requestSizeLimitHandler := handlers.WithRequestSizeLimiting(knativePublishHandler, app.options.MaxRequestSize)
	app.serveMux.HandleFunc(APIV1, requestSizeLimitHandler)
}

func (app *KnativePublishApplication) registerPublishV2Handler() {

	t, err := cloudevents.NewHTTPTransport()
	// TODO(nachtmaar):
	if err != nil {
		return
	}
	handler := handlers.CloudEventHandler{
		KnativePublisher: app.knativePublisher,
		KnativeLib:       app.knativeLib,
		Transport:        t,
		Tracer:           app.tracer,
	}

	requestSizeLimitHandler := handlers.WithRequestSizeLimiting(handler.ServeHTTP, app.options.MaxRequestSize)
	app.serveMux.HandleFunc(APIV2, requestSizeLimitHandler)
}
