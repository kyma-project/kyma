package subscription

import (
	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

var log = logf.Log.WithName("subscription-controller")

const (
	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "subscription-controller"
)

// ProvideController instantiates a reconciler which reconciles Kyma Subscriptions.
func ProvideController(mgr manager.Manager, opts *opts.Options) error {

	// init the knative lib
	knativeLib, err := knative.NewKnativeLib()
	if err != nil {
		log.Error(err, "Failed to get Knative library")
		return err
	}

	// Setup a new controller to Reconcile Kyma Subscription.
	r := &reconciler{
		recorder:   mgr.GetRecorder(controllerAgentName),
		opts:       opts,
		knativeLib: knativeLib,
		time:       util.NewDefaultCurrentTime(),
	}
	c, err := controller.New(controllerAgentName, mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		log.Error(err, "Unable to create controller")
		return err
	}

	// Watch Subscriptions.
	err = c.Watch(&source.Kind{
		Type: &eventingv1alpha1.Subscription{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Unable to watch Subscription")
		return err
	}

	return nil
}
