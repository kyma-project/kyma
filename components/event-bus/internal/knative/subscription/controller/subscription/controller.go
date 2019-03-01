package subscription

import (
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("subscription-controller")

const (
	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "subscription-controller"
)

// ProvideController returns a Controller that represents the subscription controller. It
// reconciles only the Kyma Subscription.
func ProvideController(mgr manager.Manager, opts *opts.Options) (controller.Controller, error) {

	// Setup a new controller to Reconcile Kyma Subscription.
	r := &reconciler{
		recorder: mgr.GetRecorder(controllerAgentName),
		opts: opts,
	}
	c, err := controller.New(controllerAgentName, mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		log.Error(err, "Unable to create controller")
		return nil, err
	}

	// Watch Subscriptions.
	err = c.Watch(&source.Kind{
		Type: &eventingv1alpha1.Subscription{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Unable to watch Subscription")
		return nil, err
	}

	return c, nil
}