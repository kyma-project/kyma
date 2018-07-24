package application

import (
	"log"
	"net/http"

	"github.com/kyma-project/kyma/components/event-bus/internal/sv"
	"github.com/kyma-project/kyma/components/event-bus/internal/sv/opts"
	"k8s.io/client-go/tools/cache"
)

// SubscriptionValidatorApplication ...
type SubscriptionValidatorApplication struct {
	eaController  *ea.EventActivationsController
	subController *ea.SubscriptionsController
	ServerMux     *http.ServeMux
}

// NewSubscriptionValidatorApplication ...
// func NewSubscriptionValidatorApplication(eaOpts *ea.Options, informer ...cache.SharedIndexInformer) *SubscriptionValidatorApplication {
func NewSubscriptionValidatorApplication(svOpts *opts.Options, informer ...cache.SharedIndexInformer) *SubscriptionValidatorApplication {
	log.Println("Subscription Validator :: Initializing application")
	var eaController *ea.EventActivationsController
	if len(informer) > 0 {
		eaController = ea.StartEventActivationsControllerWithInformer(informer[0])
	} else {
		eaController = ea.StartEventActivationsController()
	}

	subController := ea.StartSubscriptionsController()

	serveMux := http.NewServeMux()
	serveMux.Handle("/v1/status/live", statusLiveHandler(eaController))
	serveMux.Handle("/v1/status/ready", statusLiveHandler(eaController))

	return &SubscriptionValidatorApplication{
		eaController:  eaController,
		subController: subController,
		ServerMux:     serveMux,
	}
}

// Stop ...
func (a *SubscriptionValidatorApplication) Stop() {
	a.eaController.Stop()
	a.subController.Stop()
}

func statusLiveHandler(controller *ea.EventActivationsController) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if controller != nil && controller.IsRunning() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadGateway)
		}
	})
}

// TODO more complex readness tests
func statusReadyHandler(controller *ea.EventActivationsController) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if controller != nil && controller.IsRunning() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadGateway)
		}
	})
}
