package application

import (
	"log"
	"net/http"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/handlers"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/publisher"
	constants "github.com/kyma-project/kyma/components/event-bus/cmd/event-publish-service/util"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/publish/opts"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"

	messagingv1alpha1client "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
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
	if app.knativeLib == nil {
		log.Fatalf("Cannot start without knativeLib!")
	}

	app.registerReadinessProbe(app.knativeLib.MsgChannelClient())
	app.registerPublishV1Handler()
	app.registerPublishV2Handler()
}

// ServeMux encapsulates an http.ServeMux
func (app *KnativePublishApplication) ServeMux() *http.ServeMux {
	return app.serveMux
}

func (app *KnativePublishApplication) registerReadinessProbe(msgClientInf messagingv1alpha1client.MessagingV1alpha1Interface) {
	app.serveMux.HandleFunc("/v1/status/ready", handlers.ReadinessProbeHandler(msgClientInf))
}

func (app *KnativePublishApplication) registerPublishV1Handler() {
	knativePublishHandler := handlers.KnativePublishHandler(constants.EventAPIV1, app.knativeLib, app.knativePublisher, app.tracer, app.options)
	requestSizeLimitHandler := handlers.WithRequestSizeLimiting(knativePublishHandler, app.options.MaxRequestSize)
	app.serveMux.HandleFunc(APIV1, requestSizeLimitHandler)
}

func (app *KnativePublishApplication) registerPublishV2Handler() {
	knativePublishHandler := handlers.KnativePublishHandler(constants.EventAPIV2, app.knativeLib, app.knativePublisher, app.tracer, app.options)
	requestSizeLimitHandler := handlers.WithRequestSizeLimiting(knativePublishHandler, app.options.MaxRequestSize)
	app.serveMux.HandleFunc(APIV2, requestSizeLimitHandler)
}
