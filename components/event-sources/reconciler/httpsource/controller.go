/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/resolver"
	servingclient "knative.dev/serving/pkg/client/injection/client"
	knserviceinformersv1alpha1 "knative.dev/serving/pkg/client/injection/informers/serving/v1alpha1/service"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	sourcesscheme "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/scheme"
	sourcesclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client"
	httpsourceinformersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/informers/sources/v1alpha1/httpsource"
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
	knServiceInformer := knserviceinformersv1alpha1.Get(ctx)
	chInformer := messaginginformersv1alpha1.Get(ctx)

	rb := reconciler.NewBase(ctx, controllerAgentName, cmw)
	r := &Reconciler{
		Base:             rb,
		adapterEnvCfg:    adapterEnvCfg,
		httpsourceLister: httpSourceInformer.Lister(),
		ksvcLister:       knServiceInformer.Lister(),
		chLister:         chInformer.Lister(),
		sourcesClient:    sourcesclient.Get(ctx).SourcesV1alpha1(),
		servingClient:    servingclient.Get(ctx).ServingV1alpha1(),
		messagingClient:  rb.EventingClientSet.MessagingV1alpha1(),
	}
	impl := controller.NewImpl(r, r.Logger, reconcilerName)

	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	// set event handlers

	httpSourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	knServiceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(sourcesv1alpha1.HTTPSourceGVK()),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	chInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(sourcesv1alpha1.HTTPSourceGVK()),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	// watch for changes to metrics/logging configs

	cmw.Watch(metrics.ConfigMapName(), r.updateAdapterMetricsConfig)
	cmw.Watch(logging.ConfigMapName(), r.updateAdapterLoggingConfig)

	return impl
}
