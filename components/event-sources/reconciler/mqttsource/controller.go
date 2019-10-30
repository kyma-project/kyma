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

// Package mqttsource implements a controller for the MQTTSource custom resource.
package mqttsource

import (
	"context"
	"fmt"
	"os"

	"k8s.io/client-go/tools/cache"

	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/resolver"
	servingclient "knative.dev/serving/pkg/client/injection/client"
	knserviceinformersv1 "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	sourcesclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client"
	mqttsourceinformersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/informers/sources/v1alpha1/mqttsource"
)

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "MQTTSources"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "mqtt-source-controller"

	// adapterImageEnvVar is the name of the environment variable containing the
	// container image of the MQTT adapter.
	adapterImageEnvVar = "MQTT_ADAPTER_IMAGE"
)

// NewController returns a new controller that reconciles MQTTSource objects.
func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	mqttSourceInformer := mqttsourceinformersv1alpha1.Get(ctx)
	knServiceInformer := knserviceinformersv1.Get(ctx)

	r := &Reconciler{
		Base:             reconciler.NewBase(ctx, controllerAgentName, cmw),
		adapterImage:     getAdapterImage(),
		mqttsourceLister: mqttSourceInformer.Lister(),
		ksvcLister:       knServiceInformer.Lister(),
		sourcesClient:    sourcesclient.Get(ctx).SourcesV1alpha1(),
		servingClient:    servingclient.Get(ctx).ServingV1(),
	}
	impl := controller.NewImpl(r, r.Logger, reconcilerName)

	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	// set event handlers

	mqttSourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	knServiceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(sourcesv1alpha1.MQTTSourceGVK()),
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
