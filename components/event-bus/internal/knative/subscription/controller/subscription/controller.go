package subscription

import (
	"context"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"

	"k8s.io/client-go/kubernetes/scheme"

	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"

	eventbusscheme "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/scheme"
	eventbusclient "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client"
	eventactivationinformersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/informers/applicationconnector/v1alpha1/eventactivation"
	subscriptioninformersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/informers/eventing/v1alpha1/subscription"
)

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "Subscriptions"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "subscription-controller"
)

func init() {
	// Add sources types to the default Kubernetes Scheme so Events can be
	// logged for sources types.
	eventbusscheme.AddToScheme(scheme.Scheme)
}

// NewController returns a new controller that reconciles EventActivation objects.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	subscriptionInformer := subscriptioninformersv1alpha1.Get(ctx)
	eventActivationInformer := eventactivationinformersv1alpha1.Get(ctx)

	r := &Reconciler{
		Base:                  reconciler.NewBase(ctx, controllerAgentName, cmw),
		subscriptionLister:    subscriptionInformer.Lister(),
		eventActivationLister: eventActivationInformer.Lister(),
		kymaEventingClient:    eventbusclient.Get(ctx).EventingV1alpha1(),
		opts:                  opts.DefaultOptions(),
		time:                  util.NewDefaultCurrentTime(),
	}
	impl := controller.NewImpl(r, r.Logger, reconcilerName)

	return impl
}
