package knativesubscription

import (
	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("knative-subscription-controller")

const (
	controllerAgentName = "knative-subscription-controller"
)

// ProvideController returns an Knative-subscription controller.
func ProvideController(mgr manager.Manager) (controller.Controller, error) {

	var err error

	// Setup a new controller to Reconcile Knative Subscriptions.
	r := &reconciler{
		recorder: mgr.GetRecorder(controllerAgentName),
		time:     util.NewDefaultCurrentTime(),
	}
	c, err := controller.New(controllerAgentName, mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		log.Error(err, "Unable to create controller")
		return nil, err
	}

	// Watch EventActivations.
	err = c.Watch(&source.Kind{
		Type: &evapisv1alpha1.Subscription{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Unable to watch Knative Subscription")
		return nil, err
	}

	return c, nil
}
