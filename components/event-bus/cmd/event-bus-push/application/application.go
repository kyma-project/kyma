package application

import (
	"log"
	"net/http"

	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/actors"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/controllers"
	pushOpts "github.com/kyma-project/kyma/components/event-bus/internal/push/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
	"k8s.io/client-go/tools/cache"
)

// PushApplication ...
type PushApplication struct {
	SubscriptionsSupervisor *actors.SubscriptionsSupervisor
	subscriptionsController *controllers.SubscriptionsController
	ServerMux               *http.ServeMux
	tracer                  trace.Tracer
}

// NewPushApplication ...
func NewPushApplication(pushOpts *pushOpts.Options, informer ...cache.SharedIndexInformer) *PushApplication {
	log.Println("Push :: Initializing application")
	tracer := trace.StartNewTracer(&pushOpts.Options)
	subscriptionsSupervisor := actors.StartSubscriptionsSupervisor(pushOpts, &tracer)
	var subscriptionsController *controllers.SubscriptionsController
	if len(informer) > 0 {
		subscriptionsController = controllers.StartSubscriptionsControllerWithInformer(subscriptionsSupervisor, informer[0], pushOpts)
	} else {
		subscriptionsController = controllers.StartSubscriptionsController(subscriptionsSupervisor, pushOpts)
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/v1/status/live", statusLiveHandler(subscriptionsSupervisor))
	serveMux.Handle("/v1/status/ready", statusReadyHandler(subscriptionsSupervisor))

	return &PushApplication{
		SubscriptionsSupervisor: subscriptionsSupervisor,
		subscriptionsController: subscriptionsController,
		ServerMux:               serveMux,
		tracer:                  tracer,
	}
}

// Stop ...
func (a *PushApplication) Stop() {
	a.subscriptionsController.Stop()
	a.SubscriptionsSupervisor.PoisonPill()
	a.tracer.Stop()
}

var statusLive, statusReady common.StatusReady

func statusLiveHandler(subscriptionsSupervisor *actors.SubscriptionsSupervisor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subscriptionsSupervisor != nil && subscriptionsSupervisor.IsRunning() {
			if statusLive.SetReady() {
				log.Printf("statusLiveHandler :: Status: READY")
			}
			w.WriteHeader(http.StatusOK)
		} else {
			statusLive.SetNotReady()
			log.Printf("statusLiveHandler :: Status: NOT_READY")
			w.WriteHeader(http.StatusBadGateway)
		}
	})
}

func statusReadyHandler(subscriptionsSupervisor *actors.SubscriptionsSupervisor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subscriptionsSupervisor != nil && subscriptionsSupervisor.IsNATSConnected() {
			if statusReady.SetReady() {
				log.Printf("statusReadyHandler :: Status: READY")
			}
			w.WriteHeader(http.StatusOK)
		} else {
			statusReady.SetNotReady()
			log.Printf("statusReadyHandler :: Status: NOT_READY")
			w.WriteHeader(http.StatusBadGateway)
			go subscriptionsSupervisor.ReconnectToNATSStreaming()
		}
	})
}
