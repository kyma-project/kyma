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

package httpsource

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8stesting "k8s.io/client-go/testing"

	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	rt "knative.dev/pkg/reconciler/testing"
	"knative.dev/pkg/resolver"
	fakeservingclient "knative.dev/serving/pkg/client/injection/client/fake"

	fakesourcesclient "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client/fake"
	. "github.com/kyma-project/kyma/components/event-sources/reconciler/testing"
)

const (
	tNs      = "testns"
	tName    = "test"
	tImg     = "sources.kyma-project.io/http:latest"
	tSinkURI = "http://" + tName + "-kn-channel." + tNs + ".svc.cluster.local"
)

func TestReconcile(t *testing.T) {
	testCases := rt.TableTest{
		/* Error handling */

		{
			Name:    "Source was deleted",
			Key:     tNs + "/" + tName,
			Objects: nil,
			WantErr: false,
		},
		{
			Name:    "Invalid object key",
			Key:     tNs + "/" + tName + "/invalid",
			WantErr: true,
		},

		/* Service synchronization */

		{
			Name: "Initial source creation",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				NewHTTPSource(tNs, tName),
			},
			WantCreates: []runtime.Object{
				NewService(tNs, tName,
					WithServiceController(tName),
					WithServiceContainer(tImg),
				),
				NewChannel(tNs, tName,
					WithChannelController(tName),
				),
			},
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: NewHTTPSource(tNs, tName,
					WithInitConditions,
					WithNoSink,
				),
			}},
		},
		{
			Name: "Everything up-to-date",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				NewHTTPSource(tNs, tName,
					WithSink(tSinkURI),
					WithDeployed,
				),
				NewService(tNs, tName,
					WithServiceContainer(tImg),
					WithServiceReady,
				),
				NewChannel(tNs, tName,
					WithChannelController(tName),
					WithChannelSinkURI(tSinkURI),
				),
			},
			WantCreates:       nil,
			WantUpdates:       nil,
			WantStatusUpdates: nil,
		},
		{
			Name: "Adapter Service spec does not match expectation",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				NewHTTPSource(tNs, tName,
					WithDeployed,
					WithSink(tSinkURI),
				),
				NewService(tNs, tName,
					WithServiceContainer("outdated"),
					WithServiceReady,
				),
				NewChannel(tNs, tName,
					WithChannelController(tName),
					WithChannelSinkURI(tSinkURI),
				),
			},
			WantCreates: nil,
			WantUpdates: []k8stesting.UpdateActionImpl{{
				Object: NewService(tNs, tName,
					WithServiceController(tName),
					WithServiceContainer(tImg),
					WithServiceReady),
			}},
			WantStatusUpdates: nil,
		},

		/* Channel synchronization */

		{
			Name: "Channel missing",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				NewHTTPSource(tNs, tName,
					WithDeployed,
					WithNoSink,
				),
				NewService(tNs, tName,
					WithServiceController(tName),
					WithServiceContainer(tImg),
					WithServiceReady,
				),
			},
			WantCreates: []runtime.Object{
				NewChannel(tNs, tName,
					WithChannelController(tName),
				),
			},
			WantUpdates:       nil,
			WantStatusUpdates: nil,
		},

		/* Status updates */

		{
			Name: "Adapter Service deployment in progress",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				NewHTTPSource(tNs, tName,
					WithNotDeployed,
					WithSink(tSinkURI),
				),
				NewService(tNs, tName,
					WithServiceContainer(tImg),
					WithServiceNotReady,
				),
				NewChannel(tNs, tName,
					WithChannelController(tName),
					WithChannelSinkURI(tSinkURI),
				),
			},
			WantCreates:       nil,
			WantUpdates:       nil,
			WantStatusUpdates: nil,
		},
		{
			Name: "Adapter Service became ready",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				NewHTTPSource(tNs, tName,
					WithNotDeployed,
					WithSink(tSinkURI),
				),
				NewService(tNs, tName,
					WithServiceContainer(tImg),
					WithServiceReady,
				),
				NewChannel(tNs, tName,
					WithChannelController(tName),
					WithChannelSinkURI(tSinkURI),
				),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: NewHTTPSource(tNs, tName,
					WithDeployed,
					WithSink(tSinkURI),
				),
			}},
		},
		{
			Name: "Adapter Service became not ready",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				NewHTTPSource(tNs, tName,
					WithDeployed,
					WithSink(tSinkURI),
				),
				NewService(tNs, tName,
					WithServiceContainer(tImg),
					WithServiceNotReady,
				),
				NewChannel(tNs, tName,
					WithChannelController(tName),
					WithChannelSinkURI(tSinkURI),
				),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: NewHTTPSource(tNs, tName,
					WithNotDeployed,
					WithSink(tSinkURI),
				),
			}},
		},
		{
			Name: "Channel becomes available",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				NewHTTPSource(tNs, tName,
					WithDeployed,
					WithNoSink,
				),
				NewService(tNs, tName,
					WithServiceController(tName),
					WithServiceContainer(tImg),
					WithServiceReady,
				),
				NewChannel(tNs, tName,
					WithChannelController(tName),
					WithChannelSinkURI(tSinkURI),
				),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: NewHTTPSource(tNs, tName,
					WithDeployed,
					WithSink(tSinkURI),
				),
			}},
		},
		{
			Name: "Channel becomes unavailable",
			Key:  tNs + "/" + tName,
			Objects: []runtime.Object{
				NewHTTPSource(tNs, tName,
					WithDeployed,
					WithSink(tSinkURI),
				),
				NewService(tNs, tName,
					WithServiceController(tName),
					WithServiceContainer(tImg),
					WithServiceReady,
				),
				NewChannel(tNs, tName,
					WithChannelController(tName),
				),
			},
			WantCreates: nil,
			WantUpdates: nil,
			WantStatusUpdates: []k8stesting.UpdateActionImpl{{
				Object: NewHTTPSource(tNs, tName,
					WithDeployed,
					WithNoSink,
				),
			}},
		},
	}

	var ctor Ctor = func(ctx context.Context, ls *Listers) controller.Reconciler {
		rb := reconciler.NewBase(ctx, controllerAgentName, configmap.NewStaticWatcher())
		r := &Reconciler{
			Base:             rb,
			httpsourceLister: ls.GetHTTPSourceLister(),
			ksvcLister:       ls.GetServiceLister(),
			chLister:         ls.GetChannelLister(),
			sourcesClient:    fakesourcesclient.Get(ctx).SourcesV1alpha1(),
			servingClient:    fakeservingclient.Get(ctx).ServingV1alpha1(),
			messagingClient:  rb.EventingClientSet.MessagingV1alpha1(),
			adapterImage:     tImg,
			sinkResolver:     resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
		}

		return r
	}

	testCases.Test(t, MakeFactory(ctor))
}
