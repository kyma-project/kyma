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
	"fmt"
	"os"

	"k8s.io/client-go/tools/cache"

	messaginginformersv1alpha1 "knative.dev/eventing/pkg/client/injection/informers/messaging/v1alpha1/channel"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/resolver"
	servingclient "knative.dev/serving/pkg/client/injection/client"
	knserviceinformersv1 "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	sourcesclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client"
	httpsourceinformersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/informers/sources/v1alpha1/httpsource"
)

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "HTTPSources"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "http-source-controller"

	// adapterImageEnvVar is the name of the environment variable containing the
	// container image of the HTTP adapter.
	adapterImageEnvVar = "HTTP_ADAPTER_IMAGE"
)

// NewController returns a new controller that reconciles HTTPSource objects.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	httpSourceInformer := httpsourceinformersv1alpha1.Get(ctx)
	knServiceInformer := knserviceinformersv1.Get(ctx)
	chInformer := messaginginformersv1alpha1.Get(ctx)

	rb := reconciler.NewBase(ctx, controllerAgentName, cmw)
	r := &Reconciler{
		Base:             rb,
		adapterImage:     getAdapterImage(),
		httpsourceLister: httpSourceInformer.Lister(),
		ksvcLister:       knServiceInformer.Lister(),
		chLister:         chInformer.Lister(),
		sourcesClient:    sourcesclient.Get(ctx).SourcesV1alpha1(),
		servingClient:    servingclient.Get(ctx).ServingV1(),
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

	return impl
}

func getAdapterImage() string {
	if adapterImage := os.Getenv(adapterImageEnvVar); adapterImage != "" {
		return adapterImage
	}
	panic(fmt.Errorf("environment variable %s is not set", adapterImageEnvVar))
}
