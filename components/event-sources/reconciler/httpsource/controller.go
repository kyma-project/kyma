// Package httpsource implements a controller for the HTTPSource custom resource.
package httpsource

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	messaginginformersv1alpha1 "knative.dev/eventing/pkg/client/injection/informers/messaging/v1alpha1/channel"
	"knative.dev/eventing/pkg/reconciler"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	serviceinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/service"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/resolver"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	sourcesscheme "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/scheme"
	sourcesclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client"
	httpsourceinformersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/informers/sources/v1alpha1/httpsource"
	istioclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/istio/client"
	peerauthenticationinformersv1beta1 "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/istio/informers/security/v1beta1/peerauthentication"
)

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "HTTPSources"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "http-source-controller"
)

func init() {
	// Add sources types to the default Kubernetes Scheme so Events can be
	// logged for sources types.
	utilruntime.Must(sourcesscheme.AddToScheme(scheme.Scheme))
}

// NewController returns a new controller that reconciles HTTPSource objects.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	adapterEnvCfg := &httpAdapterEnvConfig{}
	envconfig.MustProcess("http_adapter", adapterEnvCfg)

	httpSourceInformer := httpsourceinformersv1alpha1.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)
	chInformer := messaginginformersv1alpha1.Get(ctx)
	serviceInformer := serviceinformer.Get(ctx)
	peerAuthenticationInformer := peerauthenticationinformersv1beta1.Get(ctx)

	rb := reconciler.NewBase(ctx, controllerAgentName, cmw)
	r := &Reconciler{
		Base:                     rb,
		adapterEnvCfg:            adapterEnvCfg,
		httpsourceLister:         httpSourceInformer.Lister(),
		deploymentLister:         deploymentInformer.Lister(),
		chLister:                 chInformer.Lister(),
		peerAuthenticationLister: peerAuthenticationInformer.Lister(),
		serviceLister:            serviceInformer.Lister(),
		sourcesClient:            sourcesclient.Get(ctx).SourcesV1alpha1(),
		messagingClient:          rb.EventingClientSet.MessagingV1alpha1(),
		securityClient:           istioclient.Get(ctx).SecurityV1beta1(),
	}
	impl := controller.NewImpl(r, r.Logger, reconcilerName)

	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	// set event handlers

	httpSourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	eventHandler := cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(sourcesv1alpha1.HTTPSourceGVK()),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	}

	// for the deployment of the adapter
	deploymentInformer.Informer().AddEventHandler(eventHandler)
	serviceInformer.Informer().AddEventHandler(eventHandler)

	// the eventing channel
	chInformer.Informer().AddEventHandler(eventHandler)

	// istio
	peerAuthenticationInformer.Informer().AddEventHandler(eventHandler)

	// watch for changes to metrics/logging configs
	cmw.Watch(metrics.ConfigMapName(), r.updateAdapterMetricsConfig)
	cmw.Watch(logging.ConfigMapName(), r.updateAdapterLoggingConfig)

	return impl
}
