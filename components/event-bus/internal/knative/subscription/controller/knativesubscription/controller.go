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

// ProvideController instantiates a reconciler which reconciles Knative Subscriptions.
func ProvideController(mgr manager.Manager) error {

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
		log.Error(err, "unable to create Knative subscription controller")
		return err
	}

	// Watch Knative Subscriptions
	err = c.Watch(&source.Kind{
		Type: &evapisv1alpha1.Subscription{},
	}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "unable to watch Knative Subscription")
		return err
	}

	return nil
}
