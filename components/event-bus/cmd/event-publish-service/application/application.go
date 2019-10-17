package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	cloudevents "github.com/cloudevents/sdk-go"

	cloudeventstransport "github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
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
	codec := cloudeventshttp.CodecV03{
		DefaultEncoding: cloudeventshttp.StructuredV03,
	}

	m, err := codec.Encode(ctx, event)
	if err != nil {
		return err
	}

	message, ok := m.(*cloudeventshttp.Message)
	if !ok {
		return fmt.Errorf("expected type http message, but got type: %v", reflect.TypeOf(m))
	}

	fmt.Printf("%v", message)

	etv, err := event.Context.GetExtension("event-type-version")

	var etvstring string

	if rawmessage, ok := etv.(json.RawMessage); ok {

		err := json.Unmarshal(rawmessage, &etvstring)
		if err != nil {
			panic(err)
		}
	}

	if err != nil {
		return err
	}
	ns := knative.GetDefaultChannelNamespace()
	header := map[string][]string(message.Header)

	publishError, namespace, channelname := (*app.knativePublisher).Publish(app.knativeLib, &ns, &header, &message.Body, event.Source(), event.Type(), etvstring)
	fmt.Printf("%+v\n\n%+v\n\n%+v\n\n", publishError, namespace, channelname)

	b, err := json.Marshal(publishError)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s", b)
}
