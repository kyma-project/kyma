package eventactivation

import (
	"context"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"

	"k8s.io/client-go/kubernetes/scheme"

	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	eventbusscheme "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/scheme"
	eventbusclient "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client"
	eventactivationinformersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/informers/applicationconnector/v1alpha1/eventactivation"
)

//var log = logf.Log.WithName("eventactivation-controller")
//
//const (
//	controllerAgentName = "eventactivation-controller"
//)

//// ProvideController instantiates a reconciler which reconciles EventActivations.
//func ProvideController(mgr manager.Manager) error {
//
//	var err error
//
//	// Setup a new controller to Reconcile EventActivation.
//	r := &reconciler{
//		recorder: mgr.GetEventRecorderFor(controllerAgentName),
//		time:     util.NewDefaultCurrentTime(),
//	}
//	c, err := controller.New(controllerAgentName, mgr, controller.Options{
//		Reconciler: r,
//	})
//	if err != nil {
//		log.Error(err, "Unable to create controller")
//		return err
//	}
//
//	// Watch EventActivations.
//	err = c.Watch(&source.Kind{
//		Type: &eventactivationv1alpha1.EventActivation{},
//	}, &handler.EnqueueRequestForObject{})
//	if err != nil {
//		log.Error(err, "Unable to watch EventActivation")
//		return err
//	}
//
//	return nil
//}

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "EventActivations"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "eventactivation-controller"
)

func init() {
	// Add sources types to the default Kubernetes Scheme so Events can be
	// logged for sources types.
	eventbusscheme.AddToScheme(scheme.Scheme)
}

// NewController returns a new controller that reconciles EventActivation objects.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	eventActivationInformer := eventactivationinformersv1alpha1.Get(ctx)

	r := &Reconciler{
		Base:                       reconciler.NewBase(ctx, controllerAgentName, cmw),
		eventActivationLister:      eventActivationInformer.Lister(),
		applicationconnectorClient: eventbusclient.Get(ctx).ApplicationconnectorV1alpha1(),
		eventingClient:             eventbusclient.Get(ctx).EventingV1alpha1(),
		time:                       util.NewDefaultCurrentTime(),
	}
	impl := controller.NewImpl(r, r.Logger, reconcilerName)

	return impl
}
