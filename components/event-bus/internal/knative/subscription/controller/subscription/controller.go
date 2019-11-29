package subscription

import (
	"context"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	eventbusscheme "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/scheme"
	eventbusclient "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client"
	eventactivationinformersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/informers/applicationconnector/v1alpha1/eventactivation"
	subscriptioninformersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/informers/eventing/v1alpha1/subscription"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
)

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "Subscriptions"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "subscription-controller"
)

func init() {
	// Add custom types to the default Kubernetes Scheme so Events can be
	// logged for those types.
	runtime.Must(eventbusscheme.AddToScheme(scheme.Scheme))
}

// NewController returns a new controller that reconciles Subscriptions objects.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	subscriptionInformer := subscriptioninformersv1alpha1.Get(ctx)
	eventActivationInformer := eventactivationinformersv1alpha1.Get(ctx)
	knativeLib, err := util.NewKnativeLib()
	if err != nil {
		panic("Failed to initialize knative lib")
	}
	SubscriptionsStatsReporter, err := NewStatsReporter()
	if err != nil {
		panic("Failed to Kyma Subscription Controller stats reporter")
	}

	r := &Reconciler{
		Base:                       reconciler.NewBase(ctx, controllerAgentName, cmw),
		subscriptionLister:         subscriptionInformer.Lister(),
		eventActivationLister:      eventActivationInformer.Lister(),
		kymaEventingClient:         eventbusclient.Get(ctx).EventingV1alpha1(),
		knativeLib:                 knativeLib,
		opts:                       opts.DefaultOptions(),
		time:                       util.NewDefaultCurrentTime(),
		SubscriptionsStatsReporter: SubscriptionsStatsReporter,
	}
	impl := controller.NewImpl(r, r.Logger, reconcilerName)

	subscriptionInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	registerMetrics()

	return impl
}
