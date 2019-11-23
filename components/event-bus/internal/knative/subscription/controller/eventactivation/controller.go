package eventactivation

import (
	"context"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	eventbusscheme "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/scheme"
	eventbusclient "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client"
	eventactivationinformersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/informers/applicationconnector/v1alpha1/eventactivation"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "EventActivations"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "eventactivation-controller"
)

func init() {
	// Add custom types to the default Kubernetes Scheme so Events can be
	// logged for those types.
	runtime.Must(eventbusscheme.AddToScheme(scheme.Scheme))
}

// NewController returns a new controller that reconciles EventActivation objects.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	eventActivationInformer := eventactivationinformersv1alpha1.Get(ctx)

	r := &Reconciler{
		Base:                       reconciler.NewBase(ctx, controllerAgentName, cmw),
		eventActivationLister:      eventActivationInformer.Lister(),
		applicationconnectorClient: eventbusclient.Get(ctx).ApplicationconnectorV1alpha1(),
		kymaEventingClient:         eventbusclient.Get(ctx).EventingV1alpha1(),
		time:                       util.NewDefaultCurrentTime(),
	}
	impl := controller.NewImpl(r, r.Logger, reconcilerName)

	eventActivationInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	return impl
}
