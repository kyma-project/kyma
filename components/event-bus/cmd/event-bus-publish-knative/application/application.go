package application

import (
	"net/http"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/handlers"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/publisher"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/publish/opts"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
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
	app.registerPublishHandler()
}

// ServeMux encapsulates an http.ServeMux
func (app *KnativePublishApplication) ServeMux() *http.ServeMux {
	return app.serveMux
}

func (app *KnativePublishApplication) registerReadinessProbe() {
	app.serveMux.HandleFunc("/v1/status/ready", handlers.ReadinessProbeHandler())
}

func (app *KnativePublishApplication) registerPublishHandler() {
	knativePublishHandler := handlers.KnativePublishHandler(app.knativeLib, app.knativePublisher, app.tracer, app.options)
	requestSizeLimitHandler := handlers.WithRequestSizeLimiting(knativePublishHandler, app.options.MaxRequestSize)
	app.serveMux.HandleFunc("/v1/events", requestSizeLimitHandler)
}
